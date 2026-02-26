package domain

import "time"

// Events

type Event interface {
	EventName() string
}

type GroupCreatedEvent struct {
	GroupID             string            `json:"group_id"`
	Name                string            `json:"name"`
	PhotoURL            string            `json:"photo_url"`
	ContributionAmount  float64           `json:"contribution_amount"`
	RotationFrequency   RotationFrequency `json:"rotation_frequency"`
	CustomFrequencyDays int               `json:"custom_frequency_days"`
	CreatorID           string            `json:"creator_id"`
	InviteLink          string            `json:"invite_link"`
	Members             []GroupMember     `json:"members"`
	CreatedAt           time.Time         `json:"created_at"`
}

func (e GroupCreatedEvent) EventName() string {
	return "GroupCreatedEvent"
}
