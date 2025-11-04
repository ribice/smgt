package rot

func guardMapLookup(m map[string]int, key string) int {
	value, ok := m[key]
	if !ok {
		return 0
	}
	return value
}

func guardNilPointer(ptr *int) int {
	target := ptr
	if target == nil {
		return 0
	}
	return *target
}

func guardSliceLength(fetch func() ([]string, error)) (int, error) {
	items, err := fetch()
	if err != nil {
		return 0, err
	}
	if len(items) == 0 {
		return 0, nil
	}
	return len(items), nil
}

func guardWithinLoop(keys []string, m map[string]int) int {
	var total int
	for _, key := range keys {
		value, ok := m[key]
		if !ok {
			continue
		}
		total += value
	}
	return total
}
