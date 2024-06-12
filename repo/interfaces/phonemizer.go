package interfaces

type Phonemizer interface {
	Map(lang string) map[string]map[string]struct{}
	SrcMulti(lang string) map[string]struct{}
	DstMulti(lang string) map[string]struct{}
	SrcMultiSuffix(lang string) map[string]struct{}
	DstMultiSuffix(lang string) map[string]struct{}
}
