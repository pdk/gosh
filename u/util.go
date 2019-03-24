package u

// StringIn checks if a string is in a slice of strings.
func StringIn(s string, l []string) bool {
	for _, x := range l {
		if s == x {
			return true
		}
	}
	return false
}

// KeysOf returns the string keys of a string->bool map.
func KeysOf(m map[string]bool) []string {
	x := []string{}
	for k := range m {
		x = append(x, k)
	}
	return x
}
