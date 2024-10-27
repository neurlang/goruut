package lib

import "testing"
import "github.com/neurlang/goruut/models/requests"

func TestOne(t *testing.T) {
	p := NewPhonemizer(nil)
	resp := p.Sentence(requests.PhonemizeSentence{
		Sentence: "hello world",
		Language: "English",
	})
	for i := range resp.Words {
		println(resp.Words[i].Phonetic)
	}
}
