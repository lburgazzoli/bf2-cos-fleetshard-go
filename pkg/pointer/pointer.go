package pointer

// Of ---
// TODO: improve
func Of[T any](value T) *T {
	return &value
}
