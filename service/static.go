package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

func staticHash(url string) (string, string, error) {
	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", fmt.Errorf("%s 请求创建失败:%w", url, err)
	}

	// 可选：添加请求头，伪装成浏览器
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("%s 请求发送失败:%w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("%s 响应错误: %d", url, resp.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("%s 解析 HTML 失败:%w", url, err)
	}

	doc.Find("script, style").Remove()
	bodyText := doc.Find("body").Text()
	bodyText = strings.Join(strings.Fields(bodyText), " ")
	fmt.Println(bodyText)

	// ---- SHA256 ----
	sha256Hash := sha256.Sum256([]byte(bodyText))
	fmt.Println("SHA256:", hex.EncodeToString(sha256Hash[:]))
	return bodyText, hex.EncodeToString(sha256Hash[:]), nil
}
