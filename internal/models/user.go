package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"column:password_hash" json:"-"`
	GoogleID     string    `gorm:"uniqueIndex" json:"google_id,omitempty"`
	GithubID     string    `gorm:"uniqueIndex" json:"github_id,omitempty"`
	LinkedinID   string    `gorm:"uniqueIndex" json:"linkedin_id,omitempty"`
	Firstname    string    `gorm:"not null" json:"firstname"`
	Lastname     string    `gorm:"not null" json:"lastname"`
	Bio          string    `gorm:"type:text" json:"bio"`
	Universite   string    `gorm:"type:varchar(255)" json:"universite"`
	AvatarURL    string    `json:"avatar_url"`
	Role         Role      `gorm:"type:varchar(20);not null;default:'USER'" json:"role"`
	Status       Status     `gorm:"type:varchar(20);not null;default:'ACTIVE'" json:"status"`
	BlockedUntil *time.Time `json:"blocked_until,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
