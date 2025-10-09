package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/playwright-community/playwright-go"
	"strings"
)

func dynamicHash(browser playwright.Browser, url string) (string, string, error) {
	context, _ := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36"),
	})
	defer context.Close()

	// 新建页面
	page, err := context.NewPage()
	if err != nil {
		return "", "", fmt.Errorf("%s error creating new page: %w", url, err)
	}
	defer page.Close()

	// 过滤图片和字体资源
	err = page.Route("**/*", func(route playwright.Route) {
		req := route.Request()
		rt := req.ResourceType()
		if rt == "image" || rt == "media" || rt == "font" {
			_ = route.Abort()
			return
		}
		_ = route.Continue()
	})
	if err != nil {
		return "", "", fmt.Errorf("could not set route: %w", err)
	}
	// 打开页面并等待网络空闲
	_, err = page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(30000),
	})
	if err != nil {
		return "", "", fmt.Errorf("could not goto: %w", err)
	}
	// 等待 #app 渲染
	if _, err = page.WaitForSelector("#app", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(10000),
	}); err != nil {
		return "", "", fmt.Errorf("could not wait for selector: %w", err)
	}

	html, err := page.InnerHTML("#app")
	if err != nil {
		return "", "", fmt.Errorf("%s could not get html: %w", url, err)
	}
	// 用 goquery 解析 HTML，提取纯文本
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", "", fmt.Errorf("%s could not parse html: %w", url, err)
	}
	text := doc.Text()

	// 去掉多余空格和换行
	normalized := strings.Join(strings.Fields(text), " ")
	fmt.Println(normalized)

	// 对纯文本做哈希
	sha256Hash := sha256.Sum256([]byte(normalized))
	fmt.Println("Text SHA256:", hex.EncodeToString(sha256Hash[:]))
	return normalized, hex.EncodeToString(sha256Hash[:]), nil
}
