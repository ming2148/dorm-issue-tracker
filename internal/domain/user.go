package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleResident   Role = "resident"
	RoleTechnician Role = "technician"
	RoleAdmin      Role = "admin"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	FullName     string         `gorm:"type:varchar(100);not null" json:"full_name"`
	Email        string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	PhoneNumber  string         `gorm:"type:varchar(20);not null" json:"phone_number"`
	RoomNo       string         `gorm:"type:varchar(20);index;not null" json:"room_no"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Role         Role           `gorm:"type:varchar(20);default:'resident';index;not null" json:"role"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
