// Package usecases orchestrates business workflows between controllers and services.
package usecases

import (
	"encoding/json"
	"github.com/neurlang/classifier/parallel"
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/models/requests"
	"github.com/neurlang/goruut/models/responses"
	"github.com/neurlang/goruut/repo/interfaces"
	"github.com/neurlang/goruut/repo/services"
	"strings"
)
import . "github.com/martinarisk/di/dependency_injection"

type IPhonemizeUsecase interface {
	Sentence(requests.PhonemizeSentence) responses.PhonemizeSentence
	Word(requests.ExplainWord) responses.ExplainWord
}

type PhonemizeUsecase struct {
	service services.ISplitWordsService
	phon    services.IPhonemizeWordService
	sel     services.IPartsOfSpeechSelectorService
	flavor  services.IIpaFlavorService
	maxwrds int
}

func (p *PhonemizeUsecase) Word(r requests.ExplainWord) (resp responses.ExplainWord) {
	return responses.ExplainWord{
		Rules: p.phon.ExplainWord(r.IsReverse, r.CleanWord, r.Phonetic, r.Language),
	}
}

func collapse[T any](slice [][]T) (ret []T) {
	for _, subslice := range slice {
		ret = append(ret, subslice...)
	}
	return
}

func (p *PhonemizeUsecase) Sentence(r requests.PhonemizeSentence) (resp responses.PhonemizeSentence) {
	r.Init()

	splitted := p.service.SplitWords(r.IsReverse, r.Language, r.Sentence)

	if len(splitted) > p.maxwrds {
		return responses.PhonemizeSentence{ErrorWordLimitExceeded: true}
	}

	var phonemized_all = make([][]map[string]uint32, len(splitted), len(splitted))
	var punctuation_all = make([][][2]string, len(splitted), len(splitted))

	parallel.ForEach(len(splitted), 1000, func(i int) {
		word := splitted[i]
		words, punct := p.phon.PhonemizeWords(r.IsReverse, r.Language, word, r.Languages)
		phonemized_all[i] = words
		punctuation_all[i] = punct
		log.Now().Debugf("Word: %s, Words: %v", word, words)
	})
	var phonemized = collapse(phonemized_all)
	var punctuation = collapse(punctuation_all)

	parts_of_speech_selected := p.sel.Select(r.IsReverse, r.Language, phonemized, r.Languages)
	log.Now().Debugf("Vector: %v", parts_of_speech_selected)

	var ipa_flavored [][3]string
	if r.IpaFlavors != nil {
		for _, word := range parts_of_speech_selected {
			for _, flavor := range r.IpaFlavors {
				word[1] = p.flavor.Apply(flavor, word[1])
			}
			ipa_flavored = append(ipa_flavored, word)
		}
	} else {
		ipa_flavored = parts_of_speech_selected
	}
	log.Now().Debugf("Splitted: %d, Phonemized: %d, POS: %d, Flavored: %d",
		len(splitted), len(phonemized), len(parts_of_speech_selected), len(ipa_flavored))

	resp.Init()
	for i := range ipa_flavored {
		resp.Words = append(resp.Words, responses.PhonemizeSentenceWord{
			Phonetic:  strings.Trim(ipa_flavored[i][1], "_"),
			CleanWord: ipa_flavored[i][0],
			PosTags:   json.RawMessage(ipa_flavored[i][2]),
			PrePunct:  punctuation[i][0],
			PostPunct: punctuation[i][1],
		})
		//resp.Whole += ipa_flavored[i]
	}

	return
}

func NewPhonemizeUsecase(di *DependencyInjection) *PhonemizeUsecase {
	service := MustNeed(di, services.NewSplitWordsService)
	phon := MustNeed(di, services.NewPhonemizeWordService)
	sel := MustNeed(di, services.NewPartsOfSpeechSelectorService)
	flavor := MustNeed(di, services.NewIpaFlavorService)
	policyMaxWords := MustAny[interfaces.PolicyMaxWords](di)

	return &PhonemizeUsecase{
		service: &service,
		phon:    &phon,
		sel:     &sel,
		flavor:  &flavor,
		maxwrds: policyMaxWords.GetPolicyMaxWords(),
	}
}

var _ IPhonemizeUsecase = &PhonemizeUsecase{}
