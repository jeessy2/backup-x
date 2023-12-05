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
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// backupLooper
type backupLooper struct {
	Wg      sync.WaitGroup
	Tickers []*time.Ticker
}

var bl = &backupLooper{Wg: sync.WaitGroup{}}

// RunLoop backup db loop
func RunLoop() {
	conf, err := entity.GetConfigCache()
	if err != nil {
		return
	}

	// clear
	bl.Tickers = []*time.Ticker{}

	for _, backupConf := range conf.BackupConfig {
		if !backupConf.NotEmptyProject() {
			continue
		}

		if backupConf.Enabled != 0 {
			log.Println(backupConf.ProjectName + " 项目被停用")
			continue
		}

		if !backupConf.CheckPeriod() {
			log.Println(backupConf.ProjectName + " 项目的周期值不正确")
			continue
		}

		delay := util.GetDelaySeconds(backupConf.StartTime)
		ticker := time.NewTicker(delay)
		log.Printf("%s项目将在%.1f小时后运行\n", backupConf.ProjectName, delay.Hours())

		bl.Wg.Add(1)
		go func(backupConf entity.BackupConfig) {
			defer bl.Wg.Done()
			for {
				<-ticker.C
				run(conf, backupConf)
				ticker.Reset(time.Minute * time.Duration(backupConf.Period))
				log.Printf("%s项目将等待%d分钟后循环运行\n", backupConf.ProjectName, backupConf.Period)
			}
		}(backupConf)
		bl.Tickers = append(bl.Tickers, ticker)
	}

	bl.Wg.Wait()

}

// StopRunLoop
func StopRunLoop() {
	for _, ticker := range bl.Tickers {
		if ticker != nil {
			ticker.Stop()
		}
	}
}

// RunOnce 运行一次
func RunOnce() {
	conf, err := entity.GetConfigCache()
	if err != nil {
		return
	}

	for _, backupConf := range conf.BackupConfig {
		run(conf, backupConf)
	}
}

// 运行指定的索引号
func RunByIdx(idx int) {
	conf, err := entity.GetConfigCache()
	if err != nil {
		return
	}

	run(conf, conf.BackupConfig[idx])
}

// run
func run(conf entity.Config, backupConf entity.BackupConfig) {
	if backupConf.NotEmptyProject() && backupConf.Enabled == 0 {
		err := prepare(backupConf)
		if err != nil {
			log.Println(err)
			return
		}
		// backup
		outFileName, err := backup(backupConf, conf.EncryptKey, conf.S3Config)
		result := entity.BackupResult{ProjectName: backupConf.ProjectName, Result: "失败"}
		if err == nil {
			// webhook
			if outFileName != nil {
				result.FileName = outFileName.Name()
				result.FileSize = fmt.Sprintf("%d MB", outFileName.Size()/1000/1000)
				// send file to s3
				if conf.S3Config.CheckNotEmpty() {
					go conf.S3Config.UploadFile(backupConf.GetProjectPath() + string(os.PathSeparator) + outFileName.Name())
				}
			}
			result.Result = "成功"
		}
		conf.ExecWebhook(result)
	}
}

// prepare
func prepare(backupConf entity.BackupConfig) (err error) {
	// create floder
	os.MkdirAll(backupConf.GetProjectPath(), 0750)

	return
}

func backup(backupConf entity.BackupConfig, encryptKey string, s3Conf entity.S3Config) (outFileName os.FileInfo, err error) {
	projectName := backupConf.ProjectName
	log.Printf("正在备份项目: %s ...", projectName)

	todayString := time.Now().Format(util.FileNameFormatStr)
	shellString := strings.ReplaceAll(backupConf.Command, "#{DATE}", todayString)

	// 解密pwd
	pwd := ""
	if backupConf.Pwd != "" {
		pwd, err = util.DecryptByEncryptKey(encryptKey, backupConf.Pwd)
		if err != nil {
			err = fmt.Errorf("解密失败")
			log.Println(err)
			return nil, err
		}
	}
	// 解密s3 SecretKey
	secretKey := ""
	if s3Conf.SecretKey != "" {
		secretKey, err = util.DecryptByEncryptKey(encryptKey, s3Conf.SecretKey)
		if err != nil {
			err = fmt.Errorf("解密失败")
			log.Println(err)
			return nil, err
		}
	}

	shellString = strings.ReplaceAll(shellString, "#{PWD}", pwd)
	shellString = strings.ReplaceAll(shellString, "#{AccessKey}", s3Conf.AccessKey)
	shellString = strings.ReplaceAll(shellString, "#{SecretKey}", secretKey)
	shellString = strings.ReplaceAll(shellString, "#{Endpoint}", s3Conf.Endpoint)
	shellString = strings.ReplaceAll(shellString, "#{BucketName}", s3Conf.BucketName)

	// create shell file
	var shellName string
	if runtime.GOOS == "windows" {
		shellName = time.Now().Format("shell-"+util.FileNameFormatStr+"-") + "backup.bat"
	} else {
		shellName = time.Now().Format("shell-"+util.FileNameFormatStr+"-") + "backup.sh"
	}

	shellFile, err := os.Create(backupConf.GetProjectPath() + string(os.PathSeparator) + shellName)
	shellFile.Chmod(0700)
	if err == nil {
		shellFile.WriteString(shellString)
		shellFile.Close()
	} else {
		log.Println("Create file with error: ", err)
	}

	// run shell file
	var shell *exec.Cmd
	if runtime.GOOS == "windows" {
		shell = exec.Command("cmd", "/c", shellName)
	} else {
		shell = exec.Command("bash", shellName)
	}
	shell.Dir = backupConf.GetProjectPath()
	outputBytes, err := shell.CombinedOutput()
	if len(outputBytes) > 0 {
		if util.IsGBK(outputBytes) {
			outputBytes, _ = util.GbkToUtf8(outputBytes)
		}
		log.Printf("<span style='color: #7983f5;font-weight: bold;'>%s</span> 执行shell的输出: <span class='click-layer' onclick='showLayer(this)' tip=\"%s\" style='cursor: pointer; color: #4a3a3a; font-weight: bold; border: 2px dashed;'>点击此处查看</span>\n", backupConf.ProjectName, util.EscapeShell(string(outputBytes)))
	} else {
		log.Printf("执行shell的输出为空\n")
	}

	// execute shell success
	if err == nil {
		// find backup file by todayString
		outFileName, err = findBackupFile(backupConf, todayString)
		if backupConf.BackupType == 0 {
			// 备份数据库
			// check file size
			if err != nil {
				log.Println(err)
			} else if outFileName.Size() >= 200 {
				log.Printf("成功备份项目: %s, 文件名: %s\n", projectName, outFileName.Name())
			} else {
				err = errors.New(projectName + " 备份后的文件大小小于200字节, 当前大小：" + strconv.Itoa(int(outFileName.Size())))
				log.Println(err)
			}
		} else {
			// 1 同步文件
			// err = nil
			err = nil
		}
	} else {
		err = fmt.Errorf("执行备份shell失败: %s", util.EscapeShell(string(outputBytes)))
		log.Println(err)
	}

	// remove shell file
	os.Remove(shellFile.Name())

	return
}

// find backup file by todayString
func findBackupFile(backupConf entity.BackupConfig, todayString string) (backupFile os.FileInfo, err error) {
	files, err := ioutil.ReadDir(backupConf.GetProjectPath())
	for _, file := range files {
		if strings.Contains(file.Name(), todayString) && !strings.HasPrefix(file.Name(), "shell-") {
			backupFile = file
			return
		}
	}

	err = fmt.Errorf("项目 %s 没有输出包含 %s 的文件名", backupConf.ProjectName, todayString)

	return
}
