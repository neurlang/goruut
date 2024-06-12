package services

import . "github.com/martinarisk/di/dependency_injection"

type IPartsOfSpeechSelectorService interface {
	Select(string, []map[uint64]string) []string
}

type PartsOfSpeechSelectorService struct {
}

func (p *PartsOfSpeechSelectorService) Select(lang string, sentence []map[uint64]string) (ret []string) {
	for _, words := range sentence {
		for _, word := range words {
			ret = append(ret, word)
			break
		}
	}
	return
}

func NewPartsOfSpeechSelectorService(di *DependencyInjection) *PartsOfSpeechSelectorService {

	return &PartsOfSpeechSelectorService{}
}

var _ IPartsOfSpeechSelectorService = &PartsOfSpeechSelectorService{}
