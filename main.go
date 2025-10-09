package main

import (
	"github.com/joho/godotenv"
	"github.com/playwright-community/playwright-go"
	"log"
	"store/service"
)

func init() {
	// 加载 .env 文件
	err := godotenv.Load()
	if err != nil {
		log.Fatal("No .env file found, relying on system environment variables")
	}
}

func main() {
	// 启动 Playwright
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
	service.Hash(browser)
	service.SetupGracefulShutdown()
}
