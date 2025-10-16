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
)

func SaveAndDiff(bot *TelegramBot, browser playwright.Browser, url string) error {
	pngBytes, err := playwrightWithNet(browser, url)
	if err != nil {
		fmt.Printf("screenshot failed: %v\n", err)
		return err
	}
	// 初始化基线图片
	baselinePath := filepath.Join(common.PngDir[url], "baseline.png")
	if _, err = os.Stat(baselinePath); os.IsNotExist(err) {
		if err = savePNG(pngBytes, baselinePath); err != nil {
			fmt.Printf("save baseline failed: %v\n", err)
			return err
		}
		log.Printf("基线不存在，已初始化基线: %s", baselinePath)
		return nil
	}

	// 读取基线图
	baseBytes, err := os.ReadFile(baselinePath)
	if err != nil {
		fmt.Printf("read baseline failed: %v\n", err)
		return err
	}

	// 解析基线图
	baseImg, err := decodePNG(baseBytes)
	if err != nil {
		fmt.Printf("decode baseline failed: %v\n", err)
		return err
	}
	// 解析新截图
	curImg, err := decodePNG(pngBytes)
	if err != nil {
		fmt.Printf("decode current failed: %v\n", err)
		return err
	}

	// 计算图片差异
	rects := diffBlocks(baseImg, curImg, 20, 8.0)
	if len(rects) == 0 {
		log.Printf("未检测到显著变化 (阈值=%.2f, 块大小=%d)", 8.0, 20)
		return nil
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
	if err = png.Encode(&outBuf, annotated); err != nil {
		fmt.Printf("encode annotated failed: %v\n", err)
		return err
	}
	diffPath := filepath.Join(common.PngDir[url], "diff.png")
	if err = savePNG(outBuf.Bytes(), diffPath); err != nil {
		fmt.Printf("save diff failed: %v\n", err)
		return err
	}
	log.Printf("检测到变化: %d 个区域. 差异图已保存: %s", len(rects), diffPath)
	// 更新基线
	prevPath := filepath.Join(common.PngDir[url], "prev.png")
	if err = savePNG(baseBytes, prevPath); err != nil {
		fmt.Printf("save diff failed: %v\n", err)
		return err
	}
	log.Printf("已备份旧基线为: %s", prevPath)
	if err = savePNG(pngBytes, baselinePath); err != nil {
		fmt.Printf("update baseline failed: %v\n", err)
		return err
	}
	log.Printf("基线已更新")
	// tg消息推送
	if err = bot.SendDocument(diffPath, fmt.Sprintf("检测到变化: %d 个区域", len(rects))); err != nil {
		log.Printf("发送图片到TG失败: %v", err)
		return err
	}
	return nil
}

func playwrightWithNet(browser playwright.Browser, urlStr string) ([]byte, error) {
	// 新建页面
	page, err := browser.NewPage()
	if err != nil {
		return nil, fmt.Errorf("could not create page: %w", err)
	}
	defer page.Close()

	// 设置视口大小
	if err := page.SetViewportSize(common.PngView[urlStr].Width, common.PngView[urlStr].Height); err != nil {
		return nil, fmt.Errorf("could not set viewport: %w", err)
	}

	// 打开页面并等待网络空闲
	_, err = page.Goto(urlStr, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(30000),
	})
	if err != nil {
		return nil, fmt.Errorf("could not navigate: %w", err)
	}

	// 获取截图的字节数组
	buf, err := page.Screenshot(playwright.PageScreenshotOptions{
		FullPage: playwright.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("could not take screenshot: %w", err)
	}

	log.Println("截图已获取")
	return buf, nil
}

func savePNG(pngBytes []byte, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, pngBytes, 0o644)
}

func decodePNG(pngBytes []byte) (*image.RGBA, error) {
	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, err
	}
	// ensure RGBA
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, img.Bounds(), img, image.Point{}, draw.Src)
	return rgba, nil
}

// diffBlocks does block-wise difference. Returns list of rectangles with changes.
func diffBlocks(a, b *image.RGBA, block int, threshold float64) []image.Rectangle {
	// Work on common area
	w := min(a.Bounds().Dx(), b.Bounds().Dx())
	h := min(a.Bounds().Dy(), b.Bounds().Dy())

	var rects []image.Rectangle
	for by := 0; by < h; by += block {
		for bx := 0; bx < w; bx += block {
			x2 := min(bx+block, w)
			y2 := min(by+block, h)
			changed := blockChanged(a, b, bx, by, x2, y2, threshold)
			if changed {
				rects = append(rects, image.Rect(bx, by, x2, y2))
			}
		}
	}
	// Merge adjacent rects
	return mergeRects(rects)
}

func blockChanged(a, b *image.RGBA, x1, y1, x2, y2 int, threshold float64) bool {
	var sum float64
	var n int
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			r1, g1, bl1, _ := a.At(x, y).RGBA()
			r2, g2, bl2, _ := b.At(x, y).RGBA()
			// convert 16-bit to 8-bit by shifting
			dr := float64(int(r1>>8) - int(r2>>8))
			dg := float64(int(g1>>8) - int(g2>>8))
			db := float64(int(bl1>>8) - int(bl2>>8))
			// L1 distance
			sum += abs(dr) + abs(dg) + abs(db)
			n++
		}
	}
	avg := sum / float64(n) // 0..765
	return avg >= threshold
}

func mergeRects(rs []image.Rectangle) []image.Rectangle {
	if len(rs) == 0 {
		return rs
	}
	// Simple greedy merge: if two rects overlap or touch, merge them
	merged := []image.Rectangle{}
	for _, r := range rs {
		merged = appendRectMerged(merged, r)
	}
	return merged
}

func appendRectMerged(list []image.Rectangle, r image.Rectangle) []image.Rectangle {
	for i := 0; i < len(list); i++ {
		if rectsTouchOrOverlap(list[i], r) {
			list[i] = list[i].Union(r)
			return list
		}
	}
	return append(list, r)
}

func rectsTouchOrOverlap(a, b image.Rectangle) bool {
	// expand a by 1px to treat "touching" as overlapping
	expanded := image.Rect(a.Min.X-1, a.Min.Y-1, a.Max.X+1, a.Max.Y+1)
	return expanded.Overlaps(b)
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// drawRect draws rectangle border on RGBA image
func drawRect(img *image.RGBA, r image.Rectangle, c color.RGBA, thickness int) {
	// Clamp to image bounds
	r = r.Intersect(img.Bounds())
	if r.Empty() {
		return
	}
	minX, minY := r.Min.X, r.Min.Y
	maxX, maxY := r.Max.X-1, r.Max.Y-1
	for t := 0; t < thickness; t++ {
		// top & bottom
		for x := minX; x <= maxX; x++ {
			if minY+t >= 0 && minY+t < img.Bounds().Max.Y {
				img.Set(x, minY+t, c)
			}
			if maxY-t >= 0 && maxY-t < img.Bounds().Max.Y {
				img.Set(x, maxY-t, c)
			}
		}
		// left & right
		for y := minY; y <= maxY; y++ {
			if minX+t >= 0 && minX+t < img.Bounds().Max.X {
				img.Set(minX+t, y, c)
			}
			if maxX-t >= 0 && maxX-t < img.Bounds().Max.X {
				img.Set(maxX-t, y, c)
			}
		}
	}
}
