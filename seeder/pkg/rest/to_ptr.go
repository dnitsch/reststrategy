package rest

// Bool returns a pointer value for the bool value passed in.
func Bool(v bool) *bool {
	return &v
}

// Byte returns a pointer value for the byte value passed in.
func Byte(v byte) *byte {
	return &v
}

// String returns a pointer value for the string value passed in.
func String(v string) *string {
	return &v
}

// Int returns a pointer value for the int value passed in.
func Int(v int) *int {
	return &v
}
