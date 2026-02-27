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

// SafeLastElement 安全获取最后一个元素
func SafeLastElement[T any](slice []T) (T, bool) {
	if slice == nil || len(slice) == 0 {
		var zero T
		return zero, false
	}
	return slice[len(slice)-1], true
}

// RemoveDuplicates 泛型函数：适用于任何可比较的类型
func RemoveDuplicates[T comparable](slice []T) []T {
	seen := make(map[T]bool)
	result := []T{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

// FindCommonItemsGeneric 获取公共项
func FindCommonItemsGeneric[T comparable](slices ...[]T) []T {
	if len(slices) == 0 {
		return []T{}
	}

	set := make(map[T]bool)
	for _, item := range slices[0] {
		set[item] = true
	}

	for i := 1; i < len(slices); i++ {
		currentSet := make(map[T]bool)
		for _, item := range slices[i] {
			if set[item] {
				currentSet[item] = true
			}
		}
		set = currentSet

		if len(set) == 0 {
			return []T{}
		}
	}

	result := make([]T, 0, len(set))
	for item := range set {
		result = append(result, item)
	}

	return result
}
