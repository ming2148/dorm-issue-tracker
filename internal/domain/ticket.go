package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketStatus string
type TicketCategory string
type UrgencyLevel string

const (
	StatusPending    TicketStatus = "Pending"
	StatusInProgress TicketStatus = "In_Progress"
	StatusCompleted  TicketStatus = "Completed"
	StatusClosed     TicketStatus = "Closed"
	StatusCancelled  TicketStatus = "Cancelled"

	CategoryElectrical TicketCategory = "Electrical"
	CategoryPlumbing   TicketCategory = "Plumbing"
	CategoryAppliance  TicketCategory = "Appliance"
	CategoryStructural TicketCategory = "Structural"
	CategoryOther      TicketCategory = "Other"

	UrgencyNormal    UrgencyLevel = "Normal"
	UrgencyUrgent    UrgencyLevel = "Urgent"
	UrgencyEmergency UrgencyLevel = "Emergency"
)

type Ticket struct {
	ID                   uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TicketNo             string         `gorm:"type:varchar(30);uniqueIndex;not null" json:"ticket_no"`
	UserID               uuid.UUID      `gorm:"type:uuid;index;not null" json:"user_id"`
	User                 User           `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"user,omitempty"`
	RoomNo               string         `gorm:"type:varchar(20);index;not null" json:"room_no"`
	Title                string         `gorm:"type:varchar(150);not null" json:"title"`
	Category             TicketCategory `gorm:"type:varchar(30);index;not null" json:"category"`
	Description          string         `gorm:"type:text;not null" json:"description"`
	Urgency              UrgencyLevel   `gorm:"type:varchar(20);default:'Normal';not null" json:"urgency"`
	Status               TicketStatus   `gorm:"type:varchar(30);default:'Pending';index;not null" json:"status"`
	AssignedTechnicianID *uuid.UUID     `gorm:"type:uuid;index" json:"assigned_technician_id,omitempty"`
	AssignedTechnician   *User          `gorm:"foreignKey:AssignedTechnicianID" json:"assigned_technician,omitempty"`
	InternalNote         string         `gorm:"type:text" json:"internal_note,omitempty"`

	// Relationships
	Images   []TicketImage `gorm:"foreignKey:TicketID;constraint:OnDelete:CASCADE;" json:"images"`
	Feedback *Feedback     `gorm:"foreignKey:TicketID" json:"feedback,omitempty"`

	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (t *Ticket) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	t.TicketNo = fmt.Sprintf("TKT-%s-%s", time.Now().Format("20060102"), t.ID.String()[:6])
	return
}

type TicketImage struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TicketID  uuid.UUID `gorm:"type:uuid;index;not null" json:"ticket_id"`
	FileURL   string    `gorm:"type:varchar(500);not null" json:"file_url"`
	FileKey   string    `gorm:"type:varchar(255);not null" json:"-"`
	MimeType  string    `gorm:"type:varchar(50);not null" json:"mime_type"`
	FileSize  int64     `gorm:"not null" json:"file_size"`
	CreatedAt time.Time `json:"created_at"`
}

func (ti *TicketImage) BeforeCreate(tx *gorm.DB) (err error) {
	if ti.ID == uuid.Nil {
		ti.ID = uuid.New()
	}
	return
}
