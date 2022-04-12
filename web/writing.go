package web

import (
	"backup-x/entity"
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
)

//go:embed writing.html
var writingEmbedFile embed.FS

const VersionEnv = "BACKUP_X_VERSION"

type writtingData struct {
	entity.Config
	Version string
}

// WritingConfig 填写配置信息
func WritingConfig(writer http.ResponseWriter, request *http.Request) {
	tmpl, err := template.ParseFS(writingEmbedFile, "writing.html")
	if err != nil {
		log.Println(err)
		return
	}

	conf, err := entity.GetConfigCache()
	if err == nil {
		tmpl.Execute(writer, &writtingData{Config: conf, Version: os.Getenv(VersionEnv)})
		return
	}

	// default config
	// 获得环境变量
	backupConf := []entity.BackupConfig{}
	for i := 0; i < 16; i++ {
		backupConf = append(backupConf, entity.BackupConfig{SaveDays: 30, SaveDaysS3: 60, StartTime: 1, Period: 1440, BackupType: 0})
	}
	conf = entity.Config{
		BackupConfig: backupConf,
	}

	tmpl.Execute(writer, &writtingData{Config: conf, Version: os.Getenv(VersionEnv)})
}
