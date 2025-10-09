package utils

import (
	"fmt"
	"os"
	"store/common"
	"time"
)

func AppendUpdateLog(url, text string) error {
	// 固定写到 update.txt
	filename := common.FileName

	// 打开文件（追加模式）
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	// 生成日志条目
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf(
		"\n==== %s | %s ====\n%s\n",
		timestamp, url, text,
	)

	// 写入文件
	if _, err = f.WriteString(entry); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}
