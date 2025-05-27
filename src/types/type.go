package types

import "time"

// --------------------------------------------
// @ Ticket Type
// --------------------------------------------
type Tickets struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	TrackingID      string     `json:"tracking_id"`
	HariMasuk       time.Time  `json:"hari_masuk"`
	WaktuMasuk      string     `json:"waktu_masuk"`
	HariRespon      string     `json:"hari_respon,omitempty"`
	WaktuRespon     string     `json:"waktu_respon,omitempty"`
	UserName        string     `json:"user_name,omitempty"`
	UserEmail       string     `json:"user_email"`
	CategoryName    string     `json:"category_name"`
	Priority        string     `json:"priority"`
	Status          string     `json:"status"`
	Subject         string     `json:"subject"`
	NoWhatsapp      string     `json:"no_whatsapp"`
	DetailKendala   string     `json:"detail_kendala"`
	TimeWorked      *int       `json:"time_worked,omitempty"`
	DueDate         *time.Time `json:"due_date,omitempty"`
	ResponDiberikan string     `json:"respon_diberikan,omitempty"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	PIC             string     `json:"PIC"`
	SolvedTime      *string    `json:"solved_time,omitempty"`
	UserAvatar      string     `json:"user_avatar,omitempty"`
}

type TicketsInput struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	TrackingID      string     `json:"tracking_id"`
	HariMasuk       time.Time  `json:"hari_masuk"`
	WaktuMasuk      string     `json:"waktu_masuk"`
	HariRespon      string     `json:"hari_respon,omitempty"`
	WaktuRespon     string     `json:"waktu_respon,omitempty"`
	UserName        string     `json:"user_name,omitempty"`
	ProductsName    string     `json:"products_name"`
	UserEmail       string     `json:"user_email"`
	CategoryName    string     `json:"category_name"`
	Priority        string     `json:"priority"`
	Subject         string     `json:"subject"`
	NoWhatsapp      string     `json:"no_whatsapp" binding:"required"`
	DetailKendala   string     `json:"detail_kendala"`
	ResponDiberikan string     `json:"respon_diberikan,omitempty"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       *time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	PIC             string     `json:"PIC" binding:"required"`
}

type TicketsResponseAll struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	TrackingID      string    `json:"tracking_id"`
	ProductsName    string    `json:"products_name"`
	HariMasuk       time.Time `json:"hari_masuk"`
	WaktuMasuk      string    `json:"waktu_masuk"`
	HariRespon      time.Time `json:"hari_respon,omitempty"`
	WaktuRespon     string    `json:"waktu_respon,omitempty"`
	UserName        string    `json:"user_name,omitempty"`
	UserEmail       string    `json:"user_email"`
	NoWhatsapp      string    `json:"no_whatsapp"`
	CategoryName    string    `json:"category_name"`
	Priority        string    `json:"priority"`
	Status          string    `json:"status"`
	Subject         string    `json:"subject"`
	DetailKendala   string    `json:"detail_kendala"`
	PIC             string    `json:"PIC"`
	ResponDiberikan string    `json:"respon_diberikan,omitempty"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	SolvedTime      *string   `json:"solved_time,omitempty"`
}

type TicketsLogsRaw struct {
	ID         uint       `json:"id"`
	TicketsId  string     `json:"tickets_id"`
	Priority   string     `json:"priority"`
	NewStatus  string     `json:"new_status"`
	UpdateAt   *time.Time `json:"update_at"`
	Details    string     `json:"details"`
	UserEmail  string     `json:"user_email"`
	UserName   string     `json:"user_name"`
	UserAvatar string     `json:"user_avatar"`
}

type TicketsCreator struct {
	Email  string `json:"email"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type TicketsResponse struct {
	OpenTickets     int `json:"open_tickets"`
	PendingTickets  int `json:"pending_tickets"`
	ResolvedTickets int `json:"resolved_tickets"`
	TotalTickets    int `json:"total_tickets"`
	CriticalTickets int `json:"critical_tickets"`
}

type TicketsLogs struct {
	ID             uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	TicketsId      string         `json:"tickets_id"`
	Details        string         `json:"details"`
	Priority       string         `json:"priority"`
	UserEmail      string         `json:"-"`
	UserName       string         `json:"-"`
	UserAvatar     string         `json:"-"`
	NewStatus      string         `json:"new_status"`
	CurrentStatus  string         `json:"current_status"`
	UpdateAt       *time.Time     `json:"-"`
	UpdateAtString string         `json:"update_at"`
	User           TicketsCreator `json:"user"`
}

type UpdateTicketInput struct {
	ProductsName    string    `json:"products_name"`
	CategoryName    string    `json:"category_name"`
	NoWhatsapp      string    `json:"no_whatsapp"`
	PIC             string    `json:"PIC"`
	DetailKendala   string    `json:"detail_kendala"`
	Priority        string    `json:"priority"`
	Status          string    `json:"status"`
	HariMasuk       time.Time `json:"hari_masuk"`
	WaktuMasuk      string    `json:"waktu_masuk"`
	ResponDiberikan string    `json:"respon_diberikan"`
}
type UserTicketHistory struct {
	UserEmail string `json:"user_email"`
	NewStatus string `json:"new_status"`
	TicketsID string `json:"ticket_id"`
	Priority  string `json:"priority"`
	Details   string `json:"details"`
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
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Status    *string   `json:"status"`
	OTP       *string   `json:"OTP"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}
type UserResponse struct {
	ID        uint    `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	ShiftName *string `json:"shift_name"`
	Avatar    *string `json:"avatar"`
	Status    string  `json:"status"`
	Token     string  `json:"token"`
}

type UserResponseWithoutToken struct {
	ID          uint    `json:"id"`
	Email       string  `json:"email"`
	Name        string  `json:"name"`
	Role        string  `json:"role"`
	ShiftName   *string `json:"shift_name"`
	Avatar      *string `json:"avatar"`
	Status      string  `json:"status"`
	ShiftStatus *string `json:"shift_status"`
}
type UserResponseWithoutRole struct {
	ID        uint    `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
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
	ShiftId   uint   `json:"shift_id"`
	UserEmail string `json:"user_email"`
	Name      string `json:"name"`
	ShiftName string `json:"shift_name"`
	ShiftDate string `json:"shift_date"`
	StartTime string `jsong:"start_time"`
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

// --------------------------------------------
// @ Dashboard Type
// --------------------------------------------
type DashboardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    DataContent `json:"data"`
}

type DataContent struct {
	Summary       TicketsResponse          `json:"summary"`
	RecentTickets []map[string]interface{} `json:"recent_tickets"`
	UserLogs      interface{}              `json:"user_logs"`
}

type UserLogResponse struct {
	UserResponseWithoutToken
	LoginDate time.Time `json:"login_date"`
	LoginTime time.Time `json:"login_time"`
}

// --------------------------------------------
// @ Dashboard Type
// --------------------------------------------
type ResponseFormat struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
