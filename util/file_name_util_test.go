package util

import (
	"strconv"
	"testing"
	"time"
)

// TestFileNameUtil 测试正常情况
func TestFileNameUtil(t *testing.T) {
	const days = 10
	beforeTomorrow, _ := time.ParseDuration("-" + strconv.Itoa((days+1)*24) + "h")
	fileNames := []string{
		"a2020-10-10-11-12b.sql",
		"测试2021-10-10-11-12测试.sql",
		time.Now().Add(beforeTomorrow).Format(FileNameFormatStr) + ".sql",
		time.Now().Format(FileNameFormatStr) + ".sql",
	}
	deleteFiles := FileNameBeforeDays(days, fileNames, "test")
	// 只有一个不会被删除，需要+1
	if len(fileNames) != len(deleteFiles)+1 {
		t.Error("TestFileNameUtil Test failed!")
	}
}

// TestFileNameUtilAll 测试全部删除
func TestFileNameUtilAll(t *testing.T) {
	const days = 10
	beforeTomorrow, _ := time.ParseDuration("-" + strconv.Itoa((days+1)*24) + "h")
	fileNames := []string{
		"a2020-10-10-11-12b.sql",
		"测试2021-10-10-11-12测试.sql",
		time.Now().Add(beforeTomorrow).Format(FileNameFormatStr) + ".sql",
	}
	deleteFiles := FileNameBeforeDays(days, fileNames, "test")
	if len(deleteFiles) != 0 {
		t.Error("TestFileNameUtilAll Test failed!")
	}
}
