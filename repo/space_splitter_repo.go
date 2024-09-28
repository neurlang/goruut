package repo

import (
	"encoding/json"
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/repo/interfaces"
	"strings"
	"sync"
)
import . "github.com/martinarisk/di/dependency_injection"

type ISpaceSplitterRepository interface {
	Split(string) []string
	SplitLang(string, string) []string
}
type SpaceSplitterRepository struct {
	getter *interfaces.DictGetter

	mut  sync.RWMutex
	lang *spacesplitlanguages
}

type spacesplitlanguages map[string]*spacesplitlanguage

type spacesplitlanguage struct {
	SplitAfter  []string `json:"SplitAfter"`
	SplitBefore []string `json:"SplitBefore"`
}

func (s *SpaceSplitterRepository) Split(sentence string) []string {
	return s.SplitLang("", sentence)
}
func (s *SpaceSplitterRepository) SplitLang(lang, sentence string) []string {
	s.LoadLanguage(lang)
	s.mut.RLock()
	language := (*s.lang)[lang]
	var splitAfter, splitBefore []string
	if language != nil {
		splitAfter = language.SplitAfter
		splitBefore = language.SplitBefore
	}
	s.mut.RUnlock()

	for _, v := range splitAfter {
		sentence = strings.ReplaceAll(sentence, v, v+" ")
	}
	for _, v := range splitBefore {
		sentence = strings.ReplaceAll(sentence, v, " "+v)
	}

	return strings.Fields(sentence)
}

func (p *SpaceSplitterRepository) LoadLanguage(lang string) {

	p.mut.RLock()
	existing_lang := (*p.lang)[lang]
	p.mut.RUnlock()

	if existing_lang != nil {
		return
	}

	var language_files = []string{"language.json"}
	for _, file := range language_files {
		log.Now().Debugf("Language %s loading file", file)
		data := log.Error1((*p.getter).GetDict(lang, file))

		// Parse the JSON data into the Language struct
		var langone spacesplitlanguage
		err := json.Unmarshal(data, &langone)
		if err != nil {
			log.Now().Errorf("Error parsing JSON: %v\n", err)
			continue
		}
		p.mut.Lock()
		(*p.lang)[lang] = &langone
		p.mut.Unlock()
	}
}

func NewSpaceSplitterRepository(di *DependencyInjection) *SpaceSplitterRepository {
	getter := MustAny[interfaces.DictGetter](di)
	langs := make(spacesplitlanguages)
	return &SpaceSplitterRepository{
		getter: &getter,
		lang:   &langs,
	}
}

var _ ISpaceSplitterRepository = &SpaceSplitterRepository{}
