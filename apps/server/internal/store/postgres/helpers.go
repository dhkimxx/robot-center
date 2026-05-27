package postgres

func stringFromPointer(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
