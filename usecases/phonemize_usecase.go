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
	"sync/atomic"
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
	sent    services.ISentencizerService
	maxwrds uint64
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

	var sentences = []string{r.Sentence}
	if r.SplitSentences && !r.IsReverse {
		sentences = p.sent.Split(r.Language, r.Sentence)
	}
	var totalLenSplitted atomic.Uint64
	var ipa_flavored = make([][][3]string, len(sentences), len(sentences))
	var punctuation = make([][][2]string, len(sentences), len(sentences))
	parallel.ForEach(len(sentences), 10, func(j int) {

		splitted := p.service.SplitWords(r.IsReverse, r.Language, sentences[j])

		totalLenSplitted.Add(uint64(len(splitted)))

		if totalLenSplitted.Load() > p.maxwrds {
			return
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
		punctuation[j] = collapse(punctuation_all)

		if totalLenSplitted.Load() > p.maxwrds {
			return
		}

		parts_of_speech_selected := p.sel.Select(r.IsReverse, r.Language, phonemized, r.Languages)
		log.Now().Debugf("Vector: %v", parts_of_speech_selected)

		if totalLenSplitted.Load() > p.maxwrds {
			return
		}

		if r.IpaFlavors != nil {
			for _, word := range parts_of_speech_selected {
				for _, flavor := range r.IpaFlavors {
					word[1] = p.flavor.Apply(flavor, word[1])
				}
				ipa_flavored[j] = append(ipa_flavored[j], word)
			}
		} else {
			ipa_flavored[j] = parts_of_speech_selected
		}
		log.Now().Debugf("Splitted: %d, Phonemized: %d, POS: %d, Flavored: %d",
			len(splitted), len(phonemized), len(parts_of_speech_selected), len(ipa_flavored[j]))

	})
	if totalLenSplitted.Load() > p.maxwrds {
		return responses.PhonemizeSentence{ErrorWordLimitExceeded: true}
	}
	resp.Init()
	for j := range ipa_flavored {
		for i := range ipa_flavored[j] {
			resp.Words = append(resp.Words, responses.PhonemizeSentenceWord{
				Phonetic:  strings.Trim(ipa_flavored[j][i][1], "_"),
				CleanWord: strings.TrimRight(ipa_flavored[j][i][0], " "),
				PosTags:   json.RawMessage(ipa_flavored[j][i][2]),
				PrePunct:  punctuation[j][i][0],
				PostPunct: punctuation[j][i][1],
				IsFirst:   i == 0,
				IsLast:    i == len(ipa_flavored[j])-1,
			})
			//resp.Whole += ipa_flavored[i]
		}
	}

	return
}

func NewPhonemizeUsecase(di *DependencyInjection) *PhonemizeUsecase {
	service := MustNeed(di, services.NewSplitWordsService)
	phon := MustNeed(di, services.NewPhonemizeWordService)
	sel := MustNeed(di, services.NewPartsOfSpeechSelectorService)
	flavor := MustNeed(di, services.NewIpaFlavorService)
	sent := MustNeed(di, services.NewSentencizerService)
	policyMaxWords := MustAny[interfaces.PolicyMaxWords](di)

	return &PhonemizeUsecase{
		service: &service,
		phon:    &phon,
		sel:     &sel,
		flavor:  &flavor,
		sent:    &sent,
		maxwrds: uint64(policyMaxWords.GetPolicyMaxWords()),
	}
}

var _ IPhonemizeUsecase = &PhonemizeUsecase{}
