package main

import (
	"backup-x/web"
	"embed"
	"os"

	"log"
	"net/http"
	"time"
)

var defaultPort = "9977"

//go:embed static
var staticEmbededFiles embed.FS

//go:embed favicon.ico
var faviconEmbededFile embed.FS

func main() {
	// 启动静态文件服务
	http.Handle("/static/", http.FileServer(http.FS(staticEmbededFiles)))
	http.Handle("/favicon.ico", http.FileServer(http.FS(faviconEmbededFile)))

	http.HandleFunc("/", web.BasicAuth(web.WritingConfig))
	http.HandleFunc("/save", web.BasicAuth(web.Save))
	http.HandleFunc("/logs", web.BasicAuth(web.Logs))
	http.HandleFunc("/webhookTest", web.BasicAuth(web.WebhookTest))

	// 运行
	go web.Run()

	if os.Getenv("port") != "" {
		defaultPort = os.Getenv("port")
	}

	err := http.ListenAndServe(":"+defaultPort, nil)

	if err != nil {
		log.Println("启动端口发生异常, 请检查端口是否被占用", err)
		time.Sleep(time.Minute)
	}

}
