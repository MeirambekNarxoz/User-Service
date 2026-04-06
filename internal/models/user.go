package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"column:password_hash;not null" json:"-"`
	Firstname    string    `gorm:"not null" json:"firstname"`
	Lastname     string    `gorm:"not null" json:"lastname"`
	Bio          string    `gorm:"type:text" json:"bio"`
	Role         Role      `gorm:"type:varchar(20);not null;default:'USER'" json:"role"`
	Status       Status    `gorm:"type:varchar(20);not null;default:'ACTIVE'" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}
