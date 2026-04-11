package models

import "time"

type FriendshipStatus string

const (
	FriendshipPending  FriendshipStatus = "PENDING"
	FriendshipAccepted FriendshipStatus = "ACCEPTED"
	FriendshipRejected FriendshipStatus = "REJECTED"
)

type Friendship struct {
	ID        uint             `gorm:"primaryKey" json:"id"`
	UserID    uint             `gorm:"not null;uniqueIndex:idx_user_friend" json:"user_id"`
	FriendID  uint             `gorm:"not null;uniqueIndex:idx_user_friend" json:"friend_id"`
	Status    FriendshipStatus `gorm:"type:varchar(20);not null;default:'PENDING'" json:"status"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`

	// Associations
	User   User `gorm:"foreignKey:UserID" json:"-"`
	Friend User `gorm:"foreignKey:FriendID" json:"friend_info"`
}
