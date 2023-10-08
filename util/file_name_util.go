package util

import (
	"log"
	"regexp"
	"strconv"
	"time"
)

const FileNameFormatStr = "2006-01-02-15-04"

// FileNameBeforeDays 查找文件名中有多少在指定天数之前的
func FileNameBeforeDays(days int, fileNames []string, projectName string) []string {
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
	// 待删除的过期文件为所有的文件，将不会进行删除
	if len(oldFiles) > 0 && len(oldFiles)-len(fileNames) >= 0 {
		log.Printf("项目 %s 待删除的过期文件为所有的文件，将不会进行删除！\n", projectName)
		return []string{}
	}
	return oldFiles
}
