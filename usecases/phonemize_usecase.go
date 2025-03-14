package usecases

import (
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/models/requests"
	"github.com/neurlang/goruut/models/responses"
	"github.com/neurlang/goruut/repo/interfaces"
	"github.com/neurlang/goruut/repo/services"
	"strings"
	"encoding/json"
)
import . "github.com/martinarisk/di/dependency_injection"

type IPhonemizeUsecase interface {
	Sentence(requests.PhonemizeSentence) responses.PhonemizeSentence
}

type PhonemizeUsecase struct {
	service services.ISplitWordsService
	phon    services.IPhonemizeWordService
	sel     services.IPartsOfSpeechSelectorService
	flavor  services.IIpaFlavorService
	maxwrds int
}

func (p *PhonemizeUsecase) Sentence(r requests.PhonemizeSentence) (resp responses.PhonemizeSentence) {
	r.Init()

	splitted := p.service.SplitWords(r.IsReverse, r.Language, r.Sentence)

	if len(splitted) > p.maxwrds {
		return responses.PhonemizeSentence{ErrorWordLimitExceeded: true}
	}

	var phonemized []map[uint32]string
	var punctuation [][2]string

	for _, word := range splitted {
		words, punct := p.phon.PhonemizeWords(r.IsReverse, r.Language, word, r.Languages)
		phonemized = append(phonemized, words...)
		punctuation = append(punctuation, punct...)
		log.Now().Debugf("Word: %s, Words: %v", word, words)
	}

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
