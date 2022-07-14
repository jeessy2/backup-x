# backup-x

<a href="https://github.com/jeessy2/backup-x/releases/latest"><img alt="GitHub release" src="https://img.shields.io/github/release/jeessy2/backup-x.svg?logo=github&style=flat-square"></a> <img src=https://goreportcard.com/badge/github.com/jeessy2/backup-x /> <img src=https://img.shields.io/docker/image-size/jeessy/backup-x /> <img src=https://img.shields.io/docker/pulls/jeessy/backup-x /> 

  带Web界面的数据库/文件备份增强工具。原理：执行自定义shell命令输出文件，增强备份功能。同时支持: 文件、mysql、postgres... [English](README-EN.md)
  - [x] 支持自定义命令
  - [x] 支持执行shell输出的文件备份，原理上支持各种数据库/文件备份
  - [x] 支持备份周期设置，几分钟到一年的备份周期也可以
  - [x] 支持多个项目备份，最多16个
  - [x] 支持备份后的文件另存到对象存储中 (在也不怕删库跑路了)
  - [x] 可设置备份文件最大保存天数
  - [x] webhook通知
  - [x] 支持加密关键信息(v1.3.0+)

## docker中使用
- 运行docker容器（`/opt/backup-x-files`可替换为主机的目录）
  ```
  docker run -d --name backup-x --restart=always \
    -p 9977:9977 \
    -v /opt/backup-x-files:/app/backup-x-files \
    jeessy/backup-x
  ```
- 登录 http://your_docker_ip:9977 并配置
- docker容器默认安装default-mysql-client/postgres-client/[s3sync](https://github.com/larrabee/s3sync)

## 系统中使用
- 下载并解压[https://github.com/jeessy2/backup-x/releases](https://github.com/jeessy2/backup-x/releases)
- 安装服务
  - Mac/Linux: `./backup-x -s install` 
  - Win(打开cmd): `.\backup-x.exe -s install`
  - 自定义参数 `./backup-x -s install -l 127.0.0.1:9977 -d /Users/name`
    - `-l` 监听地址（默认监听`:9977`）
    - `-d` 自定义备份目录地址（默认当前运行目录）
- [可选] 服务卸载
  - Mac/Linux: `./backup-x -s uninstall` 
  - Win(打开cmd): `.\backup-x.exe -s uninstall`
- 登录 http://127.0.0.1:9977 并配置

  ![avatar](https://raw.githubusercontent.com/jeessy2/backup-x/master/backup-x-web.png)

## 备份脚本参考
 - postgres

    |  说明   | 备份脚本  |
    |  ----  | ----  |
    | 备份单个  | PGPASSWORD="#{PWD}" pg_dump --host 192.168.1.11 --port 5432 --dbname db-name --user postgres --clean --create --file #{DATE}.sql |
    | 备份全部  | PGPASSWORD="#{PWD}" pg_dumpall --host 192.168.1.11 --port 5432 --user postgres --clean --file #{DATE}.sql |
    | 还原  | psql -U postgres -f 2021-11-12_10_29.sql |

 -  mysql/mariadb

    |  说明   | 备份脚本  |
    |  ----  | ----  |
    | 备份单个  | mysqldump -h192.168.1.11 -uroot -p#{PWD} db-name > #{DATE}.sql |
    | 备份全部  | mysqldump -h192.168.1.11 -uroot -p#{PWD} --all-databases > #{DATE}.sql |
    | 还原  | mysql -uroot -p123456 db-name <2021-11-12_10_29.sql |

 -  文件

    |  说明   | 备份脚本  |
    |  ----  | ----  |
    | 备份本地文件到对象存储 [s3sync](https://github.com/larrabee/s3sync) | s3sync --fs-disable-xattr --filter-not-exist --tk #{AccessKey} --ts #{SecretKey} --te #{Endpoint} fs:///opt/test/ s3://#{BucketName}/test/ |
    | 备份对象存储到对象存储 [s3sync](https://github.com/larrabee/s3sync) | s3sync --filter-not-exist --sk source_key -ss #{PWD} --se https://s3.source.com --tk #{AccessKey} --ts #{SecretKey} --te #{Endpoint} s3://backup/ s3://#{BucketName}/ |
    | 备份本地文件到对象存储 [rclone](https://rclone.org/) | 参考rclone文档。推荐在系统中使用rclone |

  - 变量说明

    |  变量名   | 说明  |
    |  ----  | ----  |
    |  #{DATE}  | 年-月-日_时_分  |
    |  #{PWD}   | 下方的密码变量  |
    |  #{Endpoint}  | 下方的对象存储变量 Endpoint  |
    |  #{AccessKey}  | 下方的对象存储变量 AccessKey  |
    |  #{SecretKey}  | 下方的对象存储变量 SecretKey  |
    |  #{BucketName}  | 下方的对象存储变量 BucketName  |
## webhook
- 支持webhook, 备份更新成功或不成功时, 会回调填写的URL
- 支持的变量

  |  变量名   | 说明  |
  |  ----  | ----  |
  | #{projectName}  | 项目名称 |
  | #{fileName}  | 备份后的文件名称 |
  | #{fileSize}  | 文件大小 (MB) |
  | #{result}  | 备份结果（成功/失败） |

- RequestBody为空GET请求，不为空POST请求
- Server酱: `https://sc.ftqq.com/[SCKEY].send?text=#{projectName}项目备份#{result},文件名:#{fileName},文件大小:#{fileSize}`
- Bark: `https://api.day.app/[YOUR_KEY]/#{projectName}项目备份#{result},文件名:#{fileName},文件大小:#{fileSize}`
- 钉钉:
  - 钉钉电脑端 -> 群设置 -> 智能群助手 -> 添加机器人 -> 自定义
  - 只勾选 `自定义关键词`, 输入的关键字必须包含在RequestBody的content中, 如：`项目备份`
  - URL中输入钉钉给你的 `Webhook地址`
  - RequestBody中输入 `{"msgtype": "text","text": {"content": "#{projectName}项目备份#{result},文件名:#{fileName},文件大小:#{fileSize}"}}`

## 说明
  - 从backup-db发展而来，发现不仅仅支持数据库备份，所以改名backup-x，备份届的iphone-x
