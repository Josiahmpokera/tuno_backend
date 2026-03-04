package domain

import (
	"time"
)

type RotationFrequency string

const (
	FrequencyEveryDay   RotationFrequency = "EVERY_DAY"
	FrequencyEvery2Days RotationFrequency = "EVERY_2_DAYS"
	FrequencyEvery3Days RotationFrequency = "EVERY_3_DAYS"
	FrequencyWeekly     RotationFrequency = "WEEKLY"
	FrequencyMonthly    RotationFrequency = "MONTHLY"
	FrequencyCustom     RotationFrequency = "CUSTOM"
)

type Group struct {
	ID                  string            `json:"id" db:"id"`
	Name                string            `json:"name" db:"name"`
	PhotoURL            string            `json:"photo_url" db:"photo_url"`
	ContributionAmount  float64           `json:"contribution_amount" db:"contribution_amount"`
	RotationFrequency   RotationFrequency `json:"rotation_frequency" db:"rotation_frequency"`
	CustomFrequencyDays int               `json:"custom_frequency_days" db:"custom_frequency_days"` // Only used if Frequency is CUSTOM
	CreatorID           string            `json:"creator_id" db:"creator_id"`
	InviteLink          string            `json:"invite_link" db:"invite_link"`
	CreatedAt           time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at" db:"updated_at"`
	Description         string            `json:"description" db:"description"`
	Currency            string            `json:"currency" db:"currency"`
	IsStarted           bool              `json:"is_started" db:"is_started"`
	StartDate           *time.Time        `json:"start_date" db:"start_date"`
	TotalRounds         int               `json:"total_rounds" db:"total_rounds"`
}

type ContributionStatus string

const (
	ContributionPaid   ContributionStatus = "PAID"
	ContributionUnpaid ContributionStatus = "UNPAID"
	ContributionExempt ContributionStatus = "EXEMPT"
)

type Contribution struct {
	ID        string             `json:"id" db:"id"`
	RoundID   string             `json:"round_id" db:"round_id"`
	UserID    string             `json:"user_id" db:"user_id"`
	Amount    float64            `json:"amount" db:"amount"`
	Status    ContributionStatus `json:"status" db:"status"`
	PaidAt    *time.Time         `json:"paid_at" db:"paid_at"`
	CreatedAt time.Time          `json:"created_at" db:"created_at"`
}

type UserShort struct {
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	PhotoURL string `json:"photo_url"`
}

type RoundDetail struct {
	RoundNumber     int         `json:"round_number"`
	StartDate       time.Time   `json:"start_date"`
	EndDate         time.Time   `json:"end_date"`
	Status          RoundStatus `json:"status"`
	Receiver        *UserShort  `json:"receiver"`
	PotAmount       float64     `json:"pot_amount"`
	CollectedAmount float64     `json:"collected_amount"`
}

type MemberRoundStatus struct {
	UserID        string             `json:"user_id"`
	Name          string             `json:"name"`
	PhotoURL      string             `json:"photo_url"`
	PhoneNumber   string             `json:"phone_number"`
	Role          MemberRole         `json:"role"`
	RoundNumber   int                `json:"round_number"` // When this member receives the pot
	PaymentStatus ContributionStatus `json:"payment_status"`
	JoinedAt      time.Time          `json:"joined_at"`
}

type GroupStats struct {
	TotalMembers         int     `json:"total_members"`
	PaidMembersCount     int     `json:"paid_members_count"`
	UnpaidMembersCount   int     `json:"unpaid_members_count"`
	CompletionPercentage float64 `json:"completion_percentage"`
}

type GroupHomeDetail struct {
	Group
	Role MemberRole `json:"role"`
}

type GroupHomeResponse struct {
	Group        GroupHomeDetail     `json:"group"`
	CurrentRound *RoundDetail        `json:"current_round"`
	Members      []MemberRoundStatus `json:"members"`
	Stats        GroupStats          `json:"stats"`
}

type MemberPaymentStatus struct {
	UserID        string             `json:"user_id"`
	Name          string             `json:"name"`
	PhotoURL      string             `json:"photo_url"`
	PaymentStatus ContributionStatus `json:"payment_status"`
	IsReceiver    bool               `json:"is_receiver"`
}

type RoundScheduleItem struct {
	RoundNumber int                   `json:"round_number"`
	StartDate   time.Time             `json:"start_date"`
	EndDate     time.Time             `json:"end_date"`
	Status      RoundStatus           `json:"status"`
	Receiver    *UserShort            `json:"receiver"`
	Members     []MemberPaymentStatus `json:"members"`
}

type GroupScheduleResponse struct {
	Rounds []RoundScheduleItem `json:"rounds"`
}

type MemberRole string

const (
	RoleAdmin  MemberRole = "ADMIN"
	RoleMember MemberRole = "MEMBER"
)

type MemberStatus string

const (
	StatusActive  MemberStatus = "ACTIVE"
	StatusInvited MemberStatus = "INVITED"
	StatusLeft    MemberStatus = "LEFT"
)

type GroupMember struct {
	ID       string       `json:"id" db:"id"`
	GroupID  string       `json:"group_id" db:"group_id"`
	UserID   string       `json:"user_id" db:"user_id"`
	Role     MemberRole   `json:"role" db:"role"`
	Status   MemberStatus `json:"status" db:"status"`
	JoinedAt time.Time    `json:"joined_at" db:"joined_at"`
}

type RoundStatus string

const (
	RoundStatusPending   RoundStatus = "PENDING"
	RoundStatusActive    RoundStatus = "ACTIVE"
	RoundStatusCompleted RoundStatus = "COMPLETED"
)

type Round struct {
	ID          string      `json:"id" db:"id"`
	GroupID     string      `json:"group_id" db:"group_id"`
	RoundNumber int         `json:"round_number" db:"round_number"`
	RecipientID *string     `json:"recipient_id" db:"recipient_id"` // Nullable if not yet assigned
	StartDate   time.Time   `json:"start_date" db:"start_date"`
	EndDate     time.Time   `json:"end_date" db:"end_date"`
	Status      RoundStatus `json:"status" db:"status"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
}

type GroupDetail struct {
	Group
	Role         MemberRole `json:"role"`
	MembersCount int        `json:"members_count"`
	CurrentRound *Round     `json:"current_round,omitempty"`
}

type GroupWithRole struct {
	Group
	Role        MemberRole `json:"role"`
	LastMessage *Message   `json:"last_message,omitempty"`
}

type GroupMemberView struct {
	UserID   string       `json:"user_id" db:"user_id"`
	Name     string       `json:"name" db:"name"`
	PhotoURL string       `json:"photo_url" db:"photo_url"`
	Role     MemberRole   `json:"role" db:"role"`
	Status   MemberStatus `json:"status" db:"status"`
	JoinedAt time.Time    `json:"joined_at" db:"joined_at"`
}

type GroupRepository interface {
	Create(group *Group) error
	AddMember(member *GroupMember) error
	GetMember(groupID, userID string) (*GroupMember, error)
	FindByID(id string) (*Group, error)
	CreateGroupWithMembers(group *Group, members []GroupMember, initialMessage *Message) error
	AddMemberWithSystemMessage(member *GroupMember, systemMessage *Message) error
	GetGroupsByUserID(userID string) ([]GroupWithRole, error)
	GetGroupDetail(groupID, userID string) (*GroupDetail, error)
	GetGroupMemberIDs(groupID string) ([]string, error)
	GetGroupMembers(groupID string) ([]GroupMemberView, error)
	GetGroupMembersPaginated(groupID string, limit, offset int) ([]GroupMemberView, int, error)
	HasCommonActiveGroup(user1ID, user2ID string) (bool, error)
	GetGroupHome(groupID, userID string) (*GroupHomeResponse, error)
	GetGroupSchedule(groupID string, filterStatus *RoundStatus) (*GroupScheduleResponse, error)
}
