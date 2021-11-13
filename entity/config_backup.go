package entity

// BackupConfig 备份配置
type BackupConfig struct {
	ProjectName string // 项目名称
	Command     string // 命令
	SaveDays    int
}

// GetProjectPath 获得项目路径
func (backupConfig *BackupConfig) GetProjectPath() string {
	return ParentSavePath + "/" + backupConfig.ProjectName
}

// NotEmptyProject 是不是空的项目
func (backupConfig *BackupConfig) NotEmptyProject() bool {
	return backupConfig.Command != "" && backupConfig.ProjectName != ""
}
