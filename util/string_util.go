package util

import (
	"strings"
)

// EscapeShell 转义shell输出
func EscapeShell(org string) (dst string) {
	// 双引号使用单引号替换
	return strings.ReplaceAll(org, "\"", "'")
}
