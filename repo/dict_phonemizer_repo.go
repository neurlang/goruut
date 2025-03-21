package repo

import (
	"bytes"
	"encoding/csv"
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/repo/interfaces"
	"github.com/neurlang/classifier/hash"
	"github.com/spaolacci/murmur3"
	"strings"
	"sync"
	"compress/zlib"
	"encoding/json"
	"sort"
	"fmt"
)
import . "github.com/martinarisk/di/dependency_injection"

type IDictPhonemizerRepository interface {
	LookupWords(isReverse bool, lang string, word string) []map[uint32]string
	LookupTags(isReverse bool, lang string, word1, word2 string) string
}
type DictPhonemizerRepository struct {
	getter     *interfaces.DictGetter
	lang_words *map[string]map[string]map[uint32]string
	lang_tags  *map[string]map[uint32]string
	words_tags *map[string]map[[2]string]uint32
	mut    sync.Mutex
}

func murmur3hash(str string) uint32 {
	return murmur3.Sum32WithSeed([]byte(str), 0)
}

func addTags(bag map[uint32]string, tags ...string) map[uint32]string {
	for _, v := range tags {
		bag[hash.StringHash(0, v)] = v
	}
	return bag
}

func parseTags(cell string) (ret map[uint32]string) {
	ret = make(map[uint32]string)
	if cell == "" {
		return
	}
	var tags []string
	err := json.Unmarshal([]byte(cell), &tags)
	if err != nil {
		log.Error0(fmt.Errorf("Cell tag: %s, Error: %v", cell, err))
	}
	for _, v := range tags {
		ret[hash.StringHash(0, v)] = v
	}
	return
}

func serializeTags(tags map[uint32]string) (key uint32, ret string) {
	var tagstrings = []string{}
	for k, v := range tags {
		key ^= k
		tagstrings = append(tagstrings, v)
	}
	sort.Strings(tagstrings)
	data := log.Error1(json.Marshal(tagstrings))
	if len(data) > 0 {
		ret = string(data)
	} else {
		ret = "[]"
	}
	if key == 0 {
		key++
	}
	return
}

func (r *DictPhonemizerRepository) LoadLanguage(isReverse bool, lang string) {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	r.mut.Lock()
	defer r.mut.Unlock()
	if (*r.lang_words)[lang+reverse] == nil {
		(*r.lang_words)[lang+reverse] = make(map[string]map[uint32]string)
	} else {
		log.Now().Debugf("Language %s already loaded", lang)
		return
	}
	if (*r.lang_tags)[lang+reverse] == nil {
		(*r.lang_tags)[lang+reverse] = make(map[uint32]string)
	} else {
		log.Now().Debugf("Language %s already loaded", lang)
		return
	}
	if (*r.words_tags)[lang+reverse] == nil {
		(*r.words_tags)[lang+reverse] = make(map[[2]string]uint32)
	} else {
		log.Now().Debugf("Language %s already loaded", lang)
		return
	}

	var files = []string{"missing" + reverse + ".tsv", "missing.all.zlib"}

	for _, file := range files {
		clean := log.Error1((*r.getter).GetDict(lang, file))
		var reader *csv.Reader

		if len(clean) == 0 {
			continue
		}

		if strings.HasSuffix(file, ".zlib") {
			reader = csv.NewReader(log.Error1(zlib.NewReader(bytes.NewReader(clean))))
		} else {
			reader = csv.NewReader(bytes.NewReader(clean))
		}

		reader.Comma = '\t'

		recs := log.Error1(reader.ReadAll())
		log.Now().Debugf("Language %s has %d records", lang, len(recs))
		for _, v := range recs {
			for i := range v {
				v[i] = strings.ReplaceAll(v[i], " ", "")
			}
			var src, dst, tagstr string
			if len(v) == 2 {
				if isReverse {
					src = v[1]
					dst = v[0]
				} else {
					src = v[0]
					dst = v[1]
				}
				tagstr = "[]"
			} else if len(v) == 3 {
				if isReverse {
					src = v[1]
					dst = v[0]
				} else {
					src = v[0]
					dst = v[1]
				}
				tagstr = v[2]
			} else {
				log.Now().Debugf("Language %s has wrong number of columns: %d", src, len(v))
				continue
			}
			var tagkey, tagjson = serializeTags(addTags(parseTags(tagstr), "dict"))
			if (*r.lang_words)[lang+reverse][src] == nil {
				(*r.lang_words)[lang+reverse][src] = make(map[uint32]string)
			}
			(*r.lang_words)[lang+reverse][src][tagkey] = dst
			if (*r.lang_tags)[lang+reverse] == nil {
				(*r.lang_tags)[lang+reverse] = make(map[uint32]string)
			}
			(*r.lang_tags)[lang+reverse][tagkey] = tagjson
			if (*r.words_tags)[lang+reverse] == nil {
				(*r.words_tags)[lang+reverse] = make(map[[2]string]uint32)
			}
			(*r.words_tags)[lang+reverse][[2]string{src, dst}] = tagkey
		}
	}
}

func (r *DictPhonemizerRepository) LookupWords(isReverse bool, lang, word string) (ret []map[uint32]string) {
	r.LoadLanguage(isReverse, lang)
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	r.mut.Lock()
	found := (*r.lang_words)[lang+reverse][word]
	r.mut.Unlock()

	if len(found) == 0 {
		return nil
	}
	var m = make(map[uint32]string)
	for k, v := range found {
		log.Now().Debugf("Key: %d, Value: %s", k, v)
		m[k] = v
	}
	m[0] = word
	ret = append(ret, m)
	return
}

func (r *DictPhonemizerRepository) LookupTags(isReverse bool, lang string, word1, word2 string) string {
	r.LoadLanguage(isReverse, lang)
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	r.mut.Lock()
	found := (*r.lang_tags)[lang+reverse][(*r.words_tags)[lang+reverse][[2]string{word1, word2}]]
	r.mut.Unlock()
	if found != "" {
		return found
	}
	return "[]"
}

func NewDictPhonemizerRepository(di *DependencyInjection) *DictPhonemizerRepository {
	getter := MustAny[interfaces.DictGetter](di)
	mapping := make(map[string]map[string]map[uint32]string)
	mapping2 := make(map[string]map[uint32]string)
	mapping3 := make(map[string]map[[2]string]uint32)

	return &DictPhonemizerRepository{
		getter:     &getter,
		lang_words: &mapping,
		lang_tags:  &mapping2,
		words_tags: &mapping3,
	}
}

var _ IDictPhonemizerRepository = &DictPhonemizerRepository{}
