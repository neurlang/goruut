package interfaces

type PrePhonemizationSteps interface {
	Len(isReverse bool, lang string) int
	IsNormalize(isReverse bool, lang string, n int) bool
	IsTrim(isReverse bool, lang string, n int) bool
	IsToLower(isReverse bool, lang string, n int) bool
	GetNormalize(isReverse bool, lang string, n int) string
	GetTrim(isReverse bool, lang string, n int) string
}
