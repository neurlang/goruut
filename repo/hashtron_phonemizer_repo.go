package repo

import (
	"bytes"
	"encoding/json"
	"compress/zlib"
	"github.com/neurlang/classifier/hash"
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/repo/interfaces"
	"github.com/neurlang/noaregtransformer/go/noareg"
	"strings"
	"sync"
	"unicode"
)

//import "fmt"
import . "github.com/martinarisk/di/dependency_injection"

type IHashtronPhonemizerRepository interface {
	CleanWord(isReverse bool, word string, languages []string) (ret string, lpunct string, rpunct string)
	CheckWord(isReverse bool, lang, word, ipa string) bool
	PhonemizeWords(isReverse bool, lang string, word string) []map[string]uint32
	ExplainWord(isReverse bool, word1, word2, lang string) (ret map[string][]string)
	//PhonemizeWord(isReverse bool, lang string, word string) map[uint64]string
}
type HashtronPhonemizerRepository struct {
	getter *interfaces.DictGetter

	mut    *sync.RWMutex
	lang   *languages
	phoner *interfaces.Phonemizer

	tformers *map[string]*noareg.NoaregTransformer
}

func hashtronHash(str string) uint32 {
	return hash.StringHash(0, str)
}

type languages map[string]*language

type language struct {
	Mapping        map[string][]string `json:"Map"`
	SrcMulti       []string            `json:"SrcMulti"`
	DstMulti       []string            `json:"DstMulti"`
	SrcMultiSuffix []string            `json:"SrcMultiSuffix"`
	DstMultiSuffix []string            `json:"DstMultiSuffix"`
	DropLast       []string            `json:"DropLast"`
	SrcDuplicate   [][]string          `json:"SrcDuplicate"`
	//Histogram         []string            `json:"Histogram"`
	mapTokenizer      map[uint32]map[[2]uint32]string
	mapSrcMultiLen    int
	mapSrcMultiSufLen int
	mapSrcMulti       map[string]struct{}
	mapDstMulti       map[string]struct{}
	mapSrcMultiSuffix map[string]struct{}
	mapDstMultiSuffix map[string]struct{}
	mapLetters        map[string]struct{}
	mapDropLast       map[string]struct{}
}

func mapize(arr []string) (out map[string]struct{}) {
	out = make(map[string]struct{})
	for _, v := range arr {
		out[v] = struct{}{}
	}
	return
}

func (l *language) mapize() {
	l.mapTokenizer = noareg.MakeDetokenizer(l.Mapping)
	l.mapSrcMulti = mapize(l.SrcMulti)
	l.mapDstMulti = mapize(l.DstMulti)
	l.mapSrcMultiSuffix = mapize(l.SrcMultiSuffix)
	l.mapDstMultiSuffix = mapize(l.DstMultiSuffix)
	l.mapDropLast = mapize(l.DropLast)
	l.SrcMulti = nil
	l.DstMulti = nil
	l.SrcMultiSuffix = nil
	l.DstMultiSuffix = nil
	l.DropLast = nil
}

func (l *language) srcdst() {
	for k, v := range l.Mapping {
		if v == nil || len(v) == 0 {
			continue
		}
		if len([]rune(k)) > 1 {
			l.mapSrcMulti[k] = struct{}{}
		}
		for _, w := range v {
			if len([]rune(w)) > 1 {
				l.mapDstMulti[w] = struct{}{}
			}
		}
	}
	for k := range l.mapSrcMulti {
		if len([]rune(k)) > l.mapSrcMultiLen {
			l.mapSrcMultiLen = len([]rune(k))
		}
	}
	for k := range l.mapSrcMultiSuffix {
		if len([]rune(k)) > l.mapSrcMultiSufLen {
			l.mapSrcMultiSufLen = len([]rune(k))
		}
	}
}
func (l *language) letters() {
	l.mapLetters = make(map[string]struct{})
	for k := range l.Mapping {
		addLetters(k, l.mapLetters)
	}
	for _, rule := range l.SrcDuplicate {
		for _, v := range rule {
			addLetters(v, l.mapLetters)
		}
	}
}

/*
	func (l *languages) Histogram(isReverse bool, lang string) []string {
		var reverse string
		if isReverse {
			reverse = "_reverse"
		}
		if (*l)[lang+reverse] == nil {
			return nil
		}
		return (*l)[lang+reverse].Histogram
	}
*/
func (l *languages) SrcMulti(isReverse bool, lang string) map[string]struct{} {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	return (*l)[lang+reverse].mapSrcMulti
}
func (l *languages) SrcDuplicate(isReverse bool, lang string) [][]string {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	return (*l)[lang+reverse].SrcDuplicate
}

func (l *languages) DstMulti(isReverse bool, lang string) map[string]struct{} {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	return (*l)[lang+reverse].mapDstMulti
}
func (l *languages) SrcMultiSuffix(isReverse bool, lang string) map[string]struct{} {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	return (*l)[lang+reverse].mapSrcMultiSuffix
}
func (l *languages) DstMultiSuffix(isReverse bool, lang string) map[string]struct{} {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	return (*l)[lang+reverse].mapDstMultiSuffix
}
func (l *languages) Map(isReverse bool, lang string) map[string]map[string]struct{} {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	if (*l)[lang+reverse] == nil {
		return nil
	}
	return nil
}
func (l *languages) DroppedLast(isReverse bool, lang, last string) bool {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	if (*l)[lang+reverse] == nil {
		return false
	}
	_, ok := (*l)[lang+reverse].mapDropLast[last]
	return ok
}
func (l *languages) Slice(isReverse bool, lang string) map[string][]string {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	if (*l)[lang+reverse] == nil {
		return nil
	}
	return (*l)[lang+reverse].Mapping
}
func (l *languages) IsLetter(isReverse bool, lang, run string) bool {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	if (*l)[lang+reverse] == nil {
		return false
	}
	_, ok := (*l)[lang+reverse].mapLetters[run]
	return ok
}
func (l *languages) Detokenizer(isReverse bool, lang string) map[uint32]map[[2]uint32]string {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	if (*l)[lang+reverse] == nil {
		return nil
	}
	t := (*l)[lang+reverse].mapTokenizer
	return t
}

func (l *languages) SrcSlice(isReverse bool, language string, word []rune) (o []string) {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	lang := (*l)[language+reverse]
outer:
	for i := 0; i < len(word); i++ {
		if lang != nil {
			for j := lang.mapSrcMultiLen; j > 0; j-- {
				for multi := range lang.mapSrcMulti {
					if len(multi) != j {
						continue
					}
					if strings.HasPrefix(string(word[i:]), multi) {
						o = append(o, multi)
						i += len([]rune(multi)) - 1
						if i >= len(word) {
							return
						}
						continue outer
					}
				}
			}
			for j := lang.mapSrcMultiSufLen; j > 0; j-- {
				for multi := range lang.mapSrcMultiSuffix {
					if len(multi) != j {
						continue
					}
					if len(o) > 0 && strings.HasPrefix(string(word[i:]), multi) {
						o[len(o)-1] += multi
						i += len([]rune(multi)) - 1
						if i >= len(word) {
							return
						}
						continue outer
					}
				}
			}
		}
		o = append(o, string(word[i]))
	}
	return o
}

func (r *HashtronPhonemizerRepository) LoadLanguage(isReverse bool, lang string) {
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
	/*
		aregnets := r.aregnets
		if aregnets == nil {
			aregnetss := make(map[string]*feedforward.FeedforwardNetwork)
			aregnets = &aregnetss
			r.aregnets = &aregnetss
			log.Now().Debugf("Language %s made map of nets", lang)
		}
	*/
	if (*tformers)[lang+reverse] != nil {
		log.Now().Debugf("Language %s already loaded", lang)
		return
	}

	var language_files = []string{"language" + reverse + ".json"}
	for _, file := range language_files {
		log.Now().Debugf("Language %s loading file", file)
		data := log.Error1((*r.getter).GetDict(lang, file))

		// Parse the JSON data into the Language struct
		var langone language
		err := json.Unmarshal(data, &langone)
		if err != nil {
			log.Now().Errorf("Error parsing JSON: %v\n", err)
			continue
		}

		langone.mapize()
		langone.srcdst()
		langone.letters()

		log.Now().Debugf("Language %s loaded: %v", lang, langone)

		(*r.lang)[lang+reverse] = &langone

		iface := (interfaces.Phonemizer)(&(*r.lang))
		r.phoner = &iface
	}

	var noareg_files = []string{
		"weights8" + reverse + ".bin.zlib",
	}
	for i, file := range noareg_files {
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

func isCombining(r uint32) bool {
	return unicode.Is(unicode.Mn, rune(r)) || unicode.Is(unicode.Mc, rune(r))
}
func addLetters(word string, mapping map[string]struct{}) {
	if mapping == nil {
		return
	}

	runes := []rune(word)
	n := len(runes)
	var str string
	var baseLetterFound bool

	// Process characters in reverse order
	for i := n - 1; i >= 0; i-- {
		r := runes[i]

		if isCombining(uint32(r)) {
			// Accumulate combining characters
			str = string(r) + str
		} else {
			// Found a base letter, prepend accumulated combiners
			fullLetter := string(r) + str
			log.Now().Debugf("Adding to letters: %s", fullLetter)
			mapping[fullLetter] = struct{}{}

			// Reset for next letter
			str = ""
			baseLetterFound = true
		}
	}

	// Edge case: if only combiners were present at the start
	if str != "" && !baseLetterFound {
		log.Now().Debugf("Adding standalone combiners: %s", str)
		mapping[str] = struct{}{}
	}
}

// CleanWord returns cleaned word, left punct, right punct
func (r *HashtronPhonemizerRepository) CleanWord(isReverse bool, word string, languages []string) (ret string, lpunct string, rpunct string) {
	for _, lang := range languages {
		r.LoadLanguage(isReverse, lang)
	}

	reverse := make([]uint32, len([]rune(word)))
	for i, r := range []rune(word) {
		reverse[len(reverse)-1-i] = uint32(r)
	}

	var strings []string
	var str string
	for _, run := range reverse {
		if isCombining(run) {
			// Always attach combining marks to the accumulating sequence
			str = string(rune(run)) + str
			continue
		}

		// Attach previous combining characters to this base letter
		fullGrapheme := string(rune(run)) + str
		strings = append([]string{fullGrapheme}, strings...)
		str = "" // Reset accumulator
	}
	if str != "" {
		strings = append([]string{string(str)}, strings...)
	}
	log.Now().Debugf("strings: %v, len: %v", strings, len(strings))
	for i, run := range strings {

		var isLanguageLetter = false
		r.mut.RLock()
		for _, lang := range languages {
			if isLanguageLetter {
				break
			}
			isLanguageLetter = r.lang.IsLetter(isReverse, lang, run)
		}
		r.mut.RUnlock()

		if !isLanguageLetter {
			if 2*i < len(strings) {
				lpunct += run
			} else {
				rpunct += run
			}
			log.Now().Debugf("Not language letter: %s", run)
			continue
		} else {
			log.Now().Debugf("Allowed run of word: %s", run)
		}
		ret += run
	}
	return
}

func (r *HashtronPhonemizerRepository) CheckWord(isReverse bool, lang, word, ipa string) bool {
	r.LoadLanguage(isReverse, lang)

	r.mut.RLock()
	mapLangIsNil := r.lang.Slice(isReverse, lang) == nil
	r.mut.RUnlock()
	if mapLangIsNil {
		return false
	}
	r.mut.RLock()
	srca := r.lang.SrcSlice(isReverse, lang, []rune(word))
	r.mut.RUnlock()
	if len(srca) == 0 {
		return false
	}
outer:
	for i := 0; i < len(srca); i++ {
		srcv := srca[i]
		r.mut.RLock()
		m := r.lang.Slice(isReverse, lang)[string(srcv)]
		r.mut.RUnlock()
		if len(m) == 0 {
			return false
		}
		for _, option := range m {
			if strings.HasPrefix(ipa, option) {
				ipa = ipa[len(option):]
				continue outer
			}
		}
		return false
	}
	return true
}

func copystrings(s []string) (r []string) {
	r = make([]string, len(s))
	copy(r, s)
	return
}

func (r *HashtronPhonemizerRepository) ExplainWord(isReverse bool, word1, word2, lang string) (ret map[string][]string) {
	ret = make(map[string][]string)
	r.LoadLanguage(isReverse, lang)
	r.mut.RLock()
	srca := r.lang.SrcSlice(isReverse, lang, []rune(word1))
	r.mut.RUnlock()
	for i := 0; i < len(srca); i++ {
		srcv := srca[i]
		r.mut.RLock()
		m := copystrings(r.lang.Slice(isReverse, lang)[string(srcv)])
		r.mut.RUnlock()
		ret[srcv] = m
	}
	return
}

func (r *HashtronPhonemizerRepository) PhonemizeWords(isReverse bool, lang string, word string) (ret []map[string]uint32) {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	r.LoadLanguage(isReverse, lang)

	r.mut.RLock()
	is_new := (*r.tformers)[lang+reverse] != nil
	r.mut.RUnlock()

	if is_new {

		r.mut.RLock()
		tran_new := (*r.tformers)[lang+reverse]
		detokenizer_new := r.lang.Detokenizer(isReverse, lang)
		is_ok := tran_new != nil && detokenizer_new != nil
		r.mut.RUnlock()

		if is_ok {

			out, _ := log.Error2(noareg.TransformerInferFull2(tran_new, detokenizer_new, word))
			//println(word, gbg[0], gbg[1], out)

			var fullsrc string
			var fulldst string

			for _, srcdst := range out {

				src := srcdst[0]
				dst := srcdst[1]

				if strings.HasPrefix(dst, "_") && fullsrc != "" && fulldst != "" {
					m := make(map[string]uint32)
					hsh := hashtronHash(fullsrc + "\x00" + fulldst)
					if hsh == 0 {
						hsh++
					}
					m[fulldst] = uint32(hsh)
					m[fullsrc+" "] = 0
					ret = append(ret, m)

					fullsrc = ""
					fulldst = ""
				}

				fullsrc += src
				fulldst += dst

				if strings.HasSuffix(fulldst, "_") && fullsrc != "" && fulldst != "" {
					m := make(map[string]uint32)
					hsh := hashtronHash(fullsrc + "\x00" + fulldst)
					if hsh == 0 {
						hsh++
					}
					m[fulldst] = uint32(hsh)
					m[fullsrc+" "] = 0
					ret = append(ret, m)

					fullsrc = ""
					fulldst = ""
				}
			}
			if fullsrc != "" && fulldst != "" {
				m := make(map[string]uint32)
				hsh := hashtronHash(fullsrc + "\x00" + fulldst)
				if hsh == 0 {
					hsh++
				}
				m[fulldst] = uint32(hsh)
				m[fullsrc+" "] = 0
				ret = append(ret, m)
			}

			return
		}
	}
	return
}

func NewHashtronPhonemizerRepository(di *DependencyInjection) *HashtronPhonemizerRepository {
	getter := MustAny[interfaces.DictGetter](di)
	langs := make(languages)

	return &HashtronPhonemizerRepository{
		getter: &getter,
		lang:   &langs,
		mut:    &sync.RWMutex{},
	}
}

var _ IHashtronPhonemizerRepository = &HashtronPhonemizerRepository{}
