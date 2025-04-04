// Package services encapsulates business logic and domain operations.
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
	longest *map[string]int
}

func (p *IpaFlavorService) Apply(lang, word string) (ret string) {

	for i := (*p.longest)[lang]; i > 0; i-- {
		for k, v := range (*p.mapping)[lang] {
			if len(k) != i {
				continue
			}

			if k == v {
				log.Now().Errorf("Ipa flavour %s does have identical source and dest string: %s to %s", lang, k, v)
			}

			word = strings.ReplaceAll(word, k, v)
		}
	}
	return word
}

func NewIpaFlavorService(di *DependencyInjection) *IpaFlavorService {

	mapping := MustAny[interfaces.IpaFlavor](di).GetIpaFlavors()
	longest := make(map[string]int)

	for lang, dict := range mapping {
		for k := range dict {
			if len(k) > longest[lang] {
				longest[lang] = len(k)
			}
		}
	}

	return &IpaFlavorService{
		mapping: &mapping,
		longest: &longest,
	}
}

var _ IIpaFlavorService = &IpaFlavorService{}
