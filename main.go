package main

import (
	"backup-x/client"
	"backup-x/util"
	"backup-x/web"
	"embed"
	"flag"
	"net"
	"os"

	"log"
	"net/http"
	"time"

	"github.com/kardianos/service"
)

// 监听地址
var listen = flag.String("l", ":9977", "监听地址")

// 服务管理
var serviceType = flag.String("s", "", "服务管理, 支持install, uninstall")

// 默认备份路径当前运行目录
var backupDirDefault, _ = os.Getwd()

// 配置文件路径
var backupDir = flag.String("d", backupDirDefault, "自定义备份目录地址")

//go:embed static
var staticEmbededFiles embed.FS

//go:embed favicon.ico
var faviconEmbededFile embed.FS

// version
var version = "DEV"

func main() {
	flag.Parse()
	if _, err := net.ResolveTCPAddr("tcp", *listen); err != nil {
		log.Fatalf("解析监听地址异常，%s", err)
	}

	os.Setenv(web.VersionEnv, version)

	switch *serviceType {
	case "install":
		installService()
	case "uninstall":
		uninstallService()
	default:
		if util.IsRunInDocker() {
			run()
		} else {
			s := getService()
			status, _ := s.Status()
			if status != service.StatusUnknown {
				// 以服务方式运行
				s.Run()
			} else {
				// 非服务方式运行
				switch s.Platform() {
				case "windows-service":
					log.Println("可使用 .\\backup-x.exe -s install 安装服务运行")
				default:
					log.Println("可使用 ./backup-x -s install 安装服务运行")
				}
				run()
			}
		}
	}
}

func staticFsFunc(writer http.ResponseWriter, request *http.Request) {
	http.FileServer(http.FS(staticEmbededFiles)).ServeHTTP(writer, request)
}

func faviconFsFunc(writer http.ResponseWriter, request *http.Request) {
	http.FileServer(http.FS(faviconEmbededFile)).ServeHTTP(writer, request)
}

func run() {
	// 启动静态文件服务
	http.HandleFunc("/static/", web.BasicAuth(staticFsFunc))
	http.HandleFunc("/favicon.ico", web.BasicAuth(faviconFsFunc))

	http.HandleFunc("/", web.BasicAuth(web.WritingConfig))
	http.HandleFunc("/save", web.BasicAuth(web.Save))
	http.HandleFunc("/logs", web.BasicAuth(web.Logs))
	http.HandleFunc("/clearLog", web.BasicAuth(web.ClearLog))
	http.HandleFunc("/webhookTest", web.BasicAuth(web.WebhookTest))

	// 改变工作目录
	os.Chdir(*backupDir)

	// 运行
	go client.DeleteOldBackup()
	go client.RunLoop()

	err := http.ListenAndServe(*listen, nil)

	if err != nil {
		log.Println("启动端口发生异常, 请检查端口是否被占用", err)
		time.Sleep(time.Minute)
	}
}

type program struct{}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}
func (p *program) run() {
	// 服务运行，延时20秒运行，等待网络
	time.Sleep(20 * time.Second)
	run()
}
func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	return nil
}

func getService() service.Service {
	options := make(service.KeyValue)
	var depends []string

	// 确保服务等待网络就绪后再启动
	switch service.ChosenSystem().String() {
	case "windows-service":
		// 将 Windows 服务的启动类型设为自动(延迟启动)
		options["DelayedAutoStart"] = true
	default:
		// 向 Systemd 添加网络依赖
		depends = append(depends, "Requires=network.target",
			"After=network-online.target")
	}

	svcConfig := &service.Config{
		Name:         "backup-x",
		DisplayName:  "backup-x",
		Description:  "带Web界面的数据库/文件备份增强工具",
		Arguments:    []string{"-l", *listen, "-d", *backupDir},
		Dependencies: depends,
		Option:       options,
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatalln(err)
	}
	return s
}

// 卸载服务
func uninstallService() {
	s := getService()

	status, _ := s.Status()
	// 处理卸载
	if status != service.StatusUnknown {
		s.Stop()
		if err := s.Uninstall(); err == nil {
			log.Println("backup-x 服务卸载成功!")
		} else {
			log.Printf("backup-x 服务卸载失败, ERR: %s\n", err)
		}
	} else {
		log.Printf("backup-x 服务未安装")
	}
}

// 安装服务
func installService() {
	s := getService()

	status, err := s.Status()
	if err != nil && status == service.StatusUnknown {
		// 服务未知，创建服务
		if err = s.Install(); err == nil {
			s.Start()
			log.Println("安装 backup-x 服务成功! 程序会一直运行, 包括重启后。")
			return
		}

		log.Printf("安装 backup-x 服务失败, ERR: %s\n", err)
		switch s.Platform() {
		case "windows-service":
			log.Println("请确保使用如下命令: .\\backup-x.exe -s install")
		default:
			log.Println("请确保使用如下命令: ./backup-x -s install")
		}
	}

	if status != service.StatusUnknown {
		log.Println("backup-x 服务已安装, 无需再次安装")
	}
}
