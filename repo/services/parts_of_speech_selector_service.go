package services

import "strings"
import . "github.com/martinarisk/di/dependency_injection"

type IPartsOfSpeechSelectorService interface {
	Select(string, []map[uint64]string) []string
	SelectCJK(string, []map[uint64][2]string) [2][]string
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

func (p *PartsOfSpeechSelectorService) SelectCJK(lang string, sentence []map[uint64][2]string) (ret [2][]string) {
	for _, words := range sentence {
		for _, word := range words {
			dst := strings.Split(word[0], "_")
			src := strings.Split(word[1], "_")

			if len(dst) != len(src) {
				dst = strings.Split(strings.Trim(word[0], "_"), "_")
				src = strings.Split(strings.Trim(word[1], "_"), "_")
				if len(dst) != len(src) {

					// shouldn't happen??
				} else {
					ret[0] = append(ret[0], dst...)
					ret[1] = append(ret[1], src...)
					break
				}
			} else {
				ret[0] = append(ret[0], dst...)
				ret[1] = append(ret[1], src...)
				break
			}
		}
	}
	return
}

func NewPartsOfSpeechSelectorService(di *DependencyInjection) *PartsOfSpeechSelectorService {

	return &PartsOfSpeechSelectorService{}
}

var _ IPartsOfSpeechSelectorService = &PartsOfSpeechSelectorService{}
