package models

type User struct {
	ID         uint   `json:"id"`
	Firstname  string `gorm:"not null" json:"firstname"`
	Lastname   string `gorm:"not null" json:"lastname"`
	Email      string `gorm:"unique;not null" json:"email"`
	Password   string `gorm:"not null" json:"password"`
	Universite string `gorm:"not null" json:"universite"`
	Role       string `gorm:"not null;default:USER" json:"role"`
	Status     string `gorm:"not null;default:STUDENT" json:"status"`
	Bio        string `json:"bio"`
}
