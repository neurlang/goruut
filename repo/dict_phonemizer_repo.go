package repo

import (
	"bytes"
	"compress/zlib"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/neurlang/classifier/hash"
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/repo/interfaces"
	"github.com/neurlang/table"
	"sort"
	"strings"
	"sync"
	"github.com/neurlang/classifier/parallel"
	rand "math/rand/v2"
)
import . "github.com/martinarisk/di/dependency_injection"

type IDictPhonemizerRepository interface {
	LookupWords(isReverse bool, lang string, word string) []map[string]uint32
	LookupTags(isReverse bool, lang string, word1, word2 string) string
}
type DictPhonemizerRepository struct {
	getter     *interfaces.DictGetter
	lang_table *map[string][]table.Table
	mut        sync.Mutex
}

func addTags(bag map[uint32]string, tags ...string) map[uint32]string {
	for _, v := range tags {
		bag[hash.StringHash(0, v)] = v
	}
	return bag
}

func loadTags(cell string) (ret []string) {
	if cell == "" {
		return nil
	}
	dedup := make(map[string]struct{})
	var tags []string
	err := json.Unmarshal([]byte(cell), &tags)
	if err != nil {
		log.Error0(fmt.Errorf("Cell tag: %s, Error: %v", cell, err))
	}
	for _, v := range tags {
		dedup[v] = struct{}{}
	}
	for k := range dedup {
		ret = append(ret, k)
	}
	return
}

func storeTags(tags []string) string {
	sort.Strings(tags)
	data := log.Error1(json.Marshal(tags))
	if len(data) == 0 {
		return "[]"
	}
	return string(data)
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
	return
}

func (r *DictPhonemizerRepository) LoadLanguage(isReverse bool, lang string) {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	r.mut.Lock()
	defer r.mut.Unlock()

	var t = make([]table.Table, 4096, 4096)

	if (*r.lang_table)[lang] == nil {
		(*r.lang_table)[lang] = t
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
				src = v[0]
				dst = v[1]
				tagstr = "[]"
			} else if len(v) == 3 {
				src = v[0]
				dst = v[1]
				tagstr = storeTags(loadTags(v[2]))
			} else {
				log.Now().Debugf("Language %s has wrong number of columns: %d", src, len(v))
				continue
			}

			t[rand.IntN(len(t))].Insert([][]string{{src, dst, tagstr}})

		}

	}

	parallel.ForEach(len(t), len(t), func(i int) {
		t[i].Compact()
	})
}

func (r *DictPhonemizerRepository) LookupWords(isReverse bool, lang, word string) (ret []map[string]uint32) {
	r.LoadLanguage(isReverse, lang)
	var key = 0
	if isReverse {
		key = 1
	}
	r.mut.Lock()
	var muitx sync.Mutex
	var found2 [][]string
	parallel.ForEach(len((*r.lang_table)[lang]), len((*r.lang_table)[lang]), func(i int) {

		data := (*r.lang_table)[lang][i].QueryBy(map[int]string{key: word})
		if len(data) > 0 {
			muitx.Lock()
			found2 = append(found2, data...)
			muitx.Unlock()
		}
	})
	r.mut.Unlock()
	if len(found2) == 0 {
		return nil
	}
	var results = make(map[string]map[uint32]string)
	for _, row := range found2 {
		if results[row[1-key]] == nil {
			results[row[1-key]] = make(map[uint32]string)
		}
		addTags(results[row[1-key]], loadTags(row[2])...)
	}

	var m = make(map[string]uint32)
	for k, v := range results {
		addTags(v, "dict")
		w, _ := serializeTags(v)
		log.Now().Debugf("LookupWords Key: %s, Value: %v", k, w)
		m[k] = w
	}
	m[word + " "] = 0
	ret = append(ret, m)
	return
}

func (r *DictPhonemizerRepository) LookupTags(isReverse bool, lang string, word1, word2 string) string {
	r.LoadLanguage(isReverse, lang)
	if isReverse {
		word1, word2 = word2, word1
	}
	r.mut.Lock()
	var muitx sync.Mutex
	var found2 [][]string
	parallel.ForEach(len((*r.lang_table)[lang]), len((*r.lang_table)[lang]), func(i int) {

		data := (*r.lang_table)[lang][i].QueryBy(map[int]string{0: word1, 1: word2})
		if len(data) > 0 {
			muitx.Lock()
			found2 = append(found2, data...)
			muitx.Unlock()
		}
	})
	r.mut.Unlock()
	if len(found2) == 0 {
		return "[]"
	}

	var tags = []string{"dict"}

	for _, row := range found2 {
		tags = append(tags, loadTags(row[2])...)
	}

	return storeTags(tags)
}

func NewDictPhonemizerRepository(di *DependencyInjection) *DictPhonemizerRepository {
	getter := MustAny[interfaces.DictGetter](di)
	mapping4 := make(map[string][]table.Table)

	return &DictPhonemizerRepository{
		getter:     &getter,
		lang_table: &mapping4,
	}
}

var _ IDictPhonemizerRepository = &DictPhonemizerRepository{}
