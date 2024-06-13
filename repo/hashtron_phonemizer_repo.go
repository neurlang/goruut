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
)
import . "github.com/martinarisk/di/dependency_injection"

type IHashtronPhonemizerRepository interface {
	PhonemizeWord(lang, word string) (ret map[uint64]string)
}
type HashtronPhonemizerRepository struct {
	getter *interfaces.DictGetter
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
	mapMapping        map[string]map[string]struct{}
	mapSrcMulti       map[string]struct{}
	mapDstMulti       map[string]struct{}
	mapSrcMultiSuffix map[string]struct{}
	mapDstMultiSuffix map[string]struct{}
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
	l.mapMapping = make(map[string]map[string]struct{})
	for k, v := range l.Mapping {
		l.mapMapping[k] = mapize(v)
	}
	l.SrcMulti = nil
	l.DstMulti = nil
	l.SrcMultiSuffix = nil
	l.DstMultiSuffix = nil
	l.Mapping = nil
}

func (l *language) srcdst() {
	for k, v := range l.mapMapping {
		if len([]rune(k)) > 1 {
			l.mapSrcMulti[k] = struct{}{}
		}
		for w := range v {
			if len([]rune(w)) > 1 {
				l.mapDstMulti[w] = struct{}{}
			}
		}
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
	return (*l)[lang].mapMapping
}
func (l *languages) SrcSlice(language string, word []rune) (o []string) {
	lang := (*l)[language]
outer:
	for i := 0; i < len(word); i++ {
		if lang != nil {
			for multi := range lang.mapSrcMulti {
				if strings.HasPrefix(string(word[i:]), multi) {
					o = append(o, multi)
					i += len([]rune(multi)) - 1
					if i >= len(word) {
						return
					}
					continue outer
				}
			}
			for multi := range lang.mapSrcMultiSuffix {
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
		o = append(o, string(word[i]))
	}
	return o
}

func (r *HashtronPhonemizerRepository) LoadLanguage(lang string) {
	if r.nets == nil {
		nets := make(map[string]*feedforward.FeedforwardNetwork)
		r.nets = &nets
		log.Now().Debugf("Language %s made map of nets", lang)
	}
	if (*r.nets)[lang] == nil {
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
		(*r.nets)[lang] = &net
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
		log.Now().Debugf("Language %s loaded: %v", lang, langone)
		(*r.lang)[lang] = &langone

		iface := (interfaces.Phonemizer)(&(*r.lang))
		r.phoner = &iface
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
		// Load the weights into the network
		for i, v := range data {
			*((*r.nets)[lang].GetHashtron(i)) = get_hashtron(hashtron.New(v, 1))
		}
		return
	}
}

// TODO: move to classifier repo
func stringsHash(in uint32, strs []string) (out uint32) {
	out = in
	for _, str := range strs {
		out = stringHash(out, str)
	}
	return
}

// TODO: move to classifier repo
func stringHash(in uint32, str string) (out uint32) {
	out = in
	for _, v := range []rune(str) {
		out = hash.Hash(out, uint32(v), 0xffffffff)
	}
	return
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

func (r *HashtronPhonemizerRepository) PhonemizeWord(lang, word string) (ret map[uint64]string) {
	r.LoadLanguage(lang)

	if r.lang.Map(lang) == nil {
		ret = make(map[uint64]string)
		ret[murmur3hash(word+"\x00"+word)] = word
		return
	}

	srca := r.lang.SrcSlice(lang, []rune(word))
	dsta := []string{}
outer:
	for i, srcv := range srca {
		m := r.lang.Map(lang)[string(srcv)]
		if len(m) == 0 {
			dsta = append(dsta, "")
			continue
		}
		if len(m) == 1 {
			for mfirst := range m {
				dsta = append(dsta, mfirst)
				break
			}
			continue
		}
		for option := range m {
			j := len(srca) - i
			var buf = [...]uint32{
				stringsHash(0, srca[1*i/2:i+j/2]),
				stringsHash(0, srca[2*i/3:i+j/3]),
				stringsHash(0, srca[4*i/5:i+j/5]),
				stringsHash(0, srca[6*i/7:i+j/11]),
				stringsHash(0, srca[10*i/11:i+j/11]),
				stringsHash(0, dsta[0:i]),
				stringsHash(0, srca),
				stringsHash(0, dsta[0:4*i/7]),
				stringsHash(0, dsta[4*i/7:6*i/7]),
				stringsHash(0, dsta[6*i/7:i]),
				stringsHash(0, srca[i:i+j/7]),
				stringsHash(0, srca[i+j/7:i+3*j/7]),
				stringsHash(0, srca[i+3*j/7:i+j]),
				stringHash(0, option),
			}
			var input = sample(buf)
			var predicted = (*r.nets)[lang].Infer(&input).Feature(0)
			if predicted == 1 {
				dsta = append(dsta, option)
				continue outer
			}
		}
		for mfirst := range m {
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
