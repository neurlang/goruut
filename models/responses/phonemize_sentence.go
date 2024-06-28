package responses

type PhonemizeSentence struct {
	Words []PhonemizeSentenceWord

	ErrorWordLimitExceeded bool `json:"ErrorWordLimitExceeded,omitempty"`
}

type PhonemizeSentenceWord struct {
	CleanLing  string
	Linguistic string
	Phonetic   string
}
