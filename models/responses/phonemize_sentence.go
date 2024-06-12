package responses

type PhonemizeSentence struct {
	Words []PhonemizeSentenceWord
}

type PhonemizeSentenceWord struct {
	Linguistic string
	Phonetic   string
}
