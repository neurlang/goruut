package repo

import (
	"encoding/json"
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/repo/interfaces"
	"strings"
	"sync"
	"regexp"
)
import . "github.com/martinarisk/di/dependency_injection"

type ISpaceSplitterRepository interface {
	Split(string) []string
	SplitLang(bool, string, string) []string
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

	SplitAt map[string]string `json:"SplitAt"`
	splitAt []*regexp.Regexp
	splitBy []string
}

func (l *spacesplitlanguage) load() {
	for exp, splitStr := range l.SplitAt {
		l.splitAt = append(l.splitAt, log.Error1(regexp.Compile(exp)))
		l.splitBy = append(l.splitBy, splitStr)
	}
}

func (s *SpaceSplitterRepository) Split(sentence string) []string {
	return s.SplitLang(false, "", sentence)
}
func (s *SpaceSplitterRepository) SplitLang(isReverse bool, lang, sentence string) []string {
	s.LoadLanguage(isReverse, lang)
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	s.mut.RLock()
	language := (*s.lang)[lang+reverse]
	var splitAfter, splitBefore, splitBy []string
	var splitAt []*regexp.Regexp
	if language != nil {
		splitAfter = language.SplitAfter
		splitBefore = language.SplitBefore
		splitBy = language.splitBy
		splitAt = language.splitAt
	}
	s.mut.RUnlock()

	for i, re := range splitAt {
		sentence = re.ReplaceAllString(sentence, splitBy[i])
	}

	for _, v := range splitAfter {
		sentence = strings.ReplaceAll(sentence, v, v+" ")
	}
	for _, v := range splitBefore {
		sentence = strings.ReplaceAll(sentence, v, " "+v)
	}

	return strings.Fields(sentence)
}

func (p *SpaceSplitterRepository) LoadLanguage(isReverse bool, lang string) {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}

	p.mut.RLock()
	existing_lang := (*p.lang)[lang+reverse]
	p.mut.RUnlock()

	if existing_lang != nil {
		return
	}

	var language_files = []string{"language" + reverse + ".json"}
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
		langone.load()
		p.mut.Lock()
		(*p.lang)[lang+reverse] = &langone
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
