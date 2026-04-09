package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	Id        int64          `json:"id" gorm:"type:bigint(20);primaryKey" primaryKey:"yes"`                     // id
	CreatedAt time.Time      `json:"created_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP;"`                // Created Time
	CreatedBy string         `json:"create_by" gorm:"type:varchar(50);"`                                        // Created By
	UpdatedAt time.Time      `json:"updated_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime;"` // Updated Time
	UpdatedBy string         `json:"update_by" gorm:"type:varchar(50);"`                                        // Updated By
	DeletedAt gorm.DeletedAt `json:"-" gorm:"type:datetime;"`                                                   // Deleted At
}
