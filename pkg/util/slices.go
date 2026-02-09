package util

// GetFirstNSafe 安全版本，处理nil切片
func GetFirstNSafe[T any](slice []T, n int) []T {
	if slice == nil {
		return nil
	}

	if n <= 0 {
		return []T{}
	}

	if n >= len(slice) {
		result := make([]T, len(slice))
		copy(result, slice)
		return result
	}

	result := make([]T, n)
	copy(result, slice[:n])
	return result
}
