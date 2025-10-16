package service

import (
	"encoding/json"
	"fmt"
	"github.com/playwright-community/playwright-go"
	"log"
	"os"
	"os/signal"
	"store/common"
	"store/utils"
	"sync"
	"syscall"
	"time"
)

var (
	mu    sync.Mutex
	Store HashStore
)

// HashStore 用来存储 URL 和 hash
type HashStore map[string]string

func Hash(browser playwright.Browser) {
	// 读取tg频道配置
	token := os.Getenv("TELEGRAM_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHATID")

	if token == "" || chatID == "" {
		log.Fatal("TELEGRAM_TOKEN or TELEGRAM_CHATID is not set")
	}

	fmt.Println("Token:", token)
	fmt.Println("ChatId:", chatID)

	// 加载 hash 文件
	store, err := loadHashStore()
	if err != nil {
		log.Fatalf("加载 hash 文件失败: %v", err)
	}
	Store = store
	bot := utils.NewTelegramBot(token, chatID)
	//err := bot.SendDocument(filepath.Join(pngDir[url3], "baseline.png"), "基线图片")
	//if err != nil {
	//	log.Fatal(err)
	//}
	urls := []string{common.Url1, common.Url2, common.Url3, common.Url4}
	for _, u := range urls {
		go monitor(bot, browser, u, 20*time.Second)
	}
}

func monitor(bot *utils.TelegramBot, browser playwright.Browser, url string, interval time.Duration) {
	lastHash := Store[url]

	for {

		var (
			text, hash string
			err        error
		)

		switch url {
		case common.Url1, common.Url2, common.Url3:
			text, hash, err = staticHash(url)
		case common.Url4:
			text, hash, err = dynamicHash(browser, url)
		default:
			log.Printf("未定义的监控 URL: %s", url)
			return
		}

		// 第一次启动时，写入一次 baseline
		//if lastHash == "" {
		//	if err := utils.AppendUpdateLog(url, text); err != nil {
		//		log.Printf("写入初始日志失败: %v", err)
		//	} else {
		//		log.Printf("初始内容已写入 update.txt")
		//	}
		//	lastHash = hash
		//	time.Sleep(interval)
		//	continue
		//}

		if err != nil {
			log.Println(err)
		} else {
			if lastHash != "" && lastHash != hash {
				if err := utils.AppendUpdateLog(url, text); err != nil {
					log.Printf("写入日志失败: %v", err)
				} else {
					log.Printf("变更内容已写入 update.txt")
				}
				// 根据 URL 选择更新方法
				switch url {
				case common.Url3:
					dynamicUpdate(bot, browser, url)
				default:
					staticUpdate(bot, url)
				}
			}
			// 更新内存 store，不写盘
			mu.Lock()
			Store[url] = hash
			mu.Unlock()
			lastHash = hash
		}
		time.Sleep(interval)
	}
}

func staticUpdate(bot *utils.TelegramBot, url string) {
	msg := fmt.Sprintf("%s 网站更新", url)
	err := bot.SendMessage(msg)
	if err != nil {
		log.Println(err)
	}
}

func dynamicUpdate(bot *utils.TelegramBot, browser playwright.Browser, url string) {
	msg := fmt.Sprintf("%s 网站更新", url)
	err := bot.SendMessage(msg)
	if err != nil {
		log.Println(err)
	}
	const maxRetries = 3
	for i := 1; i <= maxRetries; i++ {
		err = utils.SaveAndDiff(bot, browser, url)
		if err == nil {
			// 成功就跳出
			break
		}
		log.Printf("SaveAndDiff 失败 (第 %d 次): %v", i, err)
		if i < maxRetries {
			time.Sleep(time.Second * 1) // 等待 2 秒再试，可调
		}
	}

	if err != nil {
		log.Printf("SaveAndDiff 最终失败: %v", err)
	}
	//err = utils.SaveAndDiff(bot, browser, url)
}

// LoadHashStore 读取持久化的 hash
func loadHashStore() (HashStore, error) {
	store := make(HashStore)

	f, err := os.Open(common.HashFile)
	if os.IsNotExist(err) {
		return store, nil // 文件不存在返回空
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&store); err != nil {
		return nil, err
	}
	return store, nil
}

// SaveHashStore 更新持久化的 hash
func saveHashStore() error {
	mu.Lock()
	defer mu.Unlock()

	if Store == nil {
		// 不保存，直接返回 nil
		return nil
	}

	f, err := os.Create(common.HashFile)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(Store)
}

// SetupGracefulShutdown 监听信号并优雅关闭
func SetupGracefulShutdown() {
	// 统一处理退出信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("收到退出信号，保存 hash_store.json ...")
	if err := saveHashStore(); err != nil {
		log.Printf("保存失败: %v", err)
	}
	log.Println("退出完成")
}
