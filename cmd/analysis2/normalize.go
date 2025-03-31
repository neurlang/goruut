package main

import (
	"golang.org/x/text/unicode/norm"
	"regexp"
	"sort"
)

// Function to sort combiners runs
func SortCombiningRuns(s string) string {
	re := regexp.MustCompile(`[\p{Mn}\p{Mc}]+`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		runes := []rune(match)
		sort.Slice(runes, func(i, j int) bool { return runes[i] > runes[j] })
		return string(runes)
	})
}

// Function to normalize text to NFC
func NormalizeTo(input, form string) string {
	var format norm.Form

	switch form {
	case "NFC":
		format = norm.NFC
	case "NFD":
		format = norm.NFD
	case "NFKC":
		format = norm.NFKC
	case "NFKD":
		format = norm.NFKD
	case "Sort":
		return SortCombiningRuns(input)
	default:
		return input
	}

	return format.String(input)
}
