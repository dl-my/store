package utils

import (
	"bytes"
	"fmt"
	"github.com/playwright-community/playwright-go"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"store/common"
	"testing"
)

func Test(t *testing.T) {
	url := common.Url3
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	defer pw.Stop()

	// 启动 Chromium 浏览器（headless）
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	defer browser.Close()
	// 获取网站截图
	pngBytes, err := playwrightWithNet(browser, url)
	if err != nil {
		fmt.Printf("screenshot failed: %v\n", err)
		return
	}
	// 初始化基线图片
	baselinePath := filepath.Join(common.PngDir[url], "baseline.png")
	if _, err := os.Stat(baselinePath); os.IsNotExist(err) {
		if err := savePNG(pngBytes, baselinePath); err != nil {
			fmt.Printf("save baseline failed: %v\n", err)
			return
		}
		log.Printf("基线不存在，已初始化基线: %s", baselinePath)
		return
	}

	// 读取基线图
	baseBytes, err := os.ReadFile(baselinePath)
	if err != nil {
		fmt.Printf("read baseline failed: %v\n", err)
		return
	}

	// 解析基线图
	baseImg, err := decodePNG(baseBytes)
	if err != nil {
		fmt.Printf("decode baseline failed: %v\n", err)
		return
	}
	// 解析新截图
	curImg, err := decodePNG(pngBytes)
	if err != nil {
		fmt.Printf("decode current failed: %v\n", err)
		return
	}

	// 计算图片差异
	rects := diffBlocks(baseImg, curImg, 20, 8.0)

	if len(rects) == 0 {
		log.Printf("未检测到显著变化 (阈值=%.2f, 块大小=%d)", 8.0, 20)
		return
	}

	// 绘画红线
	annotated := image.NewRGBA(curImg.Bounds())
	draw.Draw(annotated, annotated.Bounds(), curImg, image.Point{}, draw.Src)
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	for _, r := range rects {
		// expand a bit for visibility
		exp := r.Inset(-6)
		drawRect(annotated, exp, red, 3)
	}

	// 保存绘画后的图片
	var outBuf bytes.Buffer
	if err := png.Encode(&outBuf, annotated); err != nil {
		fmt.Printf("encode annotated failed: %v\n", err)
		return
	}
	diffPath := filepath.Join(common.PngDir[url], "diff.png")
	if err := savePNG(outBuf.Bytes(), diffPath); err != nil {
		fmt.Printf("save diff failed: %v\n", err)
		return
	}
	log.Printf("检测到变化: %d 个区域. 差异图已保存: %s", len(rects), diffPath)
	// 更新基线
	prevPath := filepath.Join(common.PngDir[url], "prev.png")
	if err = savePNG(baseBytes, prevPath); err != nil {
		fmt.Printf("save diff failed: %v\n", err)
		return
	}
	log.Printf("已备份旧基线为: %s", prevPath)
	if err := savePNG(pngBytes, baselinePath); err != nil {
		fmt.Printf("update baseline failed: %v\n", err)
		return
	}
	log.Printf("基线已更新")
}

func TestScreenshot(t *testing.T) {
	url := common.Url3
	// 启动 Playwright
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start Playwright: %v", err)
	}
	defer pw.Stop()

	// 启动 Chromium 浏览器（headless）
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	defer browser.Close()
	// 新建页面
	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	// 设置视口大小
	if err := page.SetViewportSize(1280, 7133); err != nil {
		log.Fatalf("could not set viewport: %v", err)
	}

	// 打开页面并等待网络空闲
	_, err = page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(30000),
	})
	if err != nil {
		log.Fatalf("could not navigate: %v", err)
	}

	// 全页截图
	rect := playwright.Rect{
		X:      0,
		Y:      0,
		Width:  1280,
		Height: 7133,
	}
	if _, err := page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String("screenshot.png"),
		Clip: &rect,
		//FullPage: playwright.Bool(true),
	}); err != nil {
		log.Fatalf("could not take screenshot: %v", err)
	}

	log.Println("截图已保存为 screenshot.png")
}

func TestPngDiffBlocks(t *testing.T) {
	url := common.Url3
	baselinePath := filepath.Join(common.PngDir[url], "baseline.png")
	// 读取基线图
	baseBytes, err := os.ReadFile(baselinePath)
	if err != nil {
		fmt.Printf("read baseline failed: %v\n", err)
		return
	}

	// 解析基线图
	baseImg, err := decodePNG(baseBytes)
	if err != nil {
		fmt.Printf("decode prev failed: %v\n", err)
		return
	}
	prevPath := filepath.Join(common.PngDir[url], "prev.png")
	// 读取基线图
	prevBytes, err := os.ReadFile(prevPath)
	if err != nil {
		fmt.Printf("read prev failed: %v\n", err)
		return
	}
	// 解析新截图
	curImg, err := decodePNG(prevBytes)
	if err != nil {
		fmt.Printf("decode current failed: %v\n", err)
		return
	}

	// 计算图片差异
	rects := diffBlocks(baseImg, curImg, 20, 8.0)

	if len(rects) == 0 {
		log.Printf("未检测到显著变化 (阈值=%.2f, 块大小=%d)", 8.0, 20)
		return
	}

	// 绘画红线
	annotated := image.NewRGBA(curImg.Bounds())
	draw.Draw(annotated, annotated.Bounds(), curImg, image.Point{}, draw.Src)
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	for _, r := range rects {
		// expand a bit for visibility
		exp := r.Inset(-6)
		drawRect(annotated, exp, red, 3)
	}

	// 保存绘画后的图片
	var outBuf bytes.Buffer
	if err := png.Encode(&outBuf, annotated); err != nil {
		fmt.Printf("encode annotated failed: %v\n", err)
		return
	}
	diffPath := filepath.Join(common.PngDir[url], "diff.png")
	if err := savePNG(outBuf.Bytes(), diffPath); err != nil {
		fmt.Printf("save diff failed: %v\n", err)
		return
	}
	log.Printf("检测到变化: %d 个区域. 差异图已保存: %s", len(rects), diffPath)
}
