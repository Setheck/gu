package util

// Int32InSlice given a slice of int32 return the index of the value or -1 if not in the slice
func Int32InSlice(slice []int32, value int32) int {
	for idx, item := range slice {
		if value == item {
			return idx
		}
	}
	return -1
}

// StringInSlice given a slice of string return the index of s or -1 if not in the slice
func StringInSlice(slice []string, s string) int {
	for idx, str := range slice {
		if s == str {
			return idx
		}
	}
	return -1
}
