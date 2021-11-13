package util

import (
	"log"
	"os"
	"time"
)

// PathExists Get path exist
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// SleepForFileDelete Sleep For File Delete
func SleepForFileDelete() {
	sleepHours := 24 - time.Now().Hour()
	log.Printf("%d小时后再次运行：删除过期的备份文件", sleepHours)
	time.Sleep(time.Hour * time.Duration(sleepHours))
}
