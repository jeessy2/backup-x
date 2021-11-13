package web

import (
	"backup-x/entity"
	"net/http"
	"strconv"
	"strings"
)

// Save 保存
func Save(writer http.ResponseWriter, request *http.Request) {
	conf := &entity.Config{}

	// 覆盖以前的配置
	conf.Username = strings.TrimSpace(request.FormValue("Username"))
	conf.Password = request.FormValue("Password")

	forms := request.PostForm
	for index, projectName := range forms["ProjectName"] {
		saveDays, _ := strconv.Atoi(forms["SaveDays"][index])
		conf.BackupConfig = append(
			conf.BackupConfig,
			entity.BackupConfig{
				ProjectName: projectName,
				Command:     forms["Command"][index],
				SaveDays:    saveDays,
			},
		)
	}

	// Webhook
	conf.WebhookURL = strings.TrimSpace(request.FormValue("WebhookURL"))
	conf.WebhookRequestBody = strings.TrimSpace(request.FormValue("WebhookRequestBody"))

	// S3
	conf.Endpoint = strings.TrimSpace(request.FormValue("Endpoint"))
	conf.AccessKey = strings.TrimSpace(request.FormValue("AccessKey"))
	conf.SecretKey = strings.TrimSpace(request.FormValue("SecretKey"))
	conf.BucketName = strings.TrimSpace(request.FormValue("BucketName"))

	// 保存到用户目录
	err := conf.SaveConfig()

	// 没有错误，运行一次
	if err == nil {
		conf.CreateBucketIfNotExist()
		go RunOnce()
	}

	// 回写错误信息
	if err == nil {
		writer.Write([]byte("ok"))
	} else {
		writer.Write([]byte(err.Error()))
	}

}
