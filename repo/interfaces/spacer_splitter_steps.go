package interfaces

import "regexp"

type SpacerSplitterSteps interface {
	Len(lang string) int
	LeftRegexp(lang string, n int) *regexp.Regexp
	RightRegexp(lang string, n int) *regexp.Regexp
}
