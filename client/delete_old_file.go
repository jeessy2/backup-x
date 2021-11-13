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
	// sleep 30 minutes
	time.Sleep(30 * time.Minute)
	for {
		conf, err := entity.GetConfigCache()
		if err == nil {
			for _, backupConf := range conf.BackupConfig {
				// read from current path
				backupFiles, err := ioutil.ReadDir(backupConf.GetProjectPath())
				if err != nil {
					log.Println("Read dir with error :", err)
					continue
				}

				// delete client files
				ago := time.Now()
				for _, conf := range conf.BackupConfig {
					lastDay, _ := time.ParseDuration("-" + strconv.Itoa(conf.SaveDays*24) + "h")
					ago = ago.Add(lastDay)

					// delete older file when file numbers gt MaxSaveDays
					for _, backupFile := range backupFiles {
						if backupFile.ModTime().Before(ago) {
							filepath := backupConf.GetProjectPath() + "/" + backupFile.Name()
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
		// sleep
		util.SleepForFileDelete()
	}

}
