package models

import (
	"time"

	"github.com/google/uuid"
)

type SensorRequest struct {
	ID          int64     `gorm:"primaryKey;autoIncrement:true" json:"id"`
	CreatedAt   time.Time `gorm:"autoCreateTime:milli;column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime:milli;column:updated_at" json:"updated_at"`
	UUID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex;column:uuid" json:"uuid"`
	RequestData string    `gorm:"column:request_data" json:"request_data"`
}

type GatewayRequest struct {
	ID          int64     `gorm:"primaryKey;autoIncrement:true" json:"id"`
	CreatedAt   time.Time `gorm:"autoCreateTime:milli;column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime:milli;column:updated_at" json:"updated_at"`
	UUID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex;column:uuid" json:"uuid"`
	RequestData string    `gorm:"column:request_data" json:"request_data"`
}
