package core

type Status string

const (
	StatusActive        Status = "active"
	StatusInactive      Status = "inactive"
	StatusShadowDeleted Status = "shadow_deleted"
)
