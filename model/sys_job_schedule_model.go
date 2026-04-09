package model

import "time"

// SysJobScheduleModel 任务调度表
type SysJobScheduleModel struct {
	BaseModel
	Name        string     `json:"name" gorm:"uniqueIndex;size:128"`      // 任务名称
	HandlerName string     `json:"handler_name" gorm:"size:128;not null"` // 关联的任务实现名称（handler）
	Spec        string     `json:"spec" gorm:"size:128"`                  // cron 表达式（执行频率）
	Enabled     bool       `json:"enabled" gorm:"not null;default:true"`  // 是否启用
	Description string     `json:"description" gorm:"size:256"`           // 任务描述
	LastRunAt   *time.Time `json:"last_run_at"`                           // 上次执行时间
	Status      string     `json:"status" gorm:"size:32"`                 // 任务状态
	LastError   string     `json:"last_error" gorm:"size:512"`            // 最后一次执行失败的错误信息，成功时可为空。
}

// TableName 指定表名
func (SysJobScheduleModel) TableName() string {
	return "tb_sys_job_schedule"
}
