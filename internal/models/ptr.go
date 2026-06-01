package models

// StringPtr returns nil for empty strings so GORM stores NULL (unique indexes allow multiple NULLs).
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
