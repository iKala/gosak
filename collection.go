package gosak

import ()

// Qualifiable is the interface that represents those things could be filtered
// according to some criteria.
//
// IsQualified returns true if it satisfies all criteria, else return false.
//
type Qualifiable interface {
	IsQualified(...interface{}) bool
}

// GetQualifiedItems gets items which are qualified by qualifier from the item collection
func GetQualifiedItems(src []Qualifiable, criteria ...interface{}) (int, []Qualifiable) {
	count := 0
	var qualifiedItems []Qualifiable

	for _, item := range src {
		if item.IsQualified(criteria...) {
			count++
			qualifiedItems = append(qualifiedItems, item)
		}
	}

	return count, qualifiedItems
}

// InStringSlice check if the `src` string in the slice
func InStringSlice(slice []string, src string) bool {
	for _, s := range slice {
		if s == src {
			return true
		}
	}

	return false
}
