package interfaces

type PrePhonemizationSteps interface {
	Len(lang string) int
	IsNormalize(lang string, n int) bool
	IsTrim(lang string, n int) bool
	IsToLower(lang string, n int) bool
	GetNormalize(lang string, n int) string
	GetTrim(lang string, n int) string
}
