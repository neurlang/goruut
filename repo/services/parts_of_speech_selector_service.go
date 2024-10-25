package services

import . "github.com/martinarisk/di/dependency_injection"

type IPartsOfSpeechSelectorService interface {
	Select(lang string, sentence []map[uint64]string) (ret [][2]string)
}

type PartsOfSpeechSelectorService struct {
}

func (p *PartsOfSpeechSelectorService) Select(lang string, sentence []map[uint64]string) (ret [][2]string) {
	for _, words := range sentence {
		var orig = words[0]
		for k, word := range words {
			if k == 0 {
				orig = word
				continue
			}
			ret = append(ret, [2]string{orig, word})
			break
		}
	}
	return
}

func NewPartsOfSpeechSelectorService(di *DependencyInjection) *PartsOfSpeechSelectorService {

	return &PartsOfSpeechSelectorService{}
}

var _ IPartsOfSpeechSelectorService = &PartsOfSpeechSelectorService{}
