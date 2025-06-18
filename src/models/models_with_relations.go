package models

import "time"

type Role struct {
	ID   uint64 `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"type:varchar(50);not null;uniqueIndex"`

	Users []User `gorm:"foreignKey:Role;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;references:Name"`
}

func (Role) TableName() string {
	return "role"
}

// User represents users table
type User struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"type:varchar(255);not null" json:"password"`
	Avatar    string    `gorm:"type:varchar(255);default:default.jpg" json:"avatar"`
	Role      string    `gorm:"type:varchar(50);default:'pegawai';not null" json:"role"`
	Status    string    `gorm:"type:varchar(50);default:'offline';not null" json:"status"`
	OTP       string    `gorm:"column:OTP;type:varchar(100);default:null" json:"otp"`
	Token     string    `gorm:"type:varchar(255);default:null" json:"token"`
	OTPActive bool      `gorm:"column:OTP_Active;default:false" json:"otp_active"`
	CreatedAt time.Time `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relations
	EmployeeShifts []EmployeeShift `gorm:"foreignKey:UserEmail;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;references:Email"`
	ShiftLogs      []ShiftLog      `gorm:"foreignKey:UserEmail;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;references:Email"`
	UserLogs       []UserLog       `gorm:"foreignKey:UserEmail;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;references:Email"`
	UserTickets    []UserTicket    `gorm:"foreignKey:UserEmail;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;references:Email"`
	ExportLogs     []ExportLog     `gorm:"foreignKey:UserEmail;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;references:Email"`
	Notes          []Note          `gorm:"foreignKey:UserEmail;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;references:Email"`
	Tickets        []Ticket        `gorm:"foreignKey:UserEmail;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;references:Email"`
}

func (User) TableName() string {
	return "users"
}

type Shift struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement:false"`
	ShiftName string    `gorm:"type:varchar(100);not null"`
	StartTime time.Time `gorm:"type:TIME;not null"`
	EndTime   time.Time `gorm:"type:TIME;not null"`
	CreatedAt time.Time `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP"`

	// Relations
	EmployeeShifts []EmployeeShift `gorm:"foreignKey:ShiftID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ShiftLogs      []ShiftLog      `gorm:"foreignKey:ShiftID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// TableName specifies the table name for Shift
func (Shift) TableName() string {
	return "shifts"
}

// EmployeeShift represents employee_shifts table
type EmployeeShift struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	UserEmail string    `gorm:"type:varchar(255);not null;index"`
	ShiftID   uint64    `gorm:"not null;index"`
	ShiftDate time.Time `gorm:"type:date;not null"`
	CreatedAt time.Time `gorm:"type:TIMESTAMP"`

	User  User  `gorm:"foreignKey:UserEmail;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;references:Email"`
	Shift Shift `gorm:"foreignKey:ShiftID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// TableName specifies the table name for EmployeeShift
func (EmployeeShift) TableName() string {
	return "employee_shifts"
}

// ShiftLog represents shift_logs table
type ShiftLog struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	UserEmail string    `gorm:"type:varchar(255);not null;index"`
	ShiftID   uint64    `gorm:"not null;index"`
	ShiftDate time.Time `gorm:"type:date;not null"`
	Reason    string    `gorm:"type:text"`

	User  User  `gorm:"foreignKey:UserEmail;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;references:Email"`
	Shift Shift `gorm:"foreignKey:ShiftID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// TableName specifies the table name for ShiftLog
func (ShiftLog) TableName() string {
	return "shift_logs"
}

// Category represents category table
type Category struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement"`
	CategoryName string `gorm:"type:varchar(100);not null;"`
	ProductsID   uint64 `gorm:"not null;index"`

	Product Product `gorm:"foreignKey:ProductsID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// TableName specifies the table name for Category
func (Category) TableName() string {
	return "category"
}

// Product represents products table
type Product struct {
	ID   uint64 `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"type:varchar(100);not null;uniqueIndex"`

	Categories []Category `gorm:"foreignKey:ProductsID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// TableName specifies the table name for Product
func (Product) TableName() string {
	return "products"
}

// Ticket represents tickets table
type Ticket struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement"`
	TrackingID      string     `gorm:"type:varchar(100);uniqueIndex;not null"`
	HariMasuk       time.Time  `gorm:"type:date;not null"`
	WaktuMasuk      string     `gorm:"type:TIME;not null"`
	HariRespon      *time.Time `gorm:"type:date;default:null"`
	WaktuRespon     string     `gorm:"type:TIME;default:null"`
	SolvedTime      string     `gorm:"type:varchar(100);default:null"`
	UserName        string     `gorm:"type:varchar(255);not null"`
	UserEmail       string     `gorm:"type:varchar(255);not null;index"`
	NoWhatsapp      string     `gorm:"type:varchar(20);default:null"`
	CategoryId      uint64     `gorm:"not null; index"`
	ProductsName    string     `gorm:"type:varchar(255);not null;index"`
	Priority        string     `gorm:"type:enum('Low','Medium','High','Critical');default:'Low';not null"`
	PlacesID        *uint64    `gorm:"index"`
	Status          string     `gorm:"type:enum('New','On Progress','Resolved');default:'New';not null"`
	Subject         string     `gorm:"type:varchar(255);not null"`
	DetailKendala   string     `gorm:"type:text;not null"`
	PIC             string     `gorm:"column:PIC;type:varchar(255);default:null"`
	ResponDiberikan string     `gorm:"type:text;default:null"`
	CreatedAt       *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt       *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`

	Category Category `gorm:"foreignKey:CategoryId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	User     User     `gorm:"foreignKey:UserEmail;references:Email;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Product  Product  `gorm:"foreignKey:ProductsName;references:Name;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Place    *Place   `gorm:"foreignKey:PlacesID;refrences:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

func (Ticket) TableName() string {
	return "tickets"
}

type TempTickets struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement"`
	TrackingID      string     `gorm:"type:varchar(100);uniqueIndex;"`
	HariMasuk       time.Time  `gorm:"type:date;not null"`
	WaktuMasuk      string     `gorm:"type:TIME;not null"`
	HariRespon      *time.Time `gorm:"type:date;default:null"`
	WaktuRespon     string     `gorm:"type:TIME;default:null"`
	SolvedTime      string     `gorm:"type:varchar(100);default:null"`
	UserEmail       string     `gorm:"type:varchar(255);not null;index"`
	NoWhatsapp      string     `gorm:"type:varchar(20);default:null"`
	CategoryId      uint64     `gorm:"not null; index"`
	ProductsName    string     `gorm:"type:varchar(255);not null;index"`
	Priority        string     `gorm:"type:enum('Low','Medium','High','Critical');default:'Low';not null"`
	Status          string     `gorm:"type:enum('New','On Progress','Resolved');default:'New';not null"`
	Subject         string     `gorm:"type:varchar(255);not null"`
	DetailKendala   string     `gorm:"type:text;not null"`
	PIC             string     `gorm:"column:PIC;type:varchar(255);default:null"`
	ResponDiberikan string     `gorm:"type:text;default:null"`
	DeletedBy       string     `gorm:"type:varchar(255); not null;"`
	CreatedAt       *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt       *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	DeletedAt       *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`

	User     User     `gorm:"foreignKey:UserEmail;references:Email;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Category Category `gorm:"foreignKey:CategoryId;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Product  Product  `gorm:"foreignKey:ProductsName;references:Name;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (TempTickets) TableName() string {
	return "temp_tickets"
}

type TempUserTickets struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	TicketsID string    `gorm:"type:varchar(100);not null;index"`
	UserEmail string    `gorm:"type:varchar(255);not null;index"`
	NewStatus string    `gorm:"type:varchar(50);not null"`
	UpdateAt  time.Time `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP"`
	Priority  string    `gorm:"type:enum('Low','Medium','High','Critical');default:null"`
	Details   string    `gorm:"type:varchar(150);default:null"`

	User       User        `gorm:"foreignKey:UserEmail;references:Email;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	TempTicket TempTickets `gorm:"foreignKey:TicketsID;references:TrackingID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// TableName specifies the table name for TempUserTickets
func (TempUserTickets) TableName() string {
	return "temp_user_tickets"
}

// UserLog represents user_logs table
type UserLog struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement"`
	UserEmail  string     `gorm:"type:varchar(255);not null;index"`
	LoginTime  time.Time  `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP"`
	LogoutTime *time.Time `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP"`
	ShiftName  string     `gorm:"type:varchar(100);default:null"`
	OTP        string     `gorm:"type:varchar(100);default:null"`

	User User `gorm:"foreignKey:UserEmail;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;references:Email"`
}

// TableName specifies the table name for UserLog
func (UserLog) TableName() string {
	return "user_logs"
}

// UserTicket represents user_tickets table
type UserTicket struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	TicketsID string    `gorm:"type:varchar(100);not null;index"`
	UserEmail string    `gorm:"type:varchar(255);not null;index"`
	NewStatus string    `gorm:"type:varchar(50);not null"`
	UpdateAt  time.Time `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP"`
	Priority  string    `gorm:"type:enum('Low','Medium','High','Critical');default:null"`
	Details   string    `gorm:"type:varchar(150);default:null"`

	User   User   `gorm:"foreignKey:UserEmail;references:Email;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Ticket Ticket `gorm:"foreignKey:TicketsID;references:TrackingID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// TableName specifies the table name for UserTicket
func (UserTicket) TableName() string {
	return "user_tickets"
}

type Place struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement"`
	Name       string `gorm:"type:varchar(100); index"`
	ProductsID uint64 `gorm:"index"`

	Product Product `gorm:"foreignKey:ProductsID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

func (Place) TableName() string {
	return "places"
}

// ExportLog represents export_logs table
type ExportLog struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"`
	FileName    string    `gorm:"type:varchar(255);not null"`
	UserEmail   string    `gorm:"type:varchar(255);not null;index"`
	CreatedAt   time.Time `gorm:"type:TIMESTAMP"`
	HistoryType string    `gorm:"type:varchar(100);not null"`

	User User `gorm:"foreignKey:UserEmail;references:Email"`
}

// TableName specifies the table name for ExportLog
func (ExportLog) TableName() string {
	return "export_logs"
}

// Note represents note table
type Note struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement"`
	Title     string     `gorm:"type:varchar(255);not null"`
	Content   string     `gorm:"type:text;not null"`
	UserEmail string     `gorm:"type:varchar(255);not null;index"`
	CreatedAt *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`

	User User `gorm:"foreignKey:UserEmail;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;references:Email"`
}

// TableName specifies the table name for Note
func (Note) TableName() string {
	return "notes"
}
