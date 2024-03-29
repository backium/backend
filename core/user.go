package core

import (
	"context"

	"github.com/backium/backend/errors"
	"golang.org/x/crypto/bcrypt"
)

type UserKind string

const (
	UserKindOwner    UserKind = "owner"
	UserKindEmployee UserKind = "employee"
	UserKindCustomer UserKind = "customer"
	UserKindSuper    UserKind = "super"
)

type User struct {
	ID           ID       `bson:"_id"`
	Email        string   `bson:"email"`
	PasswordHash string   `bson:"password_hash,omitempty"`
	Kind         UserKind `bson:"kind"`
	EmployeeID   ID       `bson:"employee_id"`
	MerchantID   ID       `bson:"merchant_id"`
}

func NewUserOwner() User {
	return User{
		ID:   NewID("user"),
		Kind: UserKindOwner,
	}
}

func NewUserEmployee(merchantID, employeeID ID) User {
	return User{
		ID:         NewID("user"),
		Kind:       UserKindEmployee,
		EmployeeID: employeeID,
		MerchantID: merchantID,
	}
}

func (u *User) PasswordEquals(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

func (u *User) HashPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

type UserStorage interface {
	Put(context.Context, User) error
	Get(context.Context, ID) (User, error)
	GetByEmail(context.Context, string) (User, error)
}

type UserService struct {
	UserStorage       UserStorage
	MerchantStorage   MerchantStorage
	LocationStorage   LocationStorage
	EmployeeStorage   EmployeeStorage
	CashDrawerStorage CashDrawerStorage
}

func (svc *UserService) Create(ctx context.Context, user User, password string) (User, error) {
	const op = errors.Op("controller.User.Create")

	if _, err := svc.UserStorage.GetByEmail(ctx, user.Email); err == nil {
		return User{}, errors.E(op, errors.KindUserExist, "user email used")
	}

	if err := user.HashPassword(password); err != nil {
		return User{}, errors.E(op, errors.KindUnexpected, err)
	}

	switch user.Kind {
	case UserKindOwner:
		merchant := NewMerchant()
		merchant.BusinessName = "My Business"
		merchant.Currency = PEN
		if err := svc.MerchantStorage.Put(ctx, merchant); err != nil {
			return User{}, errors.E(op, errors.KindUnexpected, err)

		}

		location := NewLocation("My Business", merchant.ID)
		location.BusinessName = "My Business"
		if err := svc.LocationStorage.Put(ctx, location); err != nil {
			return User{}, errors.E(op, errors.KindUnexpected, err)
		}

		employee := NewEmployee("Jhon", "Doe", merchant.ID)
		employee.IsOwner = true
		employee.Email = user.Email
		if err := svc.EmployeeStorage.Put(ctx, employee); err != nil {
			return User{}, errors.E(op, errors.KindUnexpected, err)
		}

		cash := NewCashDrawer(location.ID, merchant.ID)
		cash.Amount = NewMoney(0, merchant.Currency)

		if err := svc.CashDrawerStorage.Put(ctx, cash); err != nil {
			return User{}, errors.E(op, errors.KindUnexpected, err)
		}

		user.MerchantID = merchant.ID
		user.EmployeeID = employee.ID
	case UserKindEmployee:
		employee, err := svc.EmployeeStorage.Get(ctx, user.EmployeeID)
		if err != nil {
			return User{}, errors.E(op, errors.KindValidation, "Provided employee not found")
		}

		if employee.MerchantID != user.MerchantID {
			return User{}, errors.E(op, errors.KindValidation, "Provided employee doesn't belong to your business")
		}

		employee.IsUser = true
		if err := svc.EmployeeStorage.Put(ctx, employee); err != nil {
			return User{}, errors.E(op, errors.KindUnexpected)
		}
	default:
		return User{}, errors.E(op, errors.KindValidation, "Unknown user kind")
	}

	if err := svc.UserStorage.Put(ctx, user); err != nil {
		return User{}, errors.E(op, errors.KindUnexpected, err)
	}

	user, err := svc.UserStorage.Get(ctx, user.ID)
	if err != nil {
		return User{}, errors.E(op, err)
	}

	return user, nil
}

func (svc *UserService) Login(ctx context.Context, email, password string) (User, error) {
	const op = errors.Op("controller.User.Login")

	user, err := svc.UserStorage.GetByEmail(ctx, email)
	if err != nil {
		return User{}, errors.E(op, errors.KindInvalidCredentials, err)
	}

	if !user.PasswordEquals(password) {
		return User{}, errors.E(op, errors.KindInvalidCredentials, "invalid password")
	}

	return user, nil
}
