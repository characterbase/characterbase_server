package models

import (
	"golang.org/x/crypto/bcrypt"
)

// User represents a CharacterBase user
type User struct {
	ID           string `json:"id" db:"id"`
	DisplayName  string `json:"display_name" db:"display_name"`
	Email        string `json:"email" gorm:"unique;not null" db:"email"`
	PasswordHash string `json:"-" gorm:"not null" db:"password_hash"`
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
