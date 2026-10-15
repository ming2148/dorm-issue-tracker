package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Feedback struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TicketID       uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"ticket_id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User           User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Rating         int       `gorm:"type:smallint;not null" json:"rating"`
	Comment        string    `gorm:"type:text" json:"comment"`
	TechEvaluation string    `gorm:"type:text" json:"tech_evaluation"`
	CreatedAt      time.Time `json:"created_at"`
}

func (f *Feedback) BeforeCreate(tx *gorm.DB) (err error) {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return
}
