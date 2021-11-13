package client

// RunCycle 周期运行
func RunCycle() {
	// delete old backup
	go DeleteOldBackup()
	// start client
	go StartBackup()
}
