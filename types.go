package tat

// private
var (
	// TatHeaderUsername is a Header send by user with his username
	TatHeaderUsername          = "Tat_username"
	TatHeaderPassword          = "Tat_password"
	TatHeaderPasswordLower     = "tat_password"
	TatHeaderPasswordLowerDash = "tat-password"
	// TatHeaderXTatRefererLower contains tat microservice name & version "X-TAT-FROM"
	TatHeaderXTatRefererLower = "X-Tat-Referer"
)

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

//CacheableCriteria must return strnig slice describing the redis key
type CacheableCriteria interface {
	CacheKey() []string
}
