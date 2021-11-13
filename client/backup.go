package client

import (
	"backup-x/entity"
	"backup-x/util"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// StartBackup start backup db
func StartBackup() {
	for {
		RunOnce()
		// sleep to tomorrow night
		sleep()
	}
}

// RunOnce 运行一次
func RunOnce() {
	conf, err := entity.GetConfigCache()
	if err != nil {
		return
	}
	// 迭代所有项目
	for _, backupConf := range conf.BackupConfig {
		if backupConf.NotEmptyProject() {
			err := prepare(backupConf)
			if err != nil {
				log.Println(err)
				continue
			}
			// backup
			outFileName, err := backup(backupConf)
			result := entity.BackupResult{ProjectName: backupConf.ProjectName, Result: "失败"}
			if err == nil {
				// webhook
				result.FileName = outFileName.Name()
				result.FileSize = fmt.Sprintf("%d MB", outFileName.Size()/1000/1000)
				result.Result = "成功"
				// send file to s3
				go conf.UploadFile(backupConf.GetProjectPath() + string(os.PathSeparator) + outFileName.Name())
			}
			conf.ExecWebhook(result)
		}
	}
}

// prepare
func prepare(backupConf entity.BackupConfig) (err error) {
	// create floder
	os.MkdirAll(backupConf.GetProjectPath(), 0750)

	if !strings.Contains(backupConf.Command, "#{DATE}") {
		err = errors.New("项目: " + backupConf.ProjectName + "的备份脚本须包含#{DATE}")
	}

	return
}

func backup(backupConf entity.BackupConfig) (outFileName os.FileInfo, err error) {
	projectName := backupConf.ProjectName
	log.Printf("正在备份项目: %s ...", projectName)

	todayString := time.Now().Format("2006-01-02_03_04")
	shellString := strings.ReplaceAll(backupConf.Command, "#{DATE}", todayString)

	// create shell file
	shellName := time.Now().Format("shell-2006-01-02-03-04-") + "backup.sh"

	shellFile, err := os.Create(backupConf.GetProjectPath() + string(os.PathSeparator) + shellName)
	shellFile.Chmod(0700)
	if err == nil {
		shellFile.WriteString(shellString)
		shellFile.Close()
	} else {
		log.Println("Create file with error: ", err)
	}

	// run shell file
	shell := exec.Command("bash", shellName)
	shell.Dir = backupConf.GetProjectPath()
	outputBytes, err := shell.CombinedOutput()
	if len(outputBytes) > 0 {
		log.Printf("<span title=\"%s\">执行shell的输出：鼠标移动此处查看</span>", util.EscapeShell(string(outputBytes)))
	} else {
		log.Printf("执行shell的输出为空")
	}

	// execute shell success
	if err == nil {
		// find backup file by todayString
		outFileName, err = findBackupFile(backupConf, todayString)

		// check file size
		if err != nil {
			log.Println(err)
		} else if outFileName.Size() >= 500 {
			log.Printf("成功备份项目: %s, 文件名: %s\n", projectName, outFileName.Name())
			// success, remove shell file
			os.Remove(shellFile.Name())
		} else {
			err = errors.New(projectName + " 备份后的文件大小小于500字节, 当前大小：" + strconv.Itoa(int(outFileName.Size())))
			log.Println(err)
		}
	} else {
		err = fmt.Errorf("执行备份shell失败: %s", util.EscapeShell(string(outputBytes)))
		log.Println(err)
	}

	return
}

// find backup file by todayString
func findBackupFile(backupConf entity.BackupConfig, todayString string) (backupFile os.FileInfo, err error) {
	files, err := ioutil.ReadDir(backupConf.GetProjectPath())
	for _, file := range files {
		if strings.Contains(file.Name(), todayString) {
			backupFile = file
			return
		}
	}
	err = errors.New("不能找到备份后的文件，没有找到包含 " + todayString + " 的文件名")
	return
}

func sleep() {
	sleepHours := 24 - time.Now().Hour()
	log.Println("下次运行时间：", sleepHours, "hours")
	time.Sleep(time.Hour * time.Duration(sleepHours))
}
