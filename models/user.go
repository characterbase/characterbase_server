package models

import (
	"golang.org/x/crypto/bcrypt"
)

// User represents a CharacterBase user
type User struct {
	Generic
	DisplayName  string `json:"display_name"`
	Email        string `json:"email" gorm:"unique;not null"`
	PasswordHash string `json:"-" gorm:"not null"`
}

// SetPassword sets the user's password
func (u *User) SetPassword(plaintext string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 10)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}
