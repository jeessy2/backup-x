package web

import (
	"io"
	"log"
	"net/http"
	"os"
)

// MemoryLogs 内存中的日志
type MemoryLogs struct {
	MaxNum int      // 保存最大条数
	Logs   []string // 日志
}

func (mlogs *MemoryLogs) Write(p []byte) (n int, err error) {
	mlogs.Logs = append(mlogs.Logs, string(p))
	// 处理日志数量
	if len(mlogs.Logs) > mlogs.MaxNum {
		mlogs.Logs = mlogs.Logs[len(mlogs.Logs)-mlogs.MaxNum:]
	}
	return len(p), nil
}

var mlogs = &MemoryLogs{MaxNum: 50}

// 初始化日志
func init() {
	log.SetOutput(io.MultiWriter(mlogs, os.Stdout))
	// log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// Logs web
func Logs(writer http.ResponseWriter, request *http.Request) {
	for _, log := range mlogs.Logs {
		writer.Write([]byte(log))
		writer.Write([]byte("<br/>"))
	}
}

// ClearLog
func ClearLog(writer http.ResponseWriter, request *http.Request) {
	mlogs.Logs = mlogs.Logs[:0]
}
