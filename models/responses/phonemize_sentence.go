package responses

import "encoding/json"

type PhonemizeSentence struct {
	Words []PhonemizeSentenceWord

	ErrorWordLimitExceeded bool `json:"ErrorWordLimitExceeded,omitempty"`
}

func (p *PhonemizeSentence) Init() {
	if len(p.Words) == 0 {
		p.Words = []PhonemizeSentenceWord{}
	}
}

type PhonemizeSentenceWord struct {
	CleanWord string
	Phonetic  string
	PosTags   json.RawMessage
	PrePunct  string
	PostPunct string

	IsFirst bool
	IsLast  bool
}
