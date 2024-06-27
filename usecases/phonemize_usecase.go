package usecases

import (
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/models/requests"
	"github.com/neurlang/goruut/models/responses"
	"github.com/neurlang/goruut/repo/interfaces"
	"github.com/neurlang/goruut/repo/services"
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

	splitted := p.service.SplitWords(r.Language, r.Sentence)

	if len(splitted) > p.maxwrds {
		return responses.PhonemizeSentence{ErrorWordLimitExceeded: true}
	}

	var phonemized []map[uint64]string

	for _, word := range splitted {
		phonemized = append(phonemized, p.phon.PhonemizeWord(r.Language, word))
		log.Now().Debugf("Word: %s %s", word, phonemized)
	}

	parts_of_speech_selected := p.sel.Select(r.Language, phonemized)

	var ipa_flavored []string
	if r.IpaFlavors != nil {
		for _, word := range parts_of_speech_selected {
			for _, flavor := range r.IpaFlavors {
				word = p.flavor.Apply(flavor, word)
			}
			ipa_flavored = append(ipa_flavored, word)
		}
	} else {
		ipa_flavored = parts_of_speech_selected
	}

	for i := range splitted {
		resp.Words = append(resp.Words, responses.PhonemizeSentenceWord{
			Phonetic:   ipa_flavored[i],
			Linguistic: splitted[i],
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
