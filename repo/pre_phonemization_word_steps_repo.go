package repo

import (
	"encoding/json"
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/repo/interfaces"
	"golang.org/x/text/unicode/norm"
	"regexp"
	"sort"
	"strings"
	"sync"
)
import . "github.com/martinarisk/di/dependency_injection"

type IPrePhonWordStepsRepository interface {
	PrePhonemizeWord(isReverse bool, lang string, word string) string
}
type PrePhonWordStepsRepository struct {
	getter *interfaces.DictGetter

	mut   sync.RWMutex
	lang  *prephonlanguages
	steps *interfaces.PrePhonemizationSteps
}

type prephonlanguages struct {
	mut  sync.RWMutex
	lang map[string]*prephonlanguage
}

type prephonlanguage struct {
	PrePhonWordSteps []PrePhonWordStep `json:"PrePhonWordSteps"`
}

type PrePhonWordStep struct {
	Normalize string `json:"Normalize"`
	Trim      string `json:"Trim"`
	ToLower   bool   `json:"ToLower"`
}

func (p *prephonlanguages) Len(isReverse bool, lang string) int {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	p.mut.RLock()
	defer p.mut.RUnlock()
	if p.lang == nil || (p.lang)[lang+reverse] == nil {
		return 0
	}
	return len((p.lang)[lang+reverse].PrePhonWordSteps)
}
func (p *prephonlanguages) IsNormalize(isReverse bool, lang string, n int) bool {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	p.mut.RLock()
	defer p.mut.RUnlock()
	if p.lang == nil || (p.lang)[lang+reverse] == nil {
		return false
	}
	return len((p.lang)[lang+reverse].PrePhonWordSteps[n].Normalize) > 0
}
func (p *prephonlanguages) IsTrim(isReverse bool, lang string, n int) bool {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	p.mut.RLock()
	defer p.mut.RUnlock()
	if p.lang == nil || (p.lang)[lang+reverse] == nil {
		return false
	}
	return len((p.lang)[lang+reverse].PrePhonWordSteps[n].Trim) > 0
}
func (p *prephonlanguages) IsToLower(isReverse bool, lang string, n int) bool {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	p.mut.RLock()
	defer p.mut.RUnlock()
	if p.lang == nil || (p.lang)[lang+reverse] == nil {
		return false
	}
	return (p.lang)[lang+reverse].PrePhonWordSteps[n].ToLower
}
func (p *prephonlanguages) GetNormalize(isReverse bool, lang string, n int) string {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	p.mut.RLock()
	defer p.mut.RUnlock()
	return (p.lang)[lang+reverse].PrePhonWordSteps[n].Normalize
}
func (p *prephonlanguages) GetTrim(isReverse bool, lang string, n int) string {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	p.mut.RLock()
	defer p.mut.RUnlock()
	return (p.lang)[lang+reverse].PrePhonWordSteps[n].Trim
}
func (p *PrePhonWordStepsRepository) LoadLanguage(isReverse bool, lang string) {

	var reverse string
	if isReverse {
		reverse = "_reverse"
	}

	p.mut.Lock()
	defer p.mut.Unlock()

	existing_lang := (p.lang.lang)[lang+reverse]

	if existing_lang != nil {
		return
	}

	var language_files = []string{"language" + reverse + ".json"}
	for _, file := range language_files {
		log.Now().Debugf("Language %s loading file", file)
		data := log.Error1((*p.getter).GetDict(lang, file))

		// Parse the JSON data into the Language struct
		var langone prephonlanguage
		err := json.Unmarshal(data, &langone)
		if err != nil {
			log.Now().Errorf("Error parsing JSON: %v\n", err)
			continue
		}
		p.lang.lang = make(map[string]*prephonlanguage)
		(p.lang.lang)[lang+reverse] = &langone

		iface := (interfaces.PrePhonemizationSteps)((p.lang))

		p.steps = &iface
	}
}

// Function to sort combiners runs
func SortCombiningRuns(s string) string {
	re := regexp.MustCompile(`[\p{Mn}\p{Mc}]+`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		runes := []rune(match)
		sort.Slice(runes, func(i, j int) bool { return runes[i] > runes[j] })
		return string(runes)
	})
}

// Function to normalize text to NFC
func NormalizeTo(input, form string) string {
	var format norm.Form

	switch form {
	case "NFC":
		format = norm.NFC
	case "NFD":
		format = norm.NFD
	case "NFKC":
		format = norm.NFKC
	case "NFKD":
		format = norm.NFKD
	case "Sort":
		return SortCombiningRuns(input)
	default:
		return input
	}

	return format.String(input)
}

func (s *PrePhonWordStepsRepository) PrePhonemizeWord(isReverse bool, lang string, word string) string {
	s.LoadLanguage(isReverse, lang)

	s.mut.RLock()
	if s.steps == nil {
		s.mut.RUnlock()
		return word
	}
	steps := *s.steps
	length := steps.Len(isReverse, lang)
	s.mut.RUnlock()
	withMutexBool := func(f func(isReverse bool, lang string, i int) bool, isReverse bool, lang string, i int) bool {
		s.mut.RLock()
		ret := f(isReverse, lang, i)
		s.mut.RUnlock()
		return ret
	}

	for i := 0; i < length; i++ {

		if withMutexBool(steps.IsNormalize, isReverse, lang, i) {
			word = NormalizeTo(word, steps.GetNormalize(isReverse, lang, i))
		}
		if withMutexBool(steps.IsTrim, isReverse, lang, i) {
			word = strings.Trim(word, steps.GetTrim(isReverse, lang, i))
		}
		if withMutexBool(steps.IsToLower, isReverse, lang, i) {
			word = strings.ToLower(word)
		}

	}

	return word
}

func NewPrePhonWordStepsRepository(di *DependencyInjection) *PrePhonWordStepsRepository {
	getter := MustAny[interfaces.DictGetter](di)
	langs := prephonlanguages{}

	return &PrePhonWordStepsRepository{
		getter: &getter,
		lang:   &langs,
	}
}

var _ IPrePhonWordStepsRepository = &PrePhonWordStepsRepository{}
