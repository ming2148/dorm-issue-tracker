package database

import (
	"log"

	"dorm-issue-tracker/internal/domain"
	"dorm-issue-tracker/pkg/hasher"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedData(db *gorm.DB, h hasher.PasswordHasher) error {
	var count int64
	db.Model(&domain.User{}).Count(&count)
	if count > 0 {
		return nil // Already seeded
	}

	log.Println("Seeding initial demo data...")

	passResident, _ := h.HashPassword("password123")
	passAdmin, _ := h.HashPassword("admin123")
	passTech, _ := h.HashPassword("tech123")

	residentID := uuid.New()
	adminID := uuid.New()
	techID := uuid.New()

	users := []domain.User{
		{
			ID:           residentID,
			FullName:     "คุณสมชาย ใจดี (ลูกบ้าน)",
			Email:        "resident@condo.local",
			PhoneNumber:  "081-234-5678",
			RoomNo:       "402",
			PasswordHash: passResident,
			Role:         domain.RoleResident,
		},
		{
			ID:           adminID,
			FullName:     "นิติบุคคล อาคาร A (Admin)",
			Email:        "admin@condo.local",
			PhoneNumber:  "02-999-8888",
			RoomNo:       "OFFICE-01",
			PasswordHash: passAdmin,
			Role:         domain.RoleAdmin,
		},
		{
			ID:           techID,
			FullName:     "ช่างสมศักดิ์ บริการดี (Technician)",
			Email:        "tech@condo.local",
			PhoneNumber:  "089-876-5432",
			RoomNo:       "TECH-01",
			PasswordHash: passTech,
			Role:         domain.RoleTechnician,
		},
	}

	for _, u := range users {
		if err := db.Create(&u).Error; err != nil {
			return err
		}
	}

	// Seed Sample Tickets
	t1 := domain.Ticket{
		UserID:      residentID,
		RoomNo:      "402",
		Title:       "ท่อน้ำใต้อ่างล้างจานรั่วซึม มีน้ำหยดตลอดเวลา",
		Category:    domain.CategoryPlumbing,
		Description: "เปิดก๊อกน้ำแล้วพบว่ามีน้ำหยดรั่วซึมบริเวณข้อต่อใต้อ่างล้างจาน ไหลนองเต็มตู้ด้านล่าง รบกวนช่างตรวจสอบด่วนครับ",
		Urgency:     domain.UrgencyUrgent,
		Status:      domain.StatusPending,
	}

	t2 := domain.Ticket{
		UserID:               residentID,
		RoomNo:               "402",
		Title:                "แอร์ห้องนอนไม่เย็น มีแต่ลมและมีเสียงดังผิดปกติ",
		Category:             domain.CategoryAppliance,
		Description:          "เปิดแอร์ 23 องศาแล้วห้องไม่เย็น คอมเพรสเซอร์ทำงานแต่มีเสียงดังแกร๊กๆ เป็นระยะ",
		Urgency:              domain.UrgencyUrgent,
		Status:               domain.StatusInProgress,
		AssignedTechnicianID: &techID,
		InternalNote:         "เข้าตรวจเช็คแล้ว น้ำยาแอร์ปกติ แต่มอเตอร์พัดลมคอยล์เย็นหลวม นัดหมายลูกบ้านเปลี่ยนอะไหล่บ่ายนี้",
	}

	t3 := domain.Ticket{
		UserID:               residentID,
		RoomNo:               "402",
		Title:                "หลอดไฟระเบียงดับ ใช้งานไม่ได้",
		Category:             domain.CategoryElectrical,
		Description:          "หลอดไฟระเบียงด้านนอกเปิดสวิตช์แล้วไม่ติด หลอดน่าจะขาดครับ",
		Urgency:              domain.UrgencyNormal,
		Status:               domain.StatusCompleted,
		AssignedTechnicianID: &techID,
		InternalNote:         "เปลี่ยนหลอดไฟ LED E27 12W และเช็คขั้วไฟเรียบร้อย ลูกบ้านตรวจรับงานได้เลย",
	}

	t4 := domain.Ticket{
		UserID:               residentID,
		RoomNo:               "402",
		Title:                "กระเบื้องปูพื้นห้องน้ำบริเวณมุมห้องแตกมีคม",
		Category:             domain.CategoryStructural,
		Description:          "กระเบื้องมุมห้องน้ำมีรอยบิ่นแตก กลัวเดินเหยียบแล้วบาดเท้าครับ",
		Urgency:              domain.UrgencyNormal,
		Status:               domain.StatusClosed,
		AssignedTechnicianID: &techID,
		InternalNote:         "สกัดกระเบื้องแผ่นเดิมออกและปูแผ่นใหม่ ยาแนวเรียบร้อย",
	}

	db.Create(&t1)
	db.Create(&t2)
	db.Create(&t3)
	db.Create(&t4)

	// Feedback for t4
	fb := domain.Feedback{
		TicketID:       t4.ID,
		UserID:         residentID,
		Rating:         5,
		Comment:        "ช่างมาตรงเวลา สุภาพ และเก็บกวาดเศษฝุ่นปูนเรียบร้อยมากครับ ยอดเยี่ยม!",
		TechEvaluation: "ช่างสมศักดิ์มีความชำนาญสูง งานเรียบร้อย",
	}
	db.Create(&fb)

	log.Println("Demo seed data created successfully.")
	return nil
}
