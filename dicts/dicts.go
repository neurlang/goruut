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
import "github.com/neurlang/goruut/dicts/luxembourgish"
import "github.com/neurlang/goruut/dicts/dutch"
import "github.com/neurlang/goruut/dicts/portuguese"
import "github.com/neurlang/goruut/dicts/russian"
import "github.com/neurlang/goruut/dicts/swedish"
import "github.com/neurlang/goruut/dicts/romanian"
import "github.com/neurlang/goruut/dicts/finnish"
import "github.com/neurlang/goruut/dicts/isan"
import "github.com/neurlang/goruut/dicts/swahili"
import "github.com/neurlang/goruut/dicts/esperanto"
import "github.com/neurlang/goruut/dicts/icelandic"
import "github.com/neurlang/goruut/dicts/norwegian"
import "github.com/neurlang/goruut/dicts/jamaican"
import "github.com/neurlang/goruut/dicts/japanese"
import "github.com/neurlang/goruut/dicts/hindi"
import "github.com/neurlang/goruut/dicts/bengali"
import "github.com/neurlang/goruut/dicts/bengali/dhaka"
import "github.com/neurlang/goruut/dicts/bengali/rahr"
import "github.com/neurlang/goruut/dicts/punjabi"
import "github.com/neurlang/goruut/dicts/telugu"
import "github.com/neurlang/goruut/dicts/marathi"
import "github.com/neurlang/goruut/dicts/chinese/mandarin"
import "github.com/neurlang/goruut/dicts/tamil"
import "github.com/neurlang/goruut/dicts/gujarati"
import "github.com/neurlang/goruut/dicts/urdu"
import "github.com/neurlang/goruut/dicts/turkish"
import "errors"

var ErrUnsupportedLanguage = errors.New("unsupportedLang")

type DictGetter struct{}

func (DictGetter) GetDict(lang, filename string) ([]byte, error) {
	return GetDict(lang, filename)
}

func lzw(model string) string {
	if model == "weights0.json.gz" {
		return "weights1.json.lzw"
	}
	return model
}

func (DictGetter) IsOldFormat(magic []byte) bool {
	if len(magic) < 2 {
		return false
	}
	// GZIP
	return magic[0] == 0x1F && magic[1] == 0x8B
}

func (DictGetter) IsNewFormat(magic []byte) bool {
	if len(magic) < 2 {
		return false
	}
	// LZW
	return (magic[0] == 0x1F && magic[1] == 0x9D) || (magic[0] == 0x1F && magic[1] == 0xA0)
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
	case "Luxembourgish":
		return luxembourgish.Language.ReadFile(filename)
	case "Dutch":
		return dutch.Language.ReadFile(filename)
	case "Portuguese":
		return portuguese.Language.ReadFile(filename)
	case "Russian":
		return russian.Language.ReadFile(filename)
	case "Swedish":
		return swedish.Language.ReadFile(filename)
	case "Romanian":
		return romanian.Language.ReadFile(lzw(filename))
	case "Finnish":
		return finnish.Language.ReadFile(lzw(filename))
	case "Isan":
		return isan.Language.ReadFile(lzw(filename))
	case "Swahili":
		return swahili.Language.ReadFile(lzw(filename))
	case "Esperanto":
		return esperanto.Language.ReadFile(lzw(filename))
	case "Icelandic":
		return icelandic.Language.ReadFile(lzw(filename))
	case "Norwegian":
		return norwegian.Language.ReadFile(lzw(filename))
	case "Jamaican":
		return jamaican.Language.ReadFile(lzw(filename))
	case "Japanese":
		return japanese.Language.ReadFile(lzw(filename))
	case "Hindi":
		return hindi.Language.ReadFile(lzw(filename))
	case "Bengali":
		return bengali.Language.ReadFile(lzw(filename))
	case "BengaliDhaka":
		return dhaka.Language.ReadFile(lzw(filename))
	case "BengaliRahr":
		return rahr.Language.ReadFile(lzw(filename))
	case "Punjabi":
		return punjabi.Language.ReadFile(lzw(filename))
	case "Telugu":
		return telugu.Language.ReadFile(lzw(filename))
	case "Marathi":
		return marathi.Language.ReadFile(lzw(filename))
	case "ChineseMandarin":
		return mandarin.Language.ReadFile(lzw(filename))
	case "Tamil":
		return tamil.Language.ReadFile(lzw(filename))
	case "Gujarati":
		return gujarati.Language.ReadFile(lzw(filename))
	case "Urdu":
		return urdu.Language.ReadFile(lzw(filename))
	case "Turkish":
		return turkish.Language.ReadFile(lzw(filename))
	default:
		return nil, ErrUnsupportedLanguage
	}
}
