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

		if !backupConf.CheckPeriod() {
			log.Println(backupConf.ProjectName + "的周期值不正确")
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

// run
func run(conf entity.Config, backupConf entity.BackupConfig) {
	if backupConf.NotEmptyProject() {
		err := prepare(backupConf)
		if err != nil {
			log.Println(err)
			return
		}
		// backup
		outFileName, err := backup(backupConf, conf.EncryptKey)
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

// prepare
func prepare(backupConf entity.BackupConfig) (err error) {
	// create floder
	os.MkdirAll(backupConf.GetProjectPath(), 0750)

	if !strings.Contains(backupConf.Command, "#{DATE}") {
		err = errors.New("项目: " + backupConf.ProjectName + "的备份脚本须包含#{DATE}")
	}

	return
}

func backup(backupConf entity.BackupConfig, encryptKey string) (outFileName os.FileInfo, err error) {
	projectName := backupConf.ProjectName
	log.Printf("正在备份项目: %s ...", projectName)

	todayString := time.Now().Format("2006-01-02_15_04")
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

	shellString = strings.ReplaceAll(shellString, "#{PWD}", pwd)

	// create shell file
	var shellName string
	if runtime.GOOS == "windows" {
		shellName = time.Now().Format("shell-2006-01-02-15-04-") + "backup.bat"
	} else {
		shellName = time.Now().Format("shell-2006-01-02-15-04-") + "backup.sh"
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
		log.Printf("<span title=\"%s\">%s 执行shell的输出：鼠标移动此处查看</span>\n", util.EscapeShell(string(outputBytes)), backupConf.ProjectName)
	} else {
		log.Printf("执行shell的输出为空\n")
	}

	// execute shell success
	if err == nil {
		// find backup file by todayString
		outFileName, err = findBackupFile(backupConf, todayString)

		// check file size
		if err != nil {
			log.Println(err)
		} else if outFileName.Size() >= 200 {
			log.Printf("成功备份项目: %s, 文件名: %s\n", projectName, outFileName.Name())
			// success, remove shell file
			os.Remove(shellFile.Name())
		} else {
			err = errors.New(projectName + " 备份后的文件大小小于200字节, 当前大小：" + strconv.Itoa(int(outFileName.Size())))
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
