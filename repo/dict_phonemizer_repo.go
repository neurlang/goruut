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
	PhonemizeWord(lang, word string) (ret map[uint64]string)
}
type DictPhonemizerRepository struct {
	getter     *interfaces.DictGetter
	lang_words *map[string]map[string]map[uint64]string
}

func murmur3hash(str string) uint64 {
	return murmur3.Sum64WithSeed([]byte(str), 0)
}

func (r *DictPhonemizerRepository) LoadLanguage(lang string) {
	if (*r.lang_words)[lang] == nil {
		(*r.lang_words)[lang] = make(map[string]map[uint64]string)
	} else {
		log.Now().Debugf("Language %s already loaded", lang)
		return
	}

	var files = []string{"clean.tsv", "missing.tsv"}

	for _, file := range files {
		clean := log.Error1((*r.getter).GetDict(lang, file))

		clean = bytes.ReplaceAll(clean, []byte("\t"), []byte(","))

		reader := csv.NewReader(bytes.NewReader(clean))

		recs := log.Error1(reader.ReadAll())
		log.Now().Debugf("Language %s has %d records", lang, len(recs))
		for _, v := range recs {
			v[0] = strings.ReplaceAll(v[0], " ", "")
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
			dst = strings.ReplaceAll(dst, " ", "")
			if (*r.lang_words)[lang][src] == nil {
				(*r.lang_words)[lang][src] = make(map[uint64]string)
			}
			(*r.lang_words)[lang][src][tag] = dst
		}
	}
}

func (r *DictPhonemizerRepository) PhonemizeWord(lang, word string) (ret map[uint64]string) {
	r.LoadLanguage(lang)

	ret = (*r.lang_words)[lang][word]

	if len(ret) == 0 {
		return nil
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
