package models

type User struct {
	Firstname  string `gorm:"unique;not null" json:"firstname"`
	Lastname   string `gorm:"unique;not null" json:"lastname"`
	Email      string `gorm:"unique;not null" json:"email"`
	Password   string `gorm:"not null" json:"password"`
	Universite string `gorm:"not null" json:"universite"`
	Role       string `gorm:"not null;default:USER" json:"role"`
	Status     string `gorm:"not null;default:STUDENT" json:"status"`
	Bio        string `gorm:"not null" json:"bio"`
}
