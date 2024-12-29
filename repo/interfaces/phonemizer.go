package interfaces

type Phonemizer interface {
	Map(isReverse bool, lang string) map[string]map[string]struct{}
	SrcMulti(isReverse bool, lang string) map[string]struct{}
	DstMulti(isReverse bool, lang string) map[string]struct{}
	SrcMultiSuffix(isReverse bool, lang string) map[string]struct{}
	DstMultiSuffix(isReverse bool, lang string) map[string]struct{}
}
