package dicts

import "github.com/neurlang/goruut/dicts/czech"
import "github.com/neurlang/goruut/dicts/spanish"
import "github.com/neurlang/goruut/dicts/slovak"
import "github.com/neurlang/goruut/dicts/arabic"
import "github.com/neurlang/goruut/dicts/farsi"
import "github.com/neurlang/goruut/dicts/english"
import "github.com/neurlang/goruut/dicts/german"
import "github.com/neurlang/goruut/dicts/french"
import "github.com/neurlang/goruut/dicts/italian"
import "errors"

var ErrUnsupportedLanguage = errors.New("unsupportedLang")

type DictGetter struct{}

func (DictGetter) GetDict(lang, filename string) ([]byte, error) {
	return GetDict(lang, filename)
}

func GetDict(lang, filename string) ([]byte, error) {
	switch lang {
	case "Czech":
		return czech.Language.ReadFile(filename)
	case "Spanish":
		return spanish.Language.ReadFile(filename)
	case "Slovak":
		return slovak.Language.ReadFile(filename)
	case "Arabic":
		return arabic.Language.ReadFile(filename)
	case "Farsi":
		return farsi.Language.ReadFile(filename)
	case "English":
		return english.Language.ReadFile(filename)
	case "German":
		return german.Language.ReadFile(filename)
	case "French":
		return french.Language.ReadFile(filename)
	case "Italian":
		return italian.Language.ReadFile(filename)
	default:
		return nil, ErrUnsupportedLanguage
	}
}
