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
import "github.com/neurlang/goruut/dicts/vietnamese/southern"
import "github.com/neurlang/goruut/dicts/vietnamese/central"
import "github.com/neurlang/goruut/dicts/vietnamese/northern"
import "github.com/neurlang/goruut/dicts/polish"
import "github.com/neurlang/goruut/dicts/greek"
import "github.com/neurlang/goruut/dicts/ukrainian"
import "github.com/neurlang/goruut/dicts/hungarian"
import "github.com/neurlang/goruut/dicts/malay/arab"
import "github.com/neurlang/goruut/dicts/malay/latin"
import "github.com/neurlang/goruut/dicts/korean"
import "github.com/neurlang/goruut/dicts/kazakh"
import "github.com/neurlang/goruut/dicts/afrikaans"
import "github.com/neurlang/goruut/dicts/azerbaijani"
import "github.com/neurlang/goruut/dicts/cebuano"
import "github.com/neurlang/goruut/dicts/hausa"
import "github.com/neurlang/goruut/dicts/indonesian"
import "github.com/neurlang/goruut/dicts/danish"
import "github.com/neurlang/goruut/dicts/malayalam"
import "github.com/neurlang/goruut/dicts/javanese"
import "github.com/neurlang/goruut/dicts/macedonian"
import "github.com/neurlang/goruut/dicts/hebrew"
import "github.com/neurlang/goruut/dicts/amharic"
import "github.com/neurlang/goruut/dicts/belarusian"
import "github.com/neurlang/goruut/dicts/chechen"
import "github.com/neurlang/goruut/dicts/dzongkha"
import "github.com/neurlang/goruut/dicts/burmese"
import "github.com/neurlang/goruut/dicts/maltese"
import "github.com/neurlang/goruut/dicts/mongolian"
import "github.com/neurlang/goruut/dicts/nepali"
import "github.com/neurlang/goruut/dicts/pashto"
import "github.com/neurlang/goruut/dicts/tibetan"
import "github.com/neurlang/goruut/dicts/uyghur"
import "github.com/neurlang/goruut/dicts/thai"
import "github.com/neurlang/goruut/dicts/zulu"
import "github.com/neurlang/goruut/dicts/catalan"
import "github.com/neurlang/goruut/dicts/armenian"
import "github.com/neurlang/goruut/dicts/croatian"
import "github.com/neurlang/goruut/dicts/serbian"
import "github.com/neurlang/goruut/dicts/bulgarian"
import "github.com/neurlang/goruut/dicts/chichewa"
import "github.com/neurlang/goruut/dicts/estonian"
import "github.com/neurlang/goruut/dicts/georgian"
import "github.com/neurlang/goruut/dicts/latvian"
import "github.com/neurlang/goruut/dicts/lithuanian"
import "github.com/neurlang/goruut/dicts/tagalog"
import "github.com/neurlang/goruut/dicts/yoruba"
import "github.com/neurlang/goruut/dicts/basque"
import "github.com/neurlang/goruut/dicts/galician"
import khmer "github.com/neurlang/goruut/dicts/khmer/central"
import "github.com/neurlang/goruut/dicts/lao"
import "github.com/neurlang/goruut/dicts/english/american"
import "github.com/neurlang/goruut/dicts/english/british"
import "errors"

var ErrUnsupportedLanguage = errors.New("unsupportedLang")

type DictGetter struct{}

func (DictGetter) GetDict(lang, filename string) ([]byte, error) {
	return GetDict(lang, filename)
}

// XXX: Doesn't work
func (DictGetter) IsOldFormat(magic []byte) bool {
	if len(magic) < 2 {
		return false
	}
	println(magic[0], magic[1])
	// LZW
	return (magic[0] == 0x1F && magic[1] == 0x9D) || (magic[0] == 0x1F && magic[1] == 0xA0)
}

func (DictGetter) IsNewFormat(magic []byte) bool {
	if len(magic) < 2 {
		return false
	}
	// ZLIB
	return (magic[0] == 0x78 && (magic[1] == 0x01 || magic[1] == 0x5E || magic[1] == 0x9C || magic[1] == 0xDA))
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
		return romanian.Language.ReadFile(filename)
	case "Finnish":
		return finnish.Language.ReadFile(filename)
	case "Isan":
		return isan.Language.ReadFile(filename)
	case "Swahili":
		return swahili.Language.ReadFile(filename)
	case "Esperanto":
		return esperanto.Language.ReadFile(filename)
	case "Icelandic":
		return icelandic.Language.ReadFile(filename)
	case "Norwegian":
		return norwegian.Language.ReadFile(filename)
	case "Jamaican":
		return jamaican.Language.ReadFile(filename)
	case "Japanese":
		return japanese.Language.ReadFile(filename)
	case "Hindi":
		return hindi.Language.ReadFile(filename)
	case "Bengali":
		return bengali.Language.ReadFile(filename)
	case "BengaliDhaka":
		return dhaka.Language.ReadFile(filename)
	case "BengaliRahr":
		return rahr.Language.ReadFile(filename)
	case "Punjabi":
		return punjabi.Language.ReadFile(filename)
	case "Telugu":
		return telugu.Language.ReadFile(filename)
	case "Marathi":
		return marathi.Language.ReadFile(filename)
	case "ChineseMandarin":
		return mandarin.Language.ReadFile(filename)
	case "Tamil":
		return tamil.Language.ReadFile(filename)
	case "Gujarati":
		return gujarati.Language.ReadFile(filename)
	case "Urdu":
		return urdu.Language.ReadFile(filename)
	case "Turkish":
		return turkish.Language.ReadFile(filename)
	case "VietnameseSouthern":
		return southern.Language.ReadFile(filename)
	case "VietnameseCentral":
		return central.Language.ReadFile(filename)
	case "VietnameseNorthern":
		return northern.Language.ReadFile(filename)
	case "Polish":
		return polish.Language.ReadFile(filename)
	case "Greek":
		return greek.Language.ReadFile(filename)
	case "Ukrainian":
		return ukrainian.Language.ReadFile(filename)
	case "Hungarian":
		return hungarian.Language.ReadFile(filename)
	case "MalayLatin":
		return latin.Language.ReadFile(filename)
	case "MalayArab":
		return arab.Language.ReadFile(filename)
	case "Korean":
		return korean.Language.ReadFile(filename)
	case "Kazakh":
		return kazakh.Language.ReadFile(filename)
	case "Afrikaans":
		return afrikaans.Language.ReadFile(filename)
	case "Azerbaijani":
		return azerbaijani.Language.ReadFile(filename)
	case "Cebuano":
		return cebuano.Language.ReadFile(filename)
	case "Hausa":
		return hausa.Language.ReadFile(filename)
	case "Indonesian":
		return indonesian.Language.ReadFile(filename)
	case "Danish":
		return danish.Language.ReadFile(filename)
	case "Malayalam":
		return malayalam.Language.ReadFile(filename)
	case "Javanese":
		return javanese.Language.ReadFile(filename)
	case "Macedonian":
		return macedonian.Language.ReadFile(filename)
	case "Hebrew":
		return hebrew.Language.ReadFile(filename)
	case "Amharic":
		return amharic.Language.ReadFile(filename)
	case "Belarusian":
		return belarusian.Language.ReadFile(filename)
	case "Chechen":
		return chechen.Language.ReadFile(filename)
	case "Dzongkha":
		return dzongkha.Language.ReadFile(filename)
	case "Burmese":
		return burmese.Language.ReadFile(filename)
	case "Maltese":
		return maltese.Language.ReadFile(filename)
	case "Mongolian":
		return mongolian.Language.ReadFile(filename)
	case "Nepali":
		return nepali.Language.ReadFile(filename)
	case "Pashto":
		return pashto.Language.ReadFile(filename)
	case "Tibetan":
		return tibetan.Language.ReadFile(filename)
	case "Uyghur":
		return uyghur.Language.ReadFile(filename)
	case "Thai":
		return thai.Language.ReadFile(filename)
	case "Zulu":
		return zulu.Language.ReadFile(filename)
	case "Catalan":
		return catalan.Language.ReadFile(filename)
	case "Armenian":
		return armenian.Language.ReadFile(filename)
	case "Croatian":
		return croatian.Language.ReadFile(filename)
	case "Serbian":
		return serbian.Language.ReadFile(filename)
	case "Bulgarian":
		return bulgarian.Language.ReadFile(filename)
	case "Chichewa":
		return chichewa.Language.ReadFile(filename)
	case "Estonian":
		return estonian.Language.ReadFile(filename)
	case "Georgian":
		return georgian.Language.ReadFile(filename)
	case "Latvian":
		return latvian.Language.ReadFile(filename)
	case "Lithuanian":
		return lithuanian.Language.ReadFile(filename)
	case "Tagalog":
		return tagalog.Language.ReadFile(filename)
	case "Yoruba":
		return yoruba.Language.ReadFile(filename)
	case "Basque":
		return basque.Language.ReadFile(filename)
	case "Galician":
		return galician.Language.ReadFile(filename)
	case "KhmerCentral":
		return khmer.Language.ReadFile(filename)
	case "Lao":
		return lao.Language.ReadFile(filename)
	case "EnglishAmerican":
		return american.Language.ReadFile(filename)
	case "EnglishBritish":
		return british.Language.ReadFile(filename)
	default:
		return nil, ErrUnsupportedLanguage
	}
}

func LangName(dir string) string {
	switch dir {
	case "afrikaans": return "Afrikaans";
	case "amharic": return "Amharic";
	case "arabic": return "Arabic";
	case "armenian": return "Armenian";
	case "azerbaijani": return "Azerbaijani";
	case "basque": return "Basque";
	case "belarusian": return "Belarusian";
	case "bengali": return "Bengali";
	case "bengali/dhaka": return "BengaliDhaka";
	case "bengali/rahr": return "BengaliRahr";
	case "bulgarian": return "Bulgarian";
	case "burmese": return "Burmese";
	case "catalan": return "Catalan";
	case "cebuano": return "Cebuano";
	case "chechen": return "Chechen";
	case "chichewa": return "Chichewa";
	case "chinese/mandarin": return "ChineseMandarin";
	case "croatian": return "Croatian";
	case "czech": return "Czech";
	case "danish": return "Danish";
	case "dutch": return "Dutch";
	case "dzongkha": return "Dzongkha";
	case "english": return "English";
	case "english/american": return "EnglishAmerican";
	case "english/british": return "EnglishBritish";
	case "esperanto": return "Esperanto";
	case "estonian": return "Estonian";
	case "farsi": return "Farsi";
	case "finnish": return "Finnish";
	case "french": return "French";
	case "galician": return "Galician";
	case "georgian": return "Georgian";
	case "german": return "German";
	case "greek": return "Greek";
	case "gujarati": return "Gujarati";
	case "hausa": return "Hausa";
	case "hebrew": return "Hebrew";
	case "hindi": return "Hindi";
	case "hungarian": return "Hungarian";
	case "icelandic": return "Icelandic";
	case "indonesian": return "Indonesian";
	case "isan": return "Isan";
	case "italian": return "Italian";
	case "jamaican": return "Jamaican";
	case "japanese": return "Japanese";
	case "javanese": return "Javanese";
	case "kazakh": return "Kazakh";
	case "khmer/central": return "KhmerCentral";
	case "korean": return "Korean";
	case "lao": return "Lao";
	case "latvian": return "Latvian";
	case "lithuanian": return "Lithuanian";
	case "luxembourgish": return "Luxembourgish";
	case "macedonian": return "Macedonian";
	case "malay/arab": return "Malayalam";
	case "malay/latin": return "MalayArab";
	case "malayalam": return "MalayLatin";
	case "maltese": return "Maltese";
	case "marathi": return "Marathi";
	case "mongolian": return "Mongolian";
	case "nepali": return "Nepali";
	case "norwegian": return "Norwegian";
	case "pashto": return "Pashto";
	case "polish": return "Polish";
	case "portuguese": return "Portuguese";
	case "punjabi": return "Punjabi";
	case "romanian": return "Romanian";
	case "russian": return "Russian";
	case "serbian": return "Serbian";
	case "slovak": return "Slovak";
	case "spanish": return "Spanish";
	case "swahili": return "Swahili";
	case "swedish": return "Swedish";
	case "tagalog": return "Tagalog";
	case "tamil": return "Tamil";
	case "telugu": return "Telugu";
	case "thai": return "Thai";
	case "tibetan": return "Tibetan";
	case "turkish": return "Turkish";
	case "ukrainian": return "Ukrainian";
	case "urdu": return "Urdu";
	case "uyghur": return "Uyghur";
	case "vietnamese/central": return "VietnameseCentral";
	case "vietnamese/northern": return "VietnameseNorthern";
	case "vietnamese/southern": return "VietnameseSouthern";
	case "yoruba": return "Yoruba";
	case "zulu": return "Zulu";
	default: return "";
	}
}
