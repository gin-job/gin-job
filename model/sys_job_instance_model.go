package model

import "time"

// SysJobInstanceModel 任务执行实例表
type SysJobInstanceModel struct {
	BaseModel
	JobName    string     `json:"job_name" gorm:"index;size:128;not null"` // 任务名称
	JobID      uint       `json:"job_id" gorm:"index;not null"`            // 任务ID
	Status     string     `json:"status" gorm:"size:32;not null"`          // 执行状态：running, success, failed
	StartedAt  time.Time  `json:"started_at" gorm:"not null"`              // 开始执行时间
	FinishedAt *time.Time `json:"finished_at"`                             // 完成时间
	DurationMs int64      `json:"duration_ms"`                             // 执行耗时（毫秒）
	Error      string     `json:"error" gorm:"type:text"`                  // 错误信息
	LogContent string     `json:"log_content" gorm:"type:text"`            // 执行日志内容
}

// TableName 指定表名
func (c *SysJobInstanceModel) TableName() string {
	return "tb_sys_job_instance"
}
