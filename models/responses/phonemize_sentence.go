package responses

import "encoding/json"

type PhonemizeSentence struct {
	Words []PhonemizeSentenceWord

	ErrorWordLimitExceeded bool `json:"ErrorWordLimitExceeded,omitempty"`
}

type PhonemizeSentenceWord struct {
	CleanWord string
	Phonetic  string
	PosTags   json.RawMessage
	PrePunct  string
	PostPunct string
}
