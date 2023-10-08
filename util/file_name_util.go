package util

import (
	"regexp"
	"strconv"
	"time"
)

const FileNameFormatStr = "2006-01-02-15-04"

// FileNameBeforeDays 查找文件名中有多少在指定天数之前的
func FileNameBeforeDays(days int, fileNames []string) []string {
	oldFiles := make([]string, 0)
	// 2006-01-02-15-04
	fileRegxp := regexp.MustCompile(`([\d]{4})-([\d]{2})-([\d]{2})-([\d]{2})-([\d]{2})`)
	subDuration, _ := time.ParseDuration("-" + strconv.Itoa(days*24) + "h")
	before := time.Now().Add(subDuration)
	for i := 0; i < len(fileNames); i++ {
		dateString := fileRegxp.FindString(fileNames[i])
		if dateString != "" {
			if fileTime, err := time.Parse(FileNameFormatStr, dateString); err == nil && fileTime.Before(before) {
				oldFiles = append(oldFiles, fileNames[i])
			}
		}

	}
	return oldFiles
}
