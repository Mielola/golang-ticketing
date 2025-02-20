package types

import "time"

// --------------------------------------------
// @ Ticket Type
// --------------------------------------------
type Tickets struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	TrackingID      string     `json:"tracking_id"`
	HariMasuk       string     `json:"hari_masuk"`
	WaktuMasuk      string     `json:"waktu_masuk"`
	HariRespon      string     `json:"hari_respon,omitempty"`
	WaktuRespon     string     `json:"waktu_respon,omitempty"`
	UserName        string     `json:"user_name,omitempty"`
	UserEmail       string     `json:"user_email"`
	Category        string     `json:"category"`
	Priority        string     `json:"priority"`
	Status          string     `json:"status"`
	Subject         string     `json:"subject"`
	DetailKendala   string     `json:"detail_kendala"`
	Owner           string     `json:"owner"`
	TimeWorked      *int       `json:"time_worked,omitempty"`
	DueDate         *time.Time `json:"due_date,omitempty"`
	ResponDiberikan string     `json:"respon_diberikan,omitempty"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// --------------------------------------------
// @ User Type
// --------------------------------------------

type User struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Password       string    `json:"password"`
	PasswordRetype string    `json:"password_retype" gorm:"-"`
	Status         *string   `json:"status"`
	OTP            *string   `json:"OTP"`
	UpdatedAt      time.Time `json:"updated_at"`
	CreatedAt      time.Time `json:"created_at"`
	Role           string    `json:"role"`
	Token          string    `json:"token"`
}
type UserPost struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Password       string    `json:"password"`
	PasswordRetype string    `json:"password_retype" gorm:"-"`
	Status         *string   `json:"status"`
	OTP            *string   `json:"OTP"`
	UpdatedAt      time.Time `json:"updated_at"`
	CreatedAt      time.Time `json:"created_at"`
}
type UserResponse struct {
	ID        uint    `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	ShiftName *string `json:"shift_name"`
	Avatar    string  `json:"avatar"`
	Status    string  `json:"status"`
	Token     string  `json:"token"`
}

type UserResponseWithoutToken struct {
	ID        uint    `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	ShiftName *string `json:"shift_name"`
	Avatar    *string `json:"avatar"`
	Status    string  `json:"status"`
}

type UserBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// --------------------------------------------
// @ Note Type
// --------------------------------------------

type NoteBody struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	UserEmail string `json:"user_email"`
}

type Note struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	UserEmail string `json:"user_email"`
}

type NoteDetail struct {
	ID      uint   `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type NoteResponse struct {
	Email string       `json:"user_email"`
	Name  string       `json:"name"`
	Notes []NoteDetail `json:"notes"`
}

// --------------------------------------------
// @ Shift Type
// --------------------------------------------

type Shift struct {
	ID        uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	ShiftName string `json:"shift_name" gorm:"type:varchar(100);not null"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type EmployeeShift struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	UserEmail string `json:"user_email"`
	ShiftID   uint   `json:"shift_id"`
	ShiftDate string `json:"shift_date"`
}

type ShiftRequest struct {
	UserEmail string `json:"user_email" binding:"required"`
	ShiftID   string `json:"shift_id"`
	ShiftDate string `json:"shift_date"`
	Reason    string `json:"reason"`
}

type ShiftResponse struct {
	ID        uint   `json:"id"`
	UserEmail string `json:"user_email"`
	ShiftName string `json:"shift_name"`
	ShiftDate string `json:"shift_date"`
}

type ShiftLogs struct {
	ID        uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	UserEmail string `json:"user_email"`
	ShiftID   uint   `json:"shift_id"`
	ShiftDate string `json:"shift_date"`
	Reason    string `json:"reason"`
}

// --------------------------------------------
// @ Login Type
// --------------------------------------------
type LoginTime struct {
	Email string    `json:"email"`
	Login time.Time `json:"Login"`
}
