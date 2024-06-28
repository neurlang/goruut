package responses

type PhonemizeSentence struct {
	Words []PhonemizeSentenceWord

	ErrorWordLimitExceeded bool `json:"ErrorWordLimitExceeded,omitempty"`
}

type PhonemizeSentenceWord struct {
	CleanWord  string
	Linguistic string
	Phonetic   string
}
