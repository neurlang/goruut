package repo

import (
	"github.com/neurlang/NumToWordsGo/NumToWords"
	"github.com/neurlang/goruut/helpers/log"
	"strconv"
	"strings"
)
import . "github.com/martinarisk/di/dependency_injection"

type INumToWordsRepository interface {
	ExpandNumericWord(isReverse bool, lang, word string, languages []string) []map[string]uint32
}
type NumToWordsRepository struct {
}

func expandNumericWord(word, lang string) (ret []map[string]uint32) {
	num, err := strconv.Atoi(word)
	if err != nil {
		log.Now().Debugf("%e", err)
		return nil
	}

	var sentence string
	switch lang {
	case "Arabic":
		sentence = log.Error1(NumToWords.Convert(num, "ar"))
	case "Czech":
		sentence = log.Error1(NumToWords.Convert(num, "cs"))
	case "German":
		sentence = log.Error1(NumToWords.Convert(num, "de"))
	case "Spanish":
		sentence = log.Error1(NumToWords.Convert(num, "es"))
	case "French":
		sentence = log.Error1(NumToWords.Convert(num, "fr"))
	case "Hungarian":
		sentence = log.Error1(NumToWords.Convert(num, "hu"))
	case "Polish":
		sentence = log.Error1(NumToWords.Convert(num, "pl"))
	case "Russian":
		sentence = log.Error1(NumToWords.Convert(num, "ru"))
	case "Slovak":
		sentence = log.Error1(NumToWords.Convert(num, "sk"))
	case "Ukrainian":
		sentence = log.Error1(NumToWords.Convert(num, "uk"))
	case "English", "EnglishAmerican", "EnglishBritish":
		sentence = log.Error1(NumToWords.Convert(num, "en"))
	default:
		return nil
	}
	log.Now().Debugf("Num: %d Output: %s", num, sentence)
	fields := strings.Fields(sentence)
	for _, field := range fields {
		log.Now().Debugf("Field: %s", field)
		var mapping = make(map[string]uint32)
		mapping[field] = 0
		ret = append(ret, mapping)
	}
	return
}

func (n *NumToWordsRepository) ExpandNumericWord(isReverse bool, lang, word string, languages []string) (ret []map[string]uint32) {

	if isReverse {
		return nil
	}

	ret = expandNumericWord(word, lang)
	if ret != nil {
		return ret
	}

	for _, lang := range languages {
		ret = expandNumericWord(word, lang)
		if ret != nil {
			return ret
		}
	}

	return nil
}

func NewNumToWordsRepository(di *DependencyInjection) *NumToWordsRepository {

	return &NumToWordsRepository{
	}
}

var _ INumToWordsRepository = &NumToWordsRepository{}
