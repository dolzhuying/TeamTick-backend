package cleaner

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// CleanFilesBySize 保证目录下所有 .xlsx 文件总大小不超过 maxSizeMB，超出则从最老的开始删除
func CleanFilesBySize(dir string, maxSizeMB int) {
	var files []os.FileInfo
	var totalSize int64
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && filepath.Ext(info.Name()) == ".xlsx" {
			files = append(files, info)
			totalSize += info.Size()
		}
		return nil
	})
	// 按修改时间从旧到新排序
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})
	// 超过限制则从最老的开始删
	for _, f := range files {
		if totalSize <= int64(maxSizeMB)*1024*1024 {
			break
		}
		path := filepath.Join(dir, f.Name())
		_ = os.Remove(path)
		totalSize -= f.Size()
		log.Printf("[cleaner] 删除超额文件: %s", path)
	}
}

// StartCleaner 启动定时清理任务，每 interval 执行一次
func StartCleaner(dir string, maxSizeMB int, interval time.Duration) {
	go func() {
		for {
			CleanFilesBySize(dir, maxSizeMB)
			time.Sleep(interval)
		}
	}()
}
