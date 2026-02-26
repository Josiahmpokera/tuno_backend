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
	Role MemberRole `json:"role"`
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
	FindByID(id string) (*Group, error)
	CreateGroupWithMembers(group *Group, members []GroupMember, initialMessage *Message) error
	GetGroupsByUserID(userID string) ([]GroupWithRole, error)
	GetGroupDetail(groupID, userID string) (*GroupDetail, error)
	GetGroupMemberIDs(groupID string) ([]string, error)
	GetGroupMembers(groupID string) ([]GroupMemberView, error)
	HasCommonActiveGroup(user1ID, user2ID string) (bool, error)
}
