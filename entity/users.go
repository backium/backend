package entity

import "golang.org/x/crypto/bcrypt"

type UserKind string

const (
	UserKindOwner    UserKind = "owner"
	UserKindEmployee          = "employee"
	UserKindSuper             = "super"
)

type User struct {
	ID           string   `bson:"_id"`
	Email        string   `bson:"email,omitempty"`
	PasswordHash string   `bson:"password_hash,omitempty"`
	Kind         UserKind `bson:"kind,omitempty"`
	MerchantID   string   `bson:"merchant_id,omitempty"`
}

func (u *User) PasswordEquals(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

func HashUserPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
