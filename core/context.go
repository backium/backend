package core

import "context"

// contextKey is a unique type to avoid clashing with other packages that use
// context's to pass data.
type contextKey string

const (
	contextKeyMerchant = contextKey("merchant")
	contextKeyUser     = contextKey("user")
	contextKeyEmployee = contextKey("employee")
	contextKeySession  = contextKey("session")
)

func ContextWithMerchant(ctx context.Context, merchant *Merchant) context.Context {
	return context.WithValue(ctx, contextKeyMerchant, merchant)
}

func ContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, contextKeyUser, user)
}

func ContextWithEmployee(ctx context.Context, employee *Employee) context.Context {
	return context.WithValue(ctx, contextKeyEmployee, employee)
}

func ContextWithSession(ctx context.Context, sess *Session) context.Context {
	return context.WithValue(ctx, contextKeySession, sess)
}

func UserFromContext(ctx context.Context) *User {
	v := ctx.Value(contextKeyUser)
	if v == nil {
		return nil
	}

	t, ok := v.(*User)
	if !ok {
		return nil
	}
	return t
}

func MerchantFromContext(ctx context.Context) *Merchant {
	v := ctx.Value(contextKeyMerchant)
	if v == nil {
		return nil
	}

	t, ok := v.(*Merchant)
	if !ok {
		return nil
	}
	return t
}

func EmployeeFromContext(ctx context.Context) *Employee {
	v := ctx.Value(contextKeyEmployee)
	if v == nil {
		return nil
	}

	t, ok := v.(*Employee)
	if !ok {
		return nil
	}
	return t
}

func SessionFromContext(ctx context.Context) *Session {
	v := ctx.Value(contextKeySession)
	if v == nil {
		return nil
	}

	t, ok := v.(*Session)
	if !ok {
		return nil
	}
	return t
}
