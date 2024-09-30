package global

func Contains[K comparable](slice []K, el K, emptyContains bool) bool {
	if emptyContains && len(slice) == 0 {
		return true
	}
	for _, s := range slice {
		if el == s {
			return true
		}
	}
	return false
}
