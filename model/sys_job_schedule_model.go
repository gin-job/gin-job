package model

import "time"

// SysJobScheduleModel Job Schedule Model
type SysJobScheduleModel struct {
	BaseModel
	Name        string     `json:"name" gorm:"uniqueIndex;size:128"`      // Job Schedule Name
	HandlerName string     `json:"handler_name" gorm:"size:128;not null"` // Handler Name
	Spec        string     `json:"spec" gorm:"size:128"`                  // Cron Spec
	Enabled     bool       `json:"enabled" gorm:"not null;default:true"`  // Enabled
	Description string     `json:"description" gorm:"size:256"`           // Description
	LastRunAt   *time.Time `json:"last_run_at"`                           // Last Run At
	Status      string     `json:"status" gorm:"size:32"`                 // Status
	LastError   string     `json:"last_error" gorm:"size:512"`            // Last Error
}

// TableName Job Schedule Table Name
func (SysJobScheduleModel) TableName() string {
	return "tb_sys_job_schedule"
}
