package utils

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// TelegramBot 结构体
type TelegramBot struct {
	Token   string
	ChatID  string
	BaseURL string
}

func NewTelegramBot(token, chatID string) *TelegramBot {
	return &TelegramBot{
		Token:   token,
		ChatID:  chatID,
		BaseURL: fmt.Sprintf("https://api.telegram.org/bot%s", token),
	}
}

func (bot *TelegramBot) SendMessage(message string) error {
	apiURL := fmt.Sprintf("%s/sendMessage", bot.BaseURL)

	data := url.Values{}
	data.Set("chat_id", bot.ChatID)
	data.Set("text", message)
	data.Set("parse_mode", "HTML")

	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		return fmt.Errorf("发送消息失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Telegram API返回错误状态: %d, 响应: %s", resp.StatusCode, string(body))
	}

	log.Println("消息已成功发送到Telegram")
	return nil
}

func (bot *TelegramBot) SendPhoto(filePath, caption string) error {
	apiURL := fmt.Sprintf("%s/sendPhoto", bot.BaseURL)

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("无法打开图片: %v", err)
	}
	defer file.Close()

	// 构造 multipart/form-data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// chat_id
	if err := writer.WriteField("chat_id", bot.ChatID); err != nil {
		return err
	}
	// 可选的文字说明
	if caption != "" {
		if err := writer.WriteField("caption", caption); err != nil {
			return err
		}
	}

	// 写入文件
	part, err := writer.CreateFormFile("photo", filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}

	writer.Close()

	// 发送请求
	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送图片失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Telegram API返回错误: %d, 响应: %s", resp.StatusCode, respBody)
	}

	log.Printf("图片已成功发送到Telegram: %s", filePath)
	return nil
}

func (bot *TelegramBot) SendDocument(filePath, caption string) error {
	apiURL := fmt.Sprintf("%s/sendDocument", bot.BaseURL)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("无法打开文件: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("chat_id", bot.ChatID); err != nil {
		return err
	}
	if caption != "" {
		if err := writer.WriteField("caption", caption); err != nil {
			return err
		}
	}

	part, err := writer.CreateFormFile("document", filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}

	writer.Close()

	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送文件失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Telegram API返回错误: %d, 响应: %s", resp.StatusCode, respBody)
	}

	log.Printf("文件已成功发送到Telegram: %s", filePath)
	return nil
}
