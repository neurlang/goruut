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

	var phonemized []map[uint64][2]string
	var cleaned []string

	for _, word := range splitted {
		clean, phon := p.phon.PhonemizeWordCJK(r.Language, word)
		phonemized = append(phonemized, phon)
		cleaned = append(cleaned, clean)
		log.Now().Debugf("Word: %s, Cleaned: %s", word, clean)
	}

	parts_of_speech_selected := p.sel.SelectCJK(r.Language, phonemized)

	var ipa_flavored [2][]string
	if r.IpaFlavors != nil {
		for _, word := range parts_of_speech_selected {
			for _, flavor := range r.IpaFlavors {
				word[0] = p.flavor.Apply(flavor, word[0])
			}
			ipa_flavored[0] = append(ipa_flavored[0], word[0])
			ipa_flavored[1] = append(ipa_flavored[1], word[1])
		}
	} else {
		ipa_flavored = parts_of_speech_selected
	}
	log.Now().Debugf("Splitted: %d, Phonemized: %d, POS: %d, Flavored: %d",
		len(splitted), len(phonemized), len(parts_of_speech_selected), len(ipa_flavored[0]))

	for i := range ipa_flavored[0] {
		resp.Words = append(resp.Words, responses.PhonemizeSentenceWord{
			Phonetic:   ipa_flavored[0][i],
			Linguistic: ipa_flavored[1][i],
			CleanWord:  p.phon.CleanWord(r.Language, ipa_flavored[1][i]),
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
