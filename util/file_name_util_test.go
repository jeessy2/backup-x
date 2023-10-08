package util

import (
	"strconv"
	"testing"
	"time"
)

// TestFileNameUtil
func TestFileNameUtil(t *testing.T) {
	const days = 10
	beforeTomorrow, _ := time.ParseDuration("-" + strconv.Itoa((days+1)*24) + "h")
	fileNames := []string{
		"a2020-10-10-11-12b.sql",
		"测试2021-10-10-11-12测试.sql",
		time.Now().Add(beforeTomorrow).Format(FileNameFormatStr) + ".sql",
	}
	if len(fileNames) != len(FileNameBeforeDays(10, fileNames)) {
		t.Error("TestFileNameUtil Test failed!")
	}
}
