package birdwatcherV2

import (
	"strconv"
)

/*
 * Types helper for parser
 */

// Assert string, provide default
func mustString(value interface{}, fallback string) string {
	sval, ok := value.(string)
	if !ok {
		return fallback
	}
	return sval
}

// Convert list of strings to int
func mustIntList(data []string) []int {
	var list []int
	for _, e := range data {
		val, _ := strconv.Atoi(e)
		list = append(list, val)
	}
	return list
}
