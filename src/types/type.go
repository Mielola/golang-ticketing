package types

import "time"

// --------------------------------------------
// @ Ticket Model
// --------------------------------------------
type Tickets struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	TrackingID      string     `json:"tracking_id"`
	HariMasuk       string     `json:"hari_masuk"`
	WaktuMasuk      string     `json:"waktu_masuk"`
	HariRespon      string     `json:"hari_respon,omitempty"`
	WaktuRespon     string     `json:"waktu_respon,omitempty"`
	NamaAdmin       string     `json:"nama_admin,omitempty"`
	Email           string     `json:"email"`
	Category        string     `json:"category"`
	Priority        string     `json:"priority"`
	Status          string     `json:"status"`
	Subject         string     `json:"subject"`
	DetailKendala   string     `json:"detail_kendala"`
	Owner           string     `json:"owner"`
	TimeWorked      *int       `json:"time_worked,omitempty"` // Menggunakan pointer untuk mendukung null
	DueDate         *time.Time `json:"due_date,omitempty"`    // Menggunakan pointer untuk mendukung null
	KategoriMasalah string     `json:"kategori_masalah,omitempty"`
	ResponDiberikan string     `json:"respon_diberikan,omitempty"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// --------------------------------------------
// @ User Model
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
}
type UserResponse struct {
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	Role      string  `json:"role"`
	ShiftName *string `json:"shift_name"`
	OTP       *string `json:"OTP,omitempty"`
}

type UserBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// --------------------------------------------
// @ Note Model
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

type LoginTime struct {
	Email string       `json:"email"`
	Login  time.Time   `json:"Login"`
	Logout  *time.Time `json:"Logout"`
}