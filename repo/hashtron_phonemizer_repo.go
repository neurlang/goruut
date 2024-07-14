package repo

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/neurlang/classifier/hash"
	"github.com/neurlang/classifier/hashtron"
	"github.com/neurlang/classifier/layer/majpool2d"
	"github.com/neurlang/classifier/net/feedforward"
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/repo/interfaces"
	"io/ioutil"
	"strings"
	"sync"
	"unicode"
)
import . "github.com/martinarisk/di/dependency_injection"

type IHashtronPhonemizerRepository interface {
	PhonemizeWord(lang, word string) (ret map[uint64]string)
	CleanWord(lang, word string) string
	CheckWord(lang, word, ipa string) bool
}
type HashtronPhonemizerRepository struct {
	getter *interfaces.DictGetter

	mut    sync.RWMutex
	lang   *languages
	phoner *interfaces.Phonemizer
	nets   *map[string]*feedforward.FeedforwardNetwork
}

func get_hashtron(h *hashtron.Hashtron, err error) hashtron.Hashtron {
	if err != nil {
		log.Now().Errorf("Failed to load one hashtron: %v", err)
		return *log.Error1(hashtron.New(nil, 1))
	}
	return *h
}

type languages map[string]*language

type language struct {
	Mapping           map[string][]string `json:"Map"`
	SrcMulti          []string            `json:"SrcMulti"`
	DstMulti          []string            `json:"DstMulti"`
	SrcMultiSuffix    []string            `json:"SrcMultiSuffix"`
	DstMultiSuffix    []string            `json:"DstMultiSuffix"`
	DropLast          []string            `json:"DropLast"`
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

}
func (l *languages) SrcMulti(lang string) map[string]struct{} {
	return (*l)[lang].mapSrcMulti
}
func (l *languages) DstMulti(lang string) map[string]struct{} {
	return (*l)[lang].mapDstMulti
}
func (l *languages) SrcMultiSuffix(lang string) map[string]struct{} {
	return (*l)[lang].mapSrcMultiSuffix
}
func (l *languages) DstMultiSuffix(lang string) map[string]struct{} {
	return (*l)[lang].mapDstMultiSuffix
}
func (l *languages) Map(lang string) map[string]map[string]struct{} {
	if (*l)[lang] == nil {
		return nil
	}
	return nil
}
func (l *languages) DroppedLast(lang, last string) bool {
	if (*l)[lang] == nil {
		return false
	}
	_, ok := (*l)[lang].mapDropLast[last]
	return ok
}
func (l *languages) Slice(lang string) map[string][]string {
	if (*l)[lang] == nil {
		return nil
	}
	return (*l)[lang].Mapping
}
func (l *languages) IsLetter(lang, run string) bool {
	if (*l)[lang] == nil {
		return false
	}
	_, ok := (*l)[lang].mapLetters[run]
	return ok
}

func (l *languages) SrcSlice(language string, word []rune) (o []string) {
	lang := (*l)[language]
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

func (r *HashtronPhonemizerRepository) LoadLanguage(lang string) {

	r.mut.RLock()
	nets := r.nets
	r.mut.RUnlock()

	if nets == nil {
		netss := make(map[string]*feedforward.FeedforwardNetwork)
		nets = &netss
		r.mut.Lock()
		r.nets = &netss
		r.mut.Unlock()
		log.Now().Debugf("Language %s made map of nets", lang)
	}
	if (*nets)[lang] == nil {
		var net feedforward.FeedforwardNetwork
		const fanout1 = 3
		const fanout2 = 12
		//const fanout3 = 3
		//const fanout4 = 10
		//net.NewLayerP(fanout1*fanout2*fanout3*fanout4, 0, 1033)
		//net.NewCombiner(majpool2d.MustNew(fanout1*fanout2*fanout4, 1, fanout3, 1, fanout4, 1, 1))
		net.NewLayerP(fanout1*fanout2, 0, 1<<fanout2)
		net.NewCombiner(majpool2d.MustNew(fanout2, 1, fanout1, 1, fanout2, 1, 1))
		net.NewLayer(1, 0)
		r.mut.Lock()
		(*r.nets)[lang] = &net
		r.mut.Unlock()
	} else {
		log.Now().Debugf("Language %s already loaded", lang)
		return
	}

	var language_files = []string{"language.json"}
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

		r.mut.Lock()
		(*r.lang)[lang] = &langone

		iface := (interfaces.Phonemizer)(&(*r.lang))
		r.phoner = &iface
		r.mut.Unlock()
	}

	var files = []string{"weights0.json.gz"}

	for _, file := range files {
		compressedData := log.Error1((*r.getter).GetDict(lang, file))

		// Step 3: Decompress the data in memory
		gzipReader, err := gzip.NewReader(bytes.NewReader(compressedData))
		if err != nil {
			log.Now().Errorf("Failed to create gzip reader: %v", err)
			continue
		}
		defer gzipReader.Close()

		// Step 4: Read the decompressed data into memory
		decompressedData, err := ioutil.ReadAll(gzipReader)
		if err != nil {
			log.Now().Errorf("Failed to read decompressed data: %v", err)
			continue
		}

		// Step 5: Parse the JSON data into the specified type
		var data [][][2]uint32
		err = json.Unmarshal(decompressedData, &data)
		if err != nil {
			log.Now().Errorf("Failed to parse JSON data: %v", err)
			continue
		}
		r.mut.Lock()
		// Load the weights into the network
		for i, v := range data {
			*((*r.nets)[lang].GetHashtron(i)) = get_hashtron(hashtron.New(v, 1))
		}
		r.mut.Unlock()
		return
	}
}

// TODO: move to classifier repo
type sample [14]uint32

// TODO: move to classifier repo
func (s *sample) Feature(n int) uint32 {
	a := hash.Hash(uint32(n), 0, 13)
	/*
		b := n % 28
		if b >= a {
			b++
		}
	*/
	return s[a] /*+ s[b]*/ + s[13]
}

func isCombining(r uint32) bool {
	return unicode.Is(unicode.Mn, rune(r)) || unicode.Is(unicode.Mc, rune(r))
}
func addLetters(word string, mapping map[string]struct{}) {
	if mapping == nil {
		return
	}
	var reverse []uint32
	for _, r := range []rune(word) {
		reverse = append([]uint32{uint32(r)}, reverse...)
	}
	var str string
	for _, r := range reverse {
		if isCombining(r) {
			str += string(rune(r))
		} else {
			mapping[string(str)+string(rune(r))] = struct{}{}
			str = ""
		}
	}
	if str != "" {
		mapping[string(str)] = struct{}{}
	}
}

func (r *HashtronPhonemizerRepository) CleanWord(lang, word string) (ret string) {
	r.LoadLanguage(lang)

	var reverse []uint32
	for _, r := range []rune(word) {
		reverse = append([]uint32{uint32(r)}, reverse...)
	}

	var strings []string
	var str string
	for _, r := range reverse {
		if isCombining(r) {
			str += string(rune(r))
		} else {
			strings = append([]string{string(str) + string(rune(r))}, strings...)
			str = ""
		}
	}
	if str != "" {
		strings = append([]string{string(str)}, strings...)
	}
	for _, run := range strings {
		r.mut.RLock()
		isLanguageLetter := r.lang.IsLetter(lang, run)
		r.mut.RUnlock()

		if !isLanguageLetter {
			continue
		}

		log.Now().Debugf("Allowed run of word: %s", run)
		for _, r := range run {
			ret += string(r)
		}
	}
	return
}
func (r *HashtronPhonemizerRepository) CheckWord(lang, word, ipa string) bool {
	r.LoadLanguage(lang)

	r.mut.RLock()
	mapLangIsNil := r.lang.Slice(lang) == nil
	r.mut.RUnlock()
	if mapLangIsNil {
		return false
	}
	r.mut.RLock()
	srca := r.lang.SrcSlice(lang, []rune(word))
	r.mut.RUnlock()
	if len(srca) == 0 {
		return false
	}
outer:
	for i := 0; i < len(srca); i++ {
		srcv := srca[i]
		r.mut.RLock()
		m := r.lang.Slice(lang)[string(srcv)]
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
func (r *HashtronPhonemizerRepository) PhonemizeWord(lang, word string) (ret map[uint64]string) {
	r.LoadLanguage(lang)

	r.mut.RLock()
	mapLangIsNil := r.lang.Slice(lang) == nil
	r.mut.RUnlock()
	if mapLangIsNil {
		ret = make(map[uint64]string)
		ret[murmur3hash(word+"\x00"+word)] = word
		return
	}

	var backoffs = 10
	r.mut.RLock()
	srca := r.lang.SrcSlice(lang, []rune(word))
	r.mut.RUnlock()
	dsta := []string{}
outer:
	for i := 0; i < len(srca); i++ {
		srcv := srca[i]
		r.mut.RLock()
		m := r.lang.Slice(lang)[string(srcv)]
		if i == len(srca)-1 {
			if r.lang.DroppedLast(lang, string(srcv)) {
				m = append([]string{""}, m...)
			}
		}
		r.mut.RUnlock()

		if len(m) == 0 {
			dsta = append(dsta, "")
			continue
		}
		if len(m) == 1 {
			for _, mfirst := range m {
				dsta = append(dsta, mfirst)
				break
			}
			continue
		}
		for _, option := range m {
			j := len(srca) - i
			var buf = [...]uint32{
				hash.StringsHash(0, srca[1*i/2:i+j/2]),
				hash.StringsHash(0, srca[2*i/3:i+j/3]),
				hash.StringsHash(0, srca[4*i/5:i+j/5]),
				hash.StringsHash(0, srca[6*i/7:i+j/11]),
				hash.StringsHash(0, srca[10*i/11:i+j/11]),
				hash.StringsHash(0, dsta[0:i]),
				hash.StringsHash(0, srca),
				hash.StringsHash(0, dsta[0:4*i/7]),
				hash.StringsHash(0, dsta[4*i/7:6*i/7]),
				hash.StringsHash(0, dsta[6*i/7:i]),
				hash.StringsHash(0, srca[i:i+j/7]),
				hash.StringsHash(0, srca[i+j/7:i+3*j/7]),
				hash.StringsHash(0, srca[i+3*j/7:i+j]),
				hash.StringHash(0, option),
			}
			var input = sample(buf)
			r.mut.RLock()
			net := (*r.nets)[lang]
			r.mut.RUnlock()
			if net == nil {
				continue
			}
			r.mut.RLock()
			var predicted = net.Infer(&input).Feature(0)
			r.mut.RUnlock()
			if predicted == 1 {
				dsta = append(dsta, option)
				continue outer
			}
		}
		if backoffs > 0 {
			i = -1
			dsta = nil
			backoffs--
			continue
		}
		for _, mfirst := range m {
			dsta = append(dsta, mfirst)
			break
		}
	}
	var dst string
	for _, v := range dsta {
		dst += v
	}

	ret = make(map[uint64]string)
	ret[murmur3hash(word+"\x00"+dst)] = dst
	return
}

func NewHashtronPhonemizerRepository(di *DependencyInjection) *HashtronPhonemizerRepository {
	getter := MustAny[interfaces.DictGetter](di)
	langs := make(languages)

	return &HashtronPhonemizerRepository{
		getter: &getter,
		lang:   &langs,
	}
}

var _ IHashtronPhonemizerRepository = &HashtronPhonemizerRepository{}
