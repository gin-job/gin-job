package model

import "time"

// SysJobInstanceModel Job Instance Model
type SysJobInstanceModel struct {
	BaseModel
	JobName    string     `json:"job_name" gorm:"index;size:128;not null"` // Job Name
	JobID      uint       `json:"job_id" gorm:"index;not null"`            // Job ID
	Status     string     `json:"status" gorm:"size:32;not null"`          // Status：running, success, failed
	StartedAt  time.Time  `json:"started_at" gorm:"not null"`              // Started At
	FinishedAt *time.Time `json:"finished_at"`                             // Finished At
	DurationMs int64      `json:"duration_ms"`                             // Duration Ms
	Error      string     `json:"error" gorm:"type:text"`                  // Error
	LogContent string     `json:"log_content" gorm:"type:text"`            // Log Content
}

// TableName Job Instance Table Name
func (c *SysJobInstanceModel) TableName() string {
	return "tb_sys_job_instance"
}
