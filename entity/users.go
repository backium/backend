package entity

import "golang.org/x/crypto/bcrypt"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	IsOwner      bool
	MerchantID   string
}

func (u *User) PasswordEquals(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

func HashUserPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
