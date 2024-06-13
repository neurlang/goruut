package dicts

import "github.com/neurlang/goruut/dicts/czech"
import "github.com/neurlang/goruut/dicts/slovak"
import "github.com/neurlang/goruut/dicts/arabic"
import "github.com/neurlang/goruut/dicts/english"
import "errors"

var ErrUnsupportedLanguage = errors.New("unsupported")

type DictGetter struct{}

func (DictGetter) GetDict(lang, filename string) ([]byte, error) {
	return GetDict(lang, filename)
}

func GetDict(lang, filename string) ([]byte, error) {
	switch lang {
	case "Czech":
		return czech.Language.ReadFile(filename)
	case "Slovak":
		return slovak.Language.ReadFile(filename)
	case "Arabic":
		return arabic.Language.ReadFile(filename)
	case "English":
		return english.Language.ReadFile(filename)
	default:
		return nil, ErrUnsupportedLanguage
	}
}
