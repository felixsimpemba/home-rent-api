package models

import (
	"time"

	"gorm.io/gorm"
)

// ─── Core Users ───────────────────────────────────────────────────────────────

// User represents a registered account
type User struct {
	ID             string         `gorm:"primaryKey;size:50"`
	FirstName      string         `gorm:"size:100;not null"`
	LastName       string         `gorm:"size:100;not null"`
	Email          string         `gorm:"size:150;uniqueIndex;not null"`
	Phone          string         `gorm:"size:50;not null"`
	Password       string         `gorm:"size:255;not null" json:"-"` // never expose in responses
	Role           string         `gorm:"size:50;default:tenant;not null"`
	AvatarURL      string         `gorm:"size:500"`
	EmailVerified  bool           `gorm:"default:false;not null"`
	Banned         bool           `gorm:"default:false;not null"`
	MFAEnabled     bool           `gorm:"default:false;not null"`
	MFASecret      string         `gorm:"size:100"            json:"-"` // never expose
	Preferences    string         `gorm:"type:text"           json:"-"` // returned via dedicated endpoint
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"               json:"-"`
}

// Session represents an active login session (JWT refresh token)
type Session struct {
	ID           string         `gorm:"primaryKey;size:50"`
	UserID       string         `gorm:"size:50;index;not null"`
	User         User           `gorm:"foreignKey:UserID"`
	RefreshToken string         `gorm:"size:512;uniqueIndex;not null"`
	DeviceName   string         `gorm:"size:200"`
	IPAddress    string         `gorm:"size:50"`
	ExpiresAt    time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// PasswordResetToken stores one-time password reset tokens
type PasswordResetToken struct {
	ID        string    `gorm:"primaryKey;size:50"`
	UserID    string    `gorm:"size:50;index;not null"`
	Token     string    `gorm:"size:200;uniqueIndex;not null"`
	UsedAt    *time.Time
	ExpiresAt time.Time
	CreatedAt time.Time
}

// EmailVerificationToken stores email verification codes
type EmailVerificationToken struct {
	ID        string    `gorm:"primaryKey;size:50"`
	UserID    string    `gorm:"size:50;index;not null"`
	Token     string    `gorm:"size:200;uniqueIndex;not null"`
	UsedAt    *time.Time
	ExpiresAt time.Time
	CreatedAt time.Time
}

// ─── Properties ───────────────────────────────────────────────────────────────

// Category for property type groupings
type Category struct {
	ID        string         `gorm:"primaryKey;size:50"`
	Name      string         `gorm:"size:100;uniqueIndex;not null"`
	Slug      string         `gorm:"size:100;uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Property represents a real estate listing
type Property struct {
	ID            string         `gorm:"primaryKey;size:50"`
	Title         string         `gorm:"size:255;not null"`
	Description   string         `gorm:"type:text;not null"`
	PropertyType  string         `gorm:"size:50;not null"` // apartment|house|studio|office|land
	ListingType   string         `gorm:"size:50;not null"` // rent|sale
	Price         float64        `gorm:"type:decimal(12,2);not null"`
	Deposit       float64        `gorm:"type:decimal(12,2)"`
	Bedrooms      int            `gorm:"default:0"`
	Bathrooms     int            `gorm:"default:0"`
	ParkingSpaces int            `gorm:"default:0"`
	Province      string         `gorm:"size:100;not null"`
	District      string         `gorm:"size:100;not null"`
	Area          string         `gorm:"size:100;not null"`
	Address       string         `gorm:"size:300"`
	Latitude      float64        `gorm:"type:decimal(10,8)"`
	Longitude     float64        `gorm:"type:decimal(11,8)"`
	Status        string         `gorm:"size:50;default:pending;not null"` // pending|active|rented|archived|rejected
	Featured      bool           `gorm:"default:false;not null"`
	OwnerID       string         `gorm:"size:50;not null"`
	Owner         User           `gorm:"foreignKey:OwnerID"`
	Images        []PropertyImage `gorm:"foreignKey:PropertyID"`
	Amenities     []Amenity      `gorm:"foreignKey:PropertyID"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

// PropertyImage stores images attached to a listing
type PropertyImage struct {
	ID         string    `gorm:"primaryKey;size:50"`
	PropertyID string    `gorm:"size:50;index;not null"`
	URL        string    `gorm:"size:500;not null"`
	Caption    string    `gorm:"size:200"`
	SortOrder  int       `gorm:"default:0"`
	CreatedAt  time.Time
}

// Amenity stores features of a property (water, electricity, etc.)
type Amenity struct {
	ID         string `gorm:"primaryKey;size:50"`
	PropertyID string `gorm:"size:50;index;not null"`
	Name       string `gorm:"size:100;not null"`
}

// AvailabilityBlock stores viewing time blocks for a property
type AvailabilityBlock struct {
	ID         string    `gorm:"primaryKey;size:50"`
	PropertyID string    `gorm:"size:50;index;not null"`
	Date       string    `gorm:"size:20;not null"` // YYYY-MM-DD
	StartTime  string    `gorm:"size:10;not null"` // HH:MM
	EndTime    string    `gorm:"size:10;not null"` // HH:MM
	CreatedAt  time.Time
}

// PropertyHistoryLog stores audit changes on a property (price changes etc.)
type PropertyHistoryLog struct {
	ID         string    `gorm:"primaryKey;size:50"`
	PropertyID string    `gorm:"size:50;index;not null"`
	Event      string    `gorm:"size:100;not null"` // e.g. price_change, status_change
	OldValue   string    `gorm:"size:300"`
	NewValue   string    `gorm:"size:300"`
	ChangedBy  string    `gorm:"size:50"` // userID
	CreatedAt  time.Time
}

// Report is a user-submitted flag on a listing
type Report struct {
	ID         string         `gorm:"primaryKey;size:50"`
	PropertyID string         `gorm:"size:50;index;not null"`
	ReporterID string         `gorm:"size:50;not null"`
	Reason     string         `gorm:"size:300;not null"`
	Status     string         `gorm:"size:50;default:open;not null"` // open|resolved
	ResolvedBy string         `gorm:"size:50"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

// ─── Tenant Activity ──────────────────────────────────────────────────────────

// Favorite represents a tenant's saved property
type Favorite struct {
	ID         string    `gorm:"primaryKey;size:50"`
	UserID     string    `gorm:"size:50;index;not null"`
	PropertyID string    `gorm:"size:50;index;not null"`
	CreatedAt  time.Time
}

// SavedSearch stores a saved search query for a tenant
type SavedSearch struct {
	ID        string    `gorm:"primaryKey;size:50"`
	UserID    string    `gorm:"size:50;index;not null"`
	Name      string    `gorm:"size:100;not null"`
	Filters   string    `gorm:"type:text;not null"` // JSON blob
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ─── Inquiries ────────────────────────────────────────────────────────────────

// Inquiry represents a tenant's message about a property
type Inquiry struct {
	ID         string         `gorm:"primaryKey;size:50"`
	PropertyID string         `gorm:"size:50;index;not null"`
	Property   Property       `gorm:"foreignKey:PropertyID"`
	TenantID   string         `gorm:"size:50;not null"`
	Tenant     User           `gorm:"foreignKey:TenantID"`
	Message    string         `gorm:"type:text;not null"`
	Status     string         `gorm:"size:50;default:open;not null"` // open|closed
	Replies    []InquiryReply `gorm:"foreignKey:InquiryID"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

// InquiryReply is a reply to an inquiry thread
type InquiryReply struct {
	ID        string    `gorm:"primaryKey;size:50"`
	InquiryID string    `gorm:"size:50;index;not null"`
	SenderID  string    `gorm:"size:50;not null"`
	Sender    User      `gorm:"foreignKey:SenderID"`
	Message   string    `gorm:"type:text;not null"`
	CreatedAt time.Time
}

// Viewing represents a scheduled property viewing
type Viewing struct {
	ID          string         `gorm:"primaryKey;size:50"`
	PropertyID  string         `gorm:"size:50;index;not null"`
	Property    Property       `gorm:"foreignKey:PropertyID"`
	TenantID    string         `gorm:"size:50;not null"`
	Tenant      User           `gorm:"foreignKey:TenantID"`
	Date        string         `gorm:"size:20;not null"` // YYYY-MM-DD
	StartTime   string         `gorm:"size:10;not null"`
	Status      string         `gorm:"size:50;default:pending;not null"` // pending|approved|rejected|cancelled
	Notes       string         `gorm:"type:text"`
	Feedback    string         `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// Booking represents a confirmed stay/rental agreement intent
type Booking struct {
	ID         string         `gorm:"primaryKey;size:50"`
	PropertyID string         `gorm:"size:50;index;not null"`
	Property   Property       `gorm:"foreignKey:PropertyID"`
	TenantID   string         `gorm:"size:50;not null"`
	Tenant     User           `gorm:"foreignKey:TenantID"`
	MoveInDate string         `gorm:"size:20"`
	Status     string         `gorm:"size:50;default:pending;not null"` // pending|approved|cancelled
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

// Lease represents a formal rental agreement
type Lease struct {
	ID          string         `gorm:"primaryKey;size:50"`
	BookingID   string         `gorm:"size:50;index;not null"`
	PropertyID  string         `gorm:"size:50;not null"`
	Property    Property       `gorm:"foreignKey:PropertyID"`
	TenantID    string         `gorm:"size:50;not null"`
	Tenant      User           `gorm:"foreignKey:TenantID"`
	LandlordID  string         `gorm:"size:50;not null"`
	StartDate   string         `gorm:"size:20;not null"`
	EndDate     string         `gorm:"size:20;not null"`
	MonthlyRent float64        `gorm:"type:decimal(12,2);not null"`
	Status      string         `gorm:"size:50;default:draft;not null"` // draft|signed|terminated
	TenantSigned  bool         `gorm:"default:false"`
	LandlordSigned bool        `gorm:"default:false"`
	TerminatedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TenantScreening holds results of a background check
type TenantScreening struct {
	ID          string         `gorm:"primaryKey;size:50"`
	TenantID    string         `gorm:"size:50;index;not null"`
	Tenant      User           `gorm:"foreignKey:TenantID"`
	PropertyID  string         `gorm:"size:50;not null"`
	Status      string         `gorm:"size:50;default:pending;not null"` // pending|approved|rejected
	Score       int
	Notes       string         `gorm:"type:text"`
	CreatedBy   string         `gorm:"size:50;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// ─── Payments ─────────────────────────────────────────────────────────────────

// Transaction represents a mobile money payment
type Transaction struct {
	ID          string         `gorm:"primaryKey;size:50"`
	Reference   string         `gorm:"size:100;uniqueIndex;not null"`
	UserID      string         `gorm:"size:50;index;not null"`
	Provider    string         `gorm:"size:50;not null"` // mtn|airtel|zamtel
	PhoneNumber string         `gorm:"size:50;not null"`
	Amount      float64        `gorm:"type:decimal(12,2);not null"`
	Currency    string         `gorm:"size:10;default:ZMW;not null"`
	Purpose     string         `gorm:"size:100"` // rent|deposit|subscription
	Status      string         `gorm:"size:50;default:pending;not null"` // pending|successful|failed
	ProviderRef string         `gorm:"size:200"` // external reference from provider
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// Refund represents a requested refund for a transaction
type Refund struct {
	ID            string         `gorm:"primaryKey;size:50"`
	TransactionID string         `gorm:"size:50;index;not null"`
	Transaction   Transaction    `gorm:"foreignKey:TransactionID"`
	UserID        string         `gorm:"size:50;not null"`
	Reason        string         `gorm:"type:text;not null"`
	Amount        float64        `gorm:"type:decimal(12,2);not null"`
	Status        string         `gorm:"size:50;default:pending;not null"` // pending|approved|rejected
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// SubscriptionPlan defines available plans
type SubscriptionPlan struct {
	ID          string    `gorm:"primaryKey;size:50"`
	Name        string    `gorm:"size:100;not null"`
	Description string    `gorm:"type:text"`
	PriceMonthly float64  `gorm:"type:decimal(12,2);not null"`
	Features    string    `gorm:"type:text"` // JSON array
	Active      bool      `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Subscription links a user to a plan
type Subscription struct {
	ID        string         `gorm:"primaryKey;size:50"`
	UserID    string         `gorm:"size:50;index;not null"`
	User      User           `gorm:"foreignKey:UserID"`
	PlanID    string         `gorm:"size:50;not null"`
	Plan      SubscriptionPlan `gorm:"foreignKey:PlanID"`
	Status    string         `gorm:"size:50;default:active;not null"` // active|cancelled|expired
	StartedAt time.Time
	EndsAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ─── Messaging ────────────────────────────────────────────────────────────────

// Conversation between two users
type Conversation struct {
	ID           string    `gorm:"primaryKey;size:50"`
	Participant1 string    `gorm:"size:50;index;not null"`
	Participant2 string    `gorm:"size:50;index;not null"`
	PropertyID   string    `gorm:"size:50"`
	LastMessage  string    `gorm:"type:text"`
	LastAt       *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Message represents a single chat message
type Message struct {
	ID             string         `gorm:"primaryKey;size:50"`
	ConversationID string         `gorm:"size:50;index;not null"`
	Conversation   Conversation   `gorm:"foreignKey:ConversationID"`
	SenderID       string         `gorm:"size:50;not null"`
	Sender         User           `gorm:"foreignKey:SenderID"`
	Content        string         `gorm:"type:text;not null"`
	CreatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

// Notification represents an in-app alert
type Notification struct {
	ID          string         `gorm:"primaryKey;size:50"`
	RecipientID string         `gorm:"size:50;index;not null"`
	Title       string         `gorm:"size:255;not null"`
	Body        string         `gorm:"type:text;not null"`
	Type        string         `gorm:"size:50"` // booking|inquiry|payment|system
	Read        bool           `gorm:"default:false;not null"`
	CreatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// ─── File Uploads ─────────────────────────────────────────────────────────────

// UploadedFile tracks files uploaded by users
type UploadedFile struct {
	ID        string         `gorm:"primaryKey;size:50"`
	UploaderID string        `gorm:"size:50;index;not null"`
	Filename  string         `gorm:"size:300;not null"`
	URL       string         `gorm:"size:500;not null"`
	MimeType  string         `gorm:"size:100"`
	SizeBytes int64
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// ─── CRM ──────────────────────────────────────────────────────────────────────

// Lead represents a sales lead for an agent
type Lead struct {
	ID          string         `gorm:"primaryKey;size:50"`
	AgentID     string         `gorm:"size:50;index;not null"`
	Name        string         `gorm:"size:200;not null"`
	Email       string         `gorm:"size:150"`
	Phone       string         `gorm:"size:50"`
	Source      string         `gorm:"size:100"` // website|referral|social
	Status      string         `gorm:"size:50;default:new;not null"` // new|contacted|qualified|lost|converted
	Notes       string         `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// LeadNote is a note attached to a lead
type LeadNote struct {
	ID        string    `gorm:"primaryKey;size:50"`
	LeadID    string    `gorm:"size:50;index;not null"`
	AgentID   string    `gorm:"size:50;not null"`
	Content   string    `gorm:"type:text;not null"`
	CreatedAt time.Time
}

// Task is an agent's to-do item
type Task struct {
	ID          string         `gorm:"primaryKey;size:50"`
	AgentID     string         `gorm:"size:50;index;not null"`
	Title       string         `gorm:"size:255;not null"`
	Description string         `gorm:"type:text"`
	DueDate     *time.Time
	Completed   bool           `gorm:"default:false"`
	Priority    string         `gorm:"size:20;default:medium"` // low|medium|high
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// Meeting represents a scheduled meeting for an agent
type Meeting struct {
	ID          string    `gorm:"primaryKey;size:50"`
	AgentID     string    `gorm:"size:50;index;not null"`
	Title       string    `gorm:"size:255;not null"`
	Description string    `gorm:"type:text"`
	Location    string    `gorm:"size:300"`
	StartAt     time.Time
	EndAt       time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ─── Admin ────────────────────────────────────────────────────────────────────

// AuditLog records significant system actions
type AuditLog struct {
	ID         string    `gorm:"primaryKey;size:50"`
	ActorID    string    `gorm:"size:50;index"`
	Action     string    `gorm:"size:100;not null"`
	Resource   string    `gorm:"size:100"` // user|property|transaction
	ResourceID string    `gorm:"size:50"`
	Details    string    `gorm:"type:text"`
	IPAddress  string    `gorm:"size:50"`
	CreatedAt  time.Time
}

// SystemSetting stores key-value config managed by admin
type SystemSetting struct {
	Key       string `gorm:"primaryKey;size:100"`
	Value     string `gorm:"type:text;not null"`
	UpdatedAt time.Time
}
