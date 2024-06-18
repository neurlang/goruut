package repo

import (
	"encoding/json"
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/repo/interfaces"
	"regexp"
)
import . "github.com/martinarisk/di/dependency_injection"

type ISpacerSplitterRepository interface {
	Split(string, string) []string
}
type SpacerSplitterRepository struct {
	getter *interfaces.DictGetter
	repo ISpacerSplitterRepository
	lang *spacerlanguages
	steps  *interfaces.SpacerSplitterSteps
}
type spacerlanguages map[string]*spacersplitterlanguage

func (l *spacerlanguages) Len(lang string) int {
	return len((*l)[lang].steps)
}
func (l *spacerlanguages) LeftRegexp(lang string, n int) *regexp.Regexp {
	return (*l)[lang].steps[n][0]
}
func (l *spacerlanguages) RightRegexp(lang string, n int) *regexp.Regexp {
	return (*l)[lang].steps[n][1]
}

type spacerlanguage struct {
	Spacer []struct {
		LeftRegexp string
		RightRegexp string
	}
}

type spacersplitterlanguage struct {
	steps [][2]*regexp.Regexp
}

func (p *SpacerSplitterRepository) LoadLanguage(lang string) {
	var language_files = []string{"spacer.json"}
	for _, file := range language_files {
		log.Now().Debugf("Language %s loading file", file)
		data := log.Error1((*p.getter).GetDict(lang, file))

		// Parse the JSON data into the Language struct
		var langone spacerlanguage
		err := json.Unmarshal(data, &langone)
		if err != nil {
			log.Now().Errorf("Error parsing JSON: %v\n", err)
			continue
		}

		var splitterlangone spacersplitterlanguage
		for _, v := range langone.Spacer {
			splitterlangone.steps = append(splitterlangone.steps, [2]*regexp.Regexp{
				log.Error1(regexp.Compile(v.LeftRegexp)),
				log.Error1(regexp.Compile(v.RightRegexp)),
			})
		}

		(*p.lang)[lang] = &splitterlangone

		iface := (interfaces.SpacerSplitterSteps)(&(*p.lang))

		p.steps = &iface
	}
}

func (s *SpacerSplitterRepository) Split(lang, sentence string) []string {
	s.LoadLanguage(lang)
	if s.steps == nil {
		return []string{sentence}
	}
	steps := *s.steps

	var all = []rune(sentence)
	var output []string

	var q = 0
	for i := range all {
		left, right := string(all[q:i]), string(all[i:])
		for j := 0; j < steps.Len(lang); j++ {
			l := steps.LeftRegexp(lang, j)
			r := steps.RightRegexp(lang, j)
			lmatch := l != nil && l.MatchString(left) || l == nil
			rmatch := r != nil && r.MatchString(right) || r == nil
			if lmatch && rmatch {
				output = append(output, left)
				q = i
			}
		}
	}


	return output
}

func NewSpacerSplitterRepository(di *DependencyInjection) *SpacerSplitterRepository {
	getter := MustAny[interfaces.DictGetter](di)
	langs := make(spacerlanguages)

	return &SpacerSplitterRepository{
		getter: &getter,
		lang:   &langs,
	}
}

var _ ISpacerSplitterRepository = &SpacerSplitterRepository{}
