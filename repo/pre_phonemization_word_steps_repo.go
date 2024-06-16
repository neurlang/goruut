package repo

import (
	"encoding/json"
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/repo/interfaces"
	"golang.org/x/text/unicode/norm"
	"strings"
)
import . "github.com/martinarisk/di/dependency_injection"

type IPrePhonWordStepsRepository interface {
	PrePhonemizeWord(string, string) string
}
type PrePhonWordStepsRepository struct {
	getter *interfaces.DictGetter
	lang   *prephonlanguages
	steps  *interfaces.PrePhonemizationSteps
}

type prephonlanguages map[string]*prephonlanguage

type prephonlanguage struct {
	PrePhonWordSteps []PrePhonWordStep `json:"PrePhonWordSteps"`
}

type PrePhonWordStep struct {
	Normalize string `json:"Normalize"`
	Trim      string `json:"Trim"`
	ToLower   bool   `json:"ToLower"`
}

func (p *prephonlanguages) Len(lang string) int {
	return len((*p)[lang].PrePhonWordSteps)
}
func (p *prephonlanguages) IsNormalize(lang string, n int) bool {
	return len((*p)[lang].PrePhonWordSteps[n].Normalize) > 0
}
func (p *prephonlanguages) IsTrim(lang string, n int) bool {
	return len((*p)[lang].PrePhonWordSteps[n].Trim) > 0
}
func (p *prephonlanguages) IsToLower(lang string, n int) bool {
	return (*p)[lang].PrePhonWordSteps[n].ToLower
}
func (p *prephonlanguages) GetNormalize(lang string, n int) string {
	return (*p)[lang].PrePhonWordSteps[n].Normalize
}
func (p *prephonlanguages) GetTrim(lang string, n int) string {
	return (*p)[lang].PrePhonWordSteps[n].Trim
}
func (p *PrePhonWordStepsRepository) LoadLanguage(lang string) {
	var language_files = []string{"language.json"}
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

		(*p.lang)[lang] = &langone

		iface := (interfaces.PrePhonemizationSteps)(&(*p.lang))

		p.steps = &iface
	}
}

// Function to normalize text to NFC
func NormalizeTo(input, form string) string {
	// Create a buffer to hold the normalized text
	buf := make([]byte, 0, len(input))

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
	default:
		return input
	}

	// Use the norm.NFC transformer to normalize the input text
	for i, r := range input {
		if format.QuickSpan([]byte(input[i:])) == len(input[i:]) {
			return input // Already NFC
		}
		buf = format.Append(buf, byte(r))
	}

	return string(buf)
}

func (s *PrePhonWordStepsRepository) PrePhonemizeWord(lang string, word string) string {
	s.LoadLanguage(lang)

	if s.steps == nil {
		return word
	}
	steps := *s.steps


	for i := 0; i < steps.Len(lang); i++ {
		if steps.IsNormalize(lang, i) {
			word = NormalizeTo(word, steps.GetNormalize(lang, i))
		}
		if steps.IsTrim(lang, i) {
			word = strings.Trim(word, steps.GetTrim(lang, i))
		}
		if steps.IsToLower(lang, i) {
			word = strings.ToLower(word)
		}
	}

	return word
}

func NewPrePhonWordStepsRepository(di *DependencyInjection) *PrePhonWordStepsRepository {
	getter := MustAny[interfaces.DictGetter](di)
	langs := make(prephonlanguages)

	return &PrePhonWordStepsRepository{
		getter: &getter,
		lang:   &langs,
	}
}

var _ IPrePhonWordStepsRepository = &PrePhonWordStepsRepository{}
