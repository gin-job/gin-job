package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	Id        int64          `json:"id" gorm:"type:bigint(20);primaryKey" primaryKey:"yes"`                     // id
	CreatedAt time.Time      `json:"created_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP;"`                // Created Time
	UpdatedAt time.Time      `json:"updated_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime;"` // Updated Time
	DeletedAt gorm.DeletedAt `json:"-" gorm:"type:datetime;"`                                                   // Deleted At
}
