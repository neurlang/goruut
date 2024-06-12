package usecases

import (
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/models/requests"
	"github.com/neurlang/goruut/models/responses"
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
}

func (p *PhonemizeUsecase) Sentence(r requests.PhonemizeSentence) (resp responses.PhonemizeSentence) {

	splitted := p.service.SplitWords(r.Language, r.Sentence)

	var phonemized []map[uint64]string

	for _, word := range splitted {
		phonemized = append(phonemized, p.phon.PhonemizeWord(r.Language, word))
		log.Now().Debugf("Word: %s %s", word, phonemized)
	}

	parts_of_speech_selected := p.sel.Select(r.Language, phonemized)

	for i := range splitted {
		resp.Words = append(resp.Words, responses.PhonemizeSentenceWord{
			Phonetic:   parts_of_speech_selected[i],
			Linguistic: splitted[i],
		})
	}

	return
}

func NewPhonemizeUsecase(di *DependencyInjection) *PhonemizeUsecase {
	service := MustNeed(di, services.NewSplitWordsService)
	phon := MustNeed(di, services.NewPhonemizeWordService)
	sel := MustNeed(di, services.NewPartsOfSpeechSelectorService)

	return &PhonemizeUsecase{
		service: &service,
		phon:    &phon,
		sel:     &sel,
	}
}

var _ IPhonemizeUsecase = &PhonemizeUsecase{}
