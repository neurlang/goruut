package repo

import (
	"bytes"
	"compress/zlib"
	"github.com/neurlang/goruut/external/classifier/hash"
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/repo/interfaces"
	"github.com/neurlang/noaregtransformer/go/noareg"
	"strings"
	"sync"
)
import . "github.com/martinarisk/di/dependency_injection"

type IHashtronHomonymSelectorRepository interface {
	Select(isReverse bool, lang string, sentence []map[string][2]uint32) (ret [][4]uint32)
}

type HashtronHomonymSelectorRepository struct {
	getter *interfaces.DictGetter

	mut   *sync.RWMutex

	tformers *map[string]*noareg.NoaregTransformer
}

func (r *HashtronHomonymSelectorRepository) LoadLanguage(isReverse bool, lang string) {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}

	// First check with read lock to avoid unnecessary write lock
	r.mut.RLock()
	if r.tformers != nil && (*r.tformers)[lang+reverse] != nil {
		log.Now().Debugf("Language %s already loaded", lang)
		r.mut.RUnlock()
		return
	}
	r.mut.RUnlock()

	// Double-checked locking: acquire write lock and check again
	r.mut.Lock()
	defer r.mut.Unlock()

	tformers := r.tformers
	if tformers == nil {
		tformerss := make(map[string]*noareg.NoaregTransformer)
		tformers = &tformerss
		r.tformers = &tformerss
		log.Now().Debugf("Language %s made map of tformers", lang)
	}
	if (*tformers)[lang+reverse] != nil {
		log.Now().Debugf("Language %s already loaded", lang)
		return
	}
	var files_new = []string{
		"weights9" + reverse + ".bin.zlib",
	}
	for i, file := range files_new {
		compressedData := log.Error1((*r.getter).GetDict(lang, file))

		if compressedData == nil {
			continue
		}
		bytesReader := bytes.NewReader(compressedData)
		zlibReader := log.Error1(zlib.NewReader(bytesReader))
		if zlibReader == nil {
			continue
		}
		defer zlibReader.Close()

		switch i {
		case 0:
			tensors := log.Error1(noareg.ReadTensors(zlibReader))
			if tensors == nil {
				break
			}
			// Initialize transformer
			transformer := noareg.NewNoaregTransformer(32, 16, 100, 4)

			noareg.LoadTransformerFile(transformer, tensors)

			(*r.tformers)[lang+reverse] = transformer

			return
		}
	}

}

func (r *HashtronHomonymSelectorRepository) Select(isReverse bool, lang string, sentence []map[string][2]uint32) (ret [][4]uint32) {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	r.LoadLanguage(isReverse, lang)
	r.mut.RLock()
	tformer := (*r.tformers)[lang+reverse]
	r.mut.RUnlock()

	if tformer == nil {
		return
	}

	var datas [][][2]uint32
	var job [][]string
	for i, mapping := range sentence {
		var origword string
		var data [][2]uint32
		var strkey []string
		for v, k := range mapping {
			if k[0] == 0 {
				origword = strings.TrimRight(v, " ")
				continue
			}
			strkey = append(strkey, v)
			data = append(data, k)
		}
		strkey = append([]string{origword}, strkey...)
		log.Now().Debugf("Sentence %d: %v | %v", i, strkey, data)
		job = append(job, strkey)
		datas = append(datas, data)
	}

	var solution = log.Error1(noareg.MultiWordInferFull(tformer, job))
	log.Now().Debugf("Solution: %s", solution)

	var fields = strings.Fields(solution)
	for i, field := range fields {
		if len(job[i]) <= 1 {
			continue
		}
		var kk [2]uint32
		for j, entry := range job[i][1:] {
			if entry == field {
				kk = datas[i][j]
			}
		}
		ret = append(ret, [4]uint32{uint32(i), hash.StringHash(0, field), kk[0], 1})
	}
	log.Now().Debugf("Ret: %v", ret)
	return
}

func NewHashtronHomonymSelectorRepository(di *DependencyInjection) *HashtronHomonymSelectorRepository {
	getter := MustAny[interfaces.DictGetter](di)
	return &HashtronHomonymSelectorRepository{
		getter: &getter,
		mut:    &sync.RWMutex{},
	}
}

var _ IHashtronHomonymSelectorRepository = &HashtronHomonymSelectorRepository{}
