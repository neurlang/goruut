package services

import (
	. "github.com/martinarisk/di/dependency_injection"
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/repo/interfaces"
	"strings"
)

type IIpaFlavorService interface {
	Apply(lang, word string) (ret string)
}

type IpaFlavorService struct {
	mapping *map[string]map[string]string
}

func (p *IpaFlavorService) Apply(lang, word string) (ret string) {

	for k, v := range (*p.mapping)[lang] {
		if k == v {
			log.Now().Errorf("Ipa flavour %s does have identical source and dest string: %s to %s", lang, k, v)
		}

		word = strings.ReplaceAll(word, k, v)
	}

	return word
}

func NewIpaFlavorService(di *DependencyInjection) *IpaFlavorService {

	mapping := MustAny[interfaces.IpaFlavor](di).GetIpaFlavors()

	return &IpaFlavorService{
		mapping: &mapping,
	}
}

var _ IIpaFlavorService = &IpaFlavorService{}
