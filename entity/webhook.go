package entity

import (
	"backup-x/util"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Webhook Webhook
type Webhook struct {
	WebhookURL         string
	WebhookRequestBody string
}

// ExecWebhook 添加或更新IPv4/IPv6记录
func (webhook Webhook) ExecWebhook(result BackupResult) {

	if webhook.WebhookURL != "" {
		// 成功和失败都要触发webhook
		method := "GET"
		postPara := ""
		contentType := "application/x-www-form-urlencoded"
		if webhook.WebhookRequestBody != "" {
			method = "POST"
			postPara = webhook.replaceBody(result)
			if json.Valid([]byte(postPara)) {
				contentType = "application/json"
			}
		}
		requestURL := webhook.replaceURL(result)
		u, err := url.Parse(requestURL)
		if err != nil {
			log.Println("Webhook配置中的URL不正确")
			return
		}
		req, err := http.NewRequest(method, fmt.Sprintf("%s://%s%s?%s", u.Scheme, u.Host, u.Path, u.Query().Encode()), strings.NewReader(postPara))
		if err != nil {
			log.Println("创建Webhook请求异常, Err:", err)
			return
		}
		req.Header.Add("content-type", contentType)

		clt := http.Client{}
		clt.Timeout = 30 * time.Second
		resp, err := clt.Do(req)
		body, err := util.GetHTTPResponseOrg(resp, requestURL, err)
		if err == nil {
			log.Println(fmt.Sprintf("Webhook调用成功, 返回数据: %s", string(body)))
		} else {
			log.Println(fmt.Sprintf("Webhook调用失败，Err：%s", err))
		}
	}
}

// replaceURL 替换url
func (webhook Webhook) replaceURL(result BackupResult) (newBody string) {
	newBody = strings.ReplaceAll(webhook.WebhookURL, "#{projectName}", result.ProjectName)
	newBody = strings.ReplaceAll(newBody, "#{fileName}", result.FileName)
	newBody = strings.ReplaceAll(newBody, "#{fileSize}", result.FileSize)
	newBody = strings.ReplaceAll(newBody, "#{result}", result.Result)

	return newBody
}

// replaceBody 替换body
func (webhook Webhook) replaceBody(result BackupResult) (newURL string) {
	newURL = strings.ReplaceAll(webhook.WebhookRequestBody, "#{projectName}", result.ProjectName)
	newURL = strings.ReplaceAll(newURL, "#{fileName}", result.FileName)
	newURL = strings.ReplaceAll(newURL, "#{fileSize}", result.FileSize)
	newURL = strings.ReplaceAll(newURL, "#{result}", result.Result)

	return newURL
}
