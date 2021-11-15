package util

import "os"

// dockerEnvFile Docker容器中包含的文件
const dockerEnvFile string = "/.dockerenv"

// IsRunInDocker 是否在docker中运行
func IsRunInDocker() bool {
	_, err := os.Stat(dockerEnvFile)
	return err == nil
}
