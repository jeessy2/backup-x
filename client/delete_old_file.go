package client

import (
	"backup-x/entity"
	"backup-x/util"
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
			// empty project and disabeld
			if !backupConf.NotEmptyProject() || backupConf.Enabled == 1 {
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
	backupFiles, err := os.ReadDir(backupConf.GetProjectPath())
	if err != nil {
		log.Printf("读取项目 %s 的本地目录失败! ERR: %s\n", backupConf.ProjectName, err)
	}
	if backupConf.SaveDays <= 0 {
		log.Printf("项目 %s 的本地保存(天)设置不正确", backupConf.ProjectName)
		return
	}
	backupFileNames := make([]string, len(backupFiles))
	for _, backupFile := range backupFiles {
		if !backupFile.IsDir() {
			info, err := backupFile.Info()
			if err == nil {
				if info.Size() >= minFileSize {
					backupFileNames = append(backupFileNames, backupFile.Name())
				} else {
					if util.IsFileNameDate(backupFile.Name()) {
						log.Printf("备份后的大小为 %d 字节，小于最低值 %d，将删除备份文件: %s", info.Size(), minFileSize, backupConf.GetProjectPath()+string(os.PathSeparator)+backupFile.Name())
						os.Remove(backupConf.GetProjectPath() + string(os.PathSeparator) + backupFile.Name())
					}
				}
			}
		}
	}

	tobeDeleteFiles := util.FileNameBeforeDays(backupConf.SaveDays, backupFileNames, backupConf.ProjectName)

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
	if backupConf.SaveDaysS3 <= 0 {
		log.Printf("项目 %s 的对象存储保存(天)设置不正确", backupConf.ProjectName)
		return
	}
	fileNames, err := s3Conf.ListFiles(backupConf.GetProjectPath())
	if err != nil {
		log.Printf("读取项目 %s 的对象存储目录失败! ERR: %s\n", backupConf.ProjectName, err)
	}

	tobeDeleteFiles := util.FileNameBeforeDays(backupConf.SaveDaysS3, fileNames, backupConf.ProjectName)

	for i := 0; i < len(tobeDeleteFiles); i++ {
		err := s3Conf.DeleteFile(tobeDeleteFiles[i])
		if err == nil {
			log.Printf("删除过期的文件(对象存储) %s 成功", tobeDeleteFiles[i])
		} else {
			log.Printf("删除过期的文件(对象存储) %s 失败: %s", tobeDeleteFiles[i], err)
		}
	}
}
