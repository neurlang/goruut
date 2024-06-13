package requests

type PhonemizeSentence struct {
	IpaFlavors []string
	Language   string
	Sentence   string
}
