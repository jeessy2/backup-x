package client

import (
	"backup-x/entity"
	"backup-x/util"
	"io/ioutil"
	"log"
	"os"
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
			// 删除本地的过时文件
			deleteLocalOlderFiles(backupConf)

			// 删除对象存储中的过时文件
			deleteS3OlderFiles(conf.S3Config, backupConf)
		}
	}

}

// deleteLocalOlderFiles 删除本地的过时文件
func deleteLocalOlderFiles(backupConf entity.BackupConfig) {
	backupFiles, err := ioutil.ReadDir(backupConf.GetProjectPath())
	if err != nil {
		log.Printf("读取项目 %s 的本地目录失败! ERR: %s\n", backupConf.ProjectName, err)
	}
	backupFileNames := make([]string, len(backupFiles))
	for _, backupFile := range backupFiles {
		backupFileNames = append(backupFileNames, backupFile.Name())
	}

	tobeDeleteFiles := util.FileNameBeforeDays(backupConf.SaveDays, backupFileNames)

	for i := 0; i < len(tobeDeleteFiles); i++ {
		err := os.Remove(backupConf.GetProjectPath() + string(os.PathSeparator) + tobeDeleteFiles[i])
		if err == nil {
			log.Printf("删除过期的文件(本地) %s 成功", backupConf.ProjectName+string(os.PathSeparator)+tobeDeleteFiles[i])
		} else {
			log.Printf("删除过期的文件(本地) %s 失败: %s", backupConf.ProjectName+string(os.PathSeparator)+tobeDeleteFiles[i], err)
		}
	}
}

// deleteS3OlderFiles 删除对象存储的过时文件
func deleteS3OlderFiles(s3Conf entity.S3Config, backupConf entity.BackupConfig) {
	if !s3Conf.CheckNotEmpty() {
		return
	}
	fileNames, err := s3Conf.ListFiles(backupConf.GetProjectPath())
	if err != nil {
		log.Printf("读取项目 %s 的对象存储目录失败! ERR: %s\n", backupConf.ProjectName, err)
	}

	tobeDeleteFiles := util.FileNameBeforeDays(backupConf.SaveDaysS3, fileNames)

	for i := 0; i < len(tobeDeleteFiles); i++ {
		err := s3Conf.DeleteFile(tobeDeleteFiles[i])
		if err == nil {
			log.Printf("删除过期的文件(对象存储) %s 成功", tobeDeleteFiles[i])
		} else {
			log.Printf("删除过期的文件(对象存储) %s 失败: %s", tobeDeleteFiles[i], err)
		}
	}
}
