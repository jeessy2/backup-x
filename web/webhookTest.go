package web

import (
	"backup-x/entity"
	"log"
	"net/http"
	"strings"
)

// WebhookTest 测试webhook
func WebhookTest(writer http.ResponseWriter, request *http.Request) {
	url := strings.TrimSpace(request.FormValue("URL"))
	requestBody := strings.TrimSpace(request.FormValue("RequestBody"))
	if url != "" {
		wb := entity.Webhook{WebhookURL: url, WebhookRequestBody: requestBody}
		wb.ExecWebhook(entity.BackupResult{ProjectName: "模拟测试", FileName: "2021-11-11_01_01.sql", FileSize: "100 MB", Result: "成功"})
	} else {
		log.Println("请输入Webhook的URL")
	}

}
