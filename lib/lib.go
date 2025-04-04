// Package lib implements G2P (grapheme-to-phoneme) IPA phonemizer/dephonemizer 
// for 136+ human languages, including validation and normalization.
package lib

import (
	. "github.com/martinarisk/di/dependency_injection"
)
import "github.com/neurlang/goruut/dicts"
import "github.com/neurlang/goruut/usecases"
import "github.com/neurlang/goruut/models/requests"
import "github.com/neurlang/goruut/models/responses"
import "github.com/neurlang/goruut/repo/interfaces"

type Phonemizer struct {
	uc usecases.IPhonemizeUsecase
}

type dummy struct {
}

func (dummy) GetIpaFlavors() map[string]map[string]string {
	return make(map[string]map[string]string)
}
func (dummy) GetPolicyMaxWords() int {
	return 99999999999
}

// NewPhonemizer creates a new phonemizer. Parameter di can be nil.
func NewPhonemizer(di *DependencyInjection) *Phonemizer {
	if di == nil {
		di = NewDependencyInjection()
		di.Add((interfaces.DictGetter)(dicts.DictGetter{}))
		di.Add((interfaces.IpaFlavor)(dummy{}))
		di.Add((interfaces.PolicyMaxWords)(dummy{}))
	}
	uc := usecases.NewPhonemizeUsecase(di)
	return &Phonemizer{
		uc: uc,
	}
}

// Sentence runs the algorithm on a sentence string in a specific language.
func (p *Phonemizer) Sentence(r requests.PhonemizeSentence) responses.PhonemizeSentence {
	return p.uc.Sentence(r)
}
