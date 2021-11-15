package entity

// BackupConfig 备份配置
type BackupConfig struct {
	ProjectName string // 项目名称
	Command     string // 命令
	SaveDays    int
	StartTime   int // 开始时间(0-23)
	Period      int // 间隔周期(分钟)
}

// GetProjectPath 获得项目路径
func (backupConfig *BackupConfig) GetProjectPath() string {
	return parentSavePath + "/" + backupConfig.ProjectName
}

// NotEmptyProject 是不是空的项目
func (backupConfig *BackupConfig) NotEmptyProject() bool {
	return backupConfig.Command != "" && backupConfig.ProjectName != ""
}

// CheckPeriod 检测周期
func (backupConfig *BackupConfig) CheckPeriod() bool {
	return backupConfig.StartTime < 24 && backupConfig.StartTime >= 0 && backupConfig.Period > 0
}
