package utils

// ArrayContains return true if element is in array
func ArrayContains(array []string, element string) bool {
	for _, cur := range array {
		if cur == element {
			return true
		}
	}
	return false
}

// ItemInBothArrays return true if an element is in both array
func ItemInBothArrays(arrayA, arrayB []string) bool {
	for _, cura := range arrayA {
		for _, curb := range arrayB {
			if cura == curb {
				return true
			}
		}
	}
	return false
}
