package common

import (
	"os"
	"path/filepath"
)

type Viewport struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

const (
	Url1     = "https://store.gavinnewsom.com/"
	Url2     = "https://store.gavinnewsom.com/the-patriot-shop/"
	Url3     = "https://stopelectionrigging.com/"
	Url4     = "https://bkokfi.com/"
	FileName = "update.txt"
)

var PngDir = map[string]string{
	Url1: "store",
	Url2: "shop",
	Url3: "california",
}

var PngView = map[string]Viewport{
	Url1: Viewport{Width: 1300, Height: 2470},
	Url2: Viewport{Width: 1300, Height: 2470},
	Url3: Viewport{Width: 1280, Height: 7100},
}

var HashFile = filepath.Join(getProjectRoot(), "hash_store.json")

func getProjectRoot() string {
	dir, _ := os.Getwd() // 程序启动时的工作目录
	return dir
}
