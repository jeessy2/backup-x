package client

import (
	"backup-x/entity"
	"backup-x/util"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

// DeleteOldBackup for client
func DeleteOldBackup() {
	for {
		delay := util.GetDelaySeconds(2)
		log.Printf("删除过期的备份文件将在 %.1f 小时后运行\n", delay.Hours())
		time.Sleep(delay)

		conf, err := entity.GetConfigCache()
		if err != nil {
			return
		}

		for _, backupConf := range conf.BackupConfig {
			if !backupConf.NotEmptyProject() {
				continue
			}
			// read from current path
			backupFiles, err := ioutil.ReadDir(backupConf.GetProjectPath())
			if err != nil {
				log.Printf("读取项目 %s 目录失败! ERR: %s\n", backupConf.ProjectName, err)
				continue
			}

			// delete client files
			subDuration, _ := time.ParseDuration("-" + strconv.Itoa(backupConf.SaveDays*24) + "h")
			before := time.Now().Add(subDuration)

			// delete older file when file numbers gt MaxSaveDays
			for _, backupFile := range backupFiles {
				if backupFile.ModTime().Before(before) {
					filepath := backupConf.GetProjectPath() + string(os.PathSeparator) + backupFile.Name()
					err := os.Remove(filepath)
					if err != nil {
						log.Printf("删除过期的文件 %s 失败", filepath)
					} else {
						log.Printf("删除过期的文件 %s 成功", filepath)
					}
				}
			}
		}
	}

}
