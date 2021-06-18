package ptr

func String(v string) *string {
	return &v
}

func GetString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func Int64(v int64) *int64 {
	return &v
}

func GetInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}
