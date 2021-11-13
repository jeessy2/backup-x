package entity

import (
	"io/ioutil"
	"log"
	"os"
	"sync"

	"gopkg.in/yaml.v2"
)

// ParentSavePath Parent Save Path
const ParentSavePath = "backup-x-files"

func init() {
	_, err := os.Stat(ParentSavePath)
	if err != nil {
		os.Mkdir(ParentSavePath, 0750)
	}
}

// Config yml格式的配置文件
// go的实体需大写对应config.yml的key, key全部小写
type Config struct {
	User
	BackupConfig []BackupConfig
	Webhook
	S3Config
}

// ConfigCache ConfigCache
type cacheType struct {
	ConfigSingle *Config
	Err          error
	Lock         sync.Mutex
}

var cache = &cacheType{}

// GetConfigCache 获得配置
func GetConfigCache() (conf Config, err error) {

	cache.Lock.Lock()
	defer cache.Lock.Unlock()

	if cache.ConfigSingle != nil {
		return *cache.ConfigSingle, cache.Err
	}

	// init config
	cache.ConfigSingle = &Config{}

	configFilePath := getConfigFilePath()
	_, err = os.Stat(configFilePath)
	if err != nil {
		log.Println("没有找到配置文件！请在网页中输入")
		cache.Err = err
		return *cache.ConfigSingle, err
	}

	byt, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Println("config.yaml读取失败")
		cache.Err = err
		return *cache.ConfigSingle, err
	}

	err = yaml.Unmarshal(byt, cache.ConfigSingle)
	if err != nil {
		log.Println("反序列化配置文件失败", err)
		cache.Err = err
		return *cache.ConfigSingle, err
	}
	// remove err
	cache.Err = nil
	return *cache.ConfigSingle, err
}

// SaveConfig 保存配置
func (conf *Config) SaveConfig() (err error) {
	cache.Lock.Lock()
	defer cache.Lock.Unlock()

	byt, err := yaml.Marshal(conf)
	if err != nil {
		log.Println(err)
		return err
	}

	err = ioutil.WriteFile(getConfigFilePath(), byt, 0600)
	if err != nil {
		log.Println(err)
		return
	}

	// 清空配置缓存
	cache.ConfigSingle = nil

	return
}

// GetConfigFilePath 获得配置文件路径, 保存到备份目录下
func getConfigFilePath() string {
	return ParentSavePath + string(os.PathSeparator) + ".backup_x_config.yaml"
}
