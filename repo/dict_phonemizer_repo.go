package repo

import (
	"bytes"
	"encoding/csv"
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/repo/interfaces"
	"github.com/spaolacci/murmur3"
	"strings"
)
import . "github.com/martinarisk/di/dependency_injection"

type IDictPhonemizerRepository interface {
	LookupWords(isReverse bool, lang string, word string) []map[uint64]string
}
type DictPhonemizerRepository struct {
	getter     *interfaces.DictGetter
	lang_words *map[string]map[string]map[uint64]string
}

func murmur3hash(str string) uint64 {
	return murmur3.Sum64WithSeed([]byte(str), 0)
}

func (r *DictPhonemizerRepository) LoadLanguage(isReverse bool, lang string) {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	if (*r.lang_words)[lang+reverse] == nil {
		(*r.lang_words)[lang+reverse] = make(map[string]map[uint64]string)
	} else {
		log.Now().Debugf("Language %s already loaded", lang)
		return
	}

	var files = []string{"missing" + reverse + ".tsv"}

	for _, file := range files {
		clean := log.Error1((*r.getter).GetDict(lang, file))

		reader := csv.NewReader(bytes.NewReader(clean))

		reader.Comma = '\t'

		recs := log.Error1(reader.ReadAll())
		log.Now().Debugf("Language %s has %d records", lang, len(recs))
		for _, v := range recs {
			for i := range v {
				v[i] = strings.ReplaceAll(v[i], " ", "")
			}
			var src, dst string
			var tag uint64
			if len(v) == 2 {
				src = v[0]
				dst = v[1]
				tag = murmur3hash(src + "\x00" + dst)
			} else if len(v) == 3 {
				src = v[0]
				tag = murmur3hash(v[1])
				dst = v[2]
			} else {
				log.Now().Debugf("Language %s has wrong number of columns: %d", src, len(v))
				continue
			}
			if (*r.lang_words)[lang+reverse][src] == nil {
				(*r.lang_words)[lang+reverse][src] = make(map[uint64]string)
			}
			(*r.lang_words)[lang+reverse][src][tag] = dst
		}
	}
}

func (r *DictPhonemizerRepository) LookupWords(isReverse bool, lang, word string) (ret []map[uint64]string) {
	r.LoadLanguage(isReverse, lang)
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	found := (*r.lang_words)[lang+reverse][word]

	if len(found) == 0 {
		return nil
	}

	for k, v := range found {
		m := make(map[uint64]string)
		m[0] = word
		m[k] = v
		ret = append(ret, m)
	}
	return
}

func NewDictPhonemizerRepository(di *DependencyInjection) *DictPhonemizerRepository {
	getter := MustAny[interfaces.DictGetter](di)
	mapping := make(map[string]map[string]map[uint64]string)

	return &DictPhonemizerRepository{
		getter:     &getter,
		lang_words: &mapping,
	}
}

var _ IDictPhonemizerRepository = &DictPhonemizerRepository{}
