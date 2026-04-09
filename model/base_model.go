package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	Id        int64          `json:"id" gorm:"type:bigint(20);comment:主键;primaryKey" primaryKey:"yes"`                      // 主键ID
	CreatedAt time.Time      `json:"created_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP;comment:创建时间"`                // 创建时间
	CreatedBy string         `json:"create_by" gorm:"type:varchar(50);comment:创建人"`                                         // 创建人
	UpdatedAt time.Time      `json:"updated_at" gorm:"type:datetime;default:CURRENT_TIMESTAMP;autoUpdateTime;comment:更新时间"` // 更新时间
	UpdatedBy string         `json:"update_by" gorm:"type:varchar(50);comment:更新人"`                                         // 更新人
	DeletedAt gorm.DeletedAt `json:"-" gorm:"type:datetime;comment:删除时间"`                                                   // 删除时间
}
