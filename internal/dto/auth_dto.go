package dto

type RegisterRequest struct {
	FullName    string `json:"full_name" binding:"required,min=2,max=100"`
	Email       string `json:"email" binding:"required,email"`
	PhoneNumber string `json:"phone_number" binding:"required,min=9,max=20"`
	RoomNo      string `json:"room_no" binding:"required,min=1,max=20"`
	Password    string `json:"password" binding:"required,min=6,max=100"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"` // Email or RoomNo
	Password   string `json:"password" binding:"required"`
}

type AuthResponse struct {
	AccessToken string      `json:"access_token"`
	User        UserProfile `json:"user"`
}

type UserProfile struct {
	ID          string `json:"id"`
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	RoomNo      string `json:"room_no"`
	Role        string `json:"role"`
}
