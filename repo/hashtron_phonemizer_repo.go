package repo

import (
	"bytes"
	"encoding/json"
	"github.com/neurlang/classifier/datasets/phonemizer"
	"github.com/neurlang/classifier/hashtron"
	//"github.com/neurlang/classifier/layer/full"
	"github.com/neurlang/classifier/layer/sochastic"
	"github.com/neurlang/classifier/layer/majpool2d"
	"github.com/neurlang/classifier/layer/crossattention"
	"github.com/neurlang/classifier/layer/parity"
	"github.com/neurlang/classifier/layer/sum"
	"github.com/neurlang/classifier/net/feedforward"
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/repo/interfaces"
	"strings"
	"sync"
	"unicode"
)
//import "fmt"
import . "github.com/martinarisk/di/dependency_injection"

type IHashtronPhonemizerRepository interface {
	CleanWord(isReverse bool, lang, word string) (ret string, lpunct string, rpunct string)
	CheckWord(isReverse bool, lang, word, ipa string) bool
	PhonemizeWords(isReverse bool, lang string, word string) []map[uint32]string
	//PhonemizeWord(isReverse bool, lang string, word string) map[uint64]string
}
type HashtronPhonemizerRepository struct {
	getter *interfaces.DictGetter

	mut    sync.RWMutex
	lang   *languages
	phoner *interfaces.Phonemizer
	nets   *map[string]*feedforward.FeedforwardNetwork

	aregnets *map[string]*feedforward.FeedforwardNetwork
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
	//Histogram         []string            `json:"Histogram"`
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

	r.mut.Lock()
	defer r.mut.Unlock()

	nets := r.nets
	if nets == nil {
		netss := make(map[string]*feedforward.FeedforwardNetwork)
		nets = &netss
		r.nets = &netss
		log.Now().Debugf("Language %s made map of nets", lang)
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
	if (*nets)[lang+reverse] != nil {
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

	var files = []string{"weights4" + reverse + ".json.zlib", "weights2" + reverse + ".json.zlib", "weights1" + reverse + ".json.zlib"}

	for i, file := range files {
		compressedData := log.Error1((*r.getter).GetDict(lang, file))

		if compressedData == nil {
			continue
		}

		if (*r.getter).IsNewFormat(compressedData) {
			bytesReader := bytes.NewReader(compressedData)

			switch i {
			case 2:
				const fanout1 = 1
				const fanout2 = 5
				const fanout3 = 3
				const fanout4 = 5

				var net feedforward.FeedforwardNetwork
				net.NewLayerP(fanout1*fanout2*fanout3*fanout4, 0, 1<<13)
				net.NewCombiner(majpool2d.MustNew2(fanout1*fanout2*fanout4, 1, fanout3, 1, fanout4, 1, 1, 0))
				net.NewLayer(fanout1*fanout2, 0)
				net.NewCombiner(majpool2d.MustNew2(fanout2, 1, fanout1, 1, fanout2, 1, 1, 0))
				net.NewLayer(1, 0)

				(*r.nets)[lang+reverse] = &net
			case 1:

				const fanout1 = 5
				var net feedforward.FeedforwardNetwork
				//net.NewLayer(fanout1, 0)
				//net.NewCombiner(sochastic.MustNew(fanout1, 32, 0))
				net.NewLayer(fanout1, 0)
				net.NewCombiner(parity.MustNew(fanout1))
				net.NewLayer(1, 0)

				(*r.nets)[lang+reverse] = &net

			case 0:

				const fanout1 = 32
				const fanout2 = 4
				const fanout3 = 3
				
				var net feedforward.FeedforwardNetwork
				//net.NewLayer(fanout1, 0)
				//net.NewCombiner(sochastic.MustNew(fanout1, 32, 0))
				net.NewLayer(fanout1*fanout2, 0)
				for i := 0; i < fanout3; i++ {
					net.NewCombiner(crossattention.MustNew(fanout1, fanout2))
					net.NewLayerPI(fanout1*fanout2, 0, 0)
					net.NewCombiner(sochastic.MustNew(fanout1*fanout2, 8*byte(i), uint32(i)))
					net.NewLayerPI(fanout1*fanout2, 0, 0)
				}
				net.NewCombiner(sochastic.MustNew(fanout1*fanout2, 32, fanout3))
				net.NewLayer(fanout1*fanout2, 0)
				net.NewCombiner(sum.MustNew([]uint{fanout1*fanout2}, 0))
				net.NewLayer(1, 0)

				(*r.nets)[lang+reverse] = &net

			}
			err := (*r.nets)[lang+reverse].ReadZlibWeights(bytesReader)
			log.Error0(err)

			return
		} /*else if !isReverse  doesnt work: && (*r.getter).IsOldFormat(compressedData) {
			bytesReader := bytes.NewReader(compressedData)
			err := (*r.nets)[lang+reverse].ReadCompressedWeights(bytesReader)
			log.Error0(err)
			return
		}*/
	}
/*
	var aregfiles = []string{"weights3" + reverse + ".json.zlib"}
	for _, file := range aregfiles {
		compressedData := log.Error1((*r.getter).GetDict(lang, file))

		if compressedData == nil {
			continue
		}
		if (*r.getter).IsNewFormat(compressedData) {
			bytesReader := bytes.NewReader(compressedData)

			const fanout1 = 5
			var net feedforward.FeedforwardNetwork
			//net.NewLayer(fanout1, 0)
			//net.NewCombiner(sochastic.MustNew(fanout1, 32, 0))
			net.NewLayer(fanout1, 0)
			net.NewCombiner(parity.MustNew(fanout1))
			net.NewLayer(1, 0)

			(*r.aregnets)[lang+reverse] = &net

			err := (*r.aregnets)[lang+reverse].ReadZlibWeights(bytesReader)
			log.Error0(err)

			break
		}
	}
*/
}

func isCombining(r uint32) bool {
	return unicode.Is(unicode.Mn, rune(r)) || unicode.Is(unicode.Mc, rune(r))
}
func addLetters(word string, mapping map[string]struct{}) {
	if mapping == nil {
		return
	}
	reverse := make([]uint32, len([]rune(word)))
	for i, r := range []rune(word) {
		reverse[len(reverse)-1-i] = uint32(r)
	}
	var str string
	for _, r := range reverse {
		if isCombining(r) {
			str = string(rune(r)) + str
		} else {
			log.Now().Debugf("Adding to letters: %x", string(rune(r))+string(str))
			mapping[string(rune(r))+string(str)] = struct{}{}
			str = ""
		}
	}
	if str != "" {
		log.Now().Debugf("Adding to letters: %x", str)
		mapping[string(str)] = struct{}{}
	}
}

// CleanWord returns cleaned word, left punct, right punct
func (r *HashtronPhonemizerRepository) CleanWord(isReverse bool, lang, word string) (ret string, lpunct string, rpunct string) {
	r.LoadLanguage(isReverse, lang)

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
		r.mut.RLock()
		isLanguageLetter := r.lang.IsLetter(isReverse, lang, run)
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
/*
func (r *HashtronPhonemizerRepository) PhonemizeWord(isReverse bool, lang string, word string) (ret map[uint64]string) {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	r.LoadLanguage(isReverse, lang)

	r.mut.RLock()
	mapLangIsNil := r.lang.Slice(isReverse, lang) == nil
	r.mut.RUnlock()
	if mapLangIsNil {
		m := make(map[uint64]string)
		return m
	}

	r.mut.RLock()
	histogram := r.lang.Histogram(isReverse, lang)
	net := (*r.aregnets)[lang+reverse]
	r.mut.RUnlock()

	if net == nil {
		m := make(map[uint64]string)
		return m
	}

	for i := 0; i < 32; i++ {
		var input = phonemizer.AregSample{
			Src: word,
			Dst: fmt.Sprint(i),
		}
		r.mut.RLock()
		pred := net.Infer2(&input) == 1
		r.mut.RUnlock()

		if !pred {
			m := make(map[uint64]string)
			return m
		}
	}
	pred := true
	var out string
	for pred && len(out) < len(word)*2 {
		for _, val := range histogram {
			var input2 = phonemizer.AregSample{
				Src: word,
				Dst: out + val,
			}
			r.mut.RLock()
			pred = net.Infer2(&input2) == 1
			r.mut.RUnlock()
			//fmt.Println(word, out, val, pred)
			if pred {
				out += val
				break
			}
		}
	}
	//fmt.Println(word, out)
	m := make(map[uint64]string)
	hsh := murmur3hash(word + "\x00" + out)
	if hsh == 0 {
		hsh++
	}
	m[hsh] = out
	m[0] = word
	return m
}
*/
func (r *HashtronPhonemizerRepository) PhonemizeWords(isReverse bool, lang string, word string) (ret []map[uint32]string) {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}
	r.LoadLanguage(isReverse, lang)

	r.mut.RLock()
	mapLangIsNil := r.lang.Slice(isReverse, lang) == nil
	r.mut.RUnlock()
	if mapLangIsNil {
		return []map[uint32]string{}
	}

	var backoffs = 10
	r.mut.RLock()
	srca := r.lang.SrcSlice(isReverse, lang, []rune(word))
	r.mut.RUnlock()
	dsta := []string{}

	var lastspace = 0
outer:
	for i := 0; i < len(srca); i++ {
		srcv := srca[i]
		r.mut.RLock()
		m := r.lang.Slice(isReverse, lang)[string(srcv)]
		if i == len(srca)-1 {
			if r.lang.DroppedLast(isReverse, lang, string(srcv)) {
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
				if mfirst == "_" {
					lastspace = i + 1
				} else if strings.HasPrefix(mfirst, "_") {
					lastspace = i
				} else if strings.HasSuffix(mfirst, "_") {
					lastspace = i + 1
				}
				dsta = append(dsta, mfirst)
				break
			}
			continue
		}
		for _, option := range m {
			srcaR := srca[lastspace:]
			dstaR := dsta[lastspace:]
			origi := i
			i := i - lastspace
			r.mut.RLock()
			net := (*r.nets)[lang+reverse]
			r.mut.RUnlock()
			var multiword = lastspace > 0
			if net == nil {
				log.Now().Errorf("Net is nil")
				continue
			}
			var predicted int
			for q := 0; (!multiword && q == 0) || (multiword && q < len(srcaR)-i); q++ {
				var input = phonemizer.NewSample{
					SrcA:   copystrings(srcaR[:len(srcaR)-q]),
					DstA:   copystrings(dstaR[0:i]),
					SrcCut: copystrings(srcaR[0:i]),
					SrcFut: copystrings(srcaR[i : len(srcaR)-q]),
					Option: option,
				}
				var pred int
				r.mut.RLock()
				if net.LenLayers() == 3 {
					pred = int(net.Infer2(input.V1()))
				} else if net.LenLayers() == 5 {
					pred = int(net.Infer2(&input))
				} else { // newest model
					const fanout1 = 32
					pred = int(net.Infer2(input.V2(fanout1)))
				}
				r.mut.RUnlock()
				predicted += pred
				log.Now().Debugf("Model predicted: %v %v %v %v %v -> %d", input.SrcA, input.DstA, input.SrcCut, input.SrcFut, input.Option, pred)
			}
			if (!multiword && predicted == 1) || (multiword && 2*predicted > len(srcaR)) {
				if option == "_" {
					lastspace = origi + 1
				} else if strings.HasPrefix(option, "_") {
					lastspace = origi
				} else if strings.HasSuffix(option, "_") {
					lastspace = origi + 1
				}
				dsta = append(dsta, option)
				continue outer
			}
		}
		if backoffs > 0 {
			i = lastspace - 1
			dsta = dsta[:lastspace]
			backoffs--
			continue
		}
		for _, mfirst := range m {
			if mfirst == "_" {
				lastspace = i + 1
			} else if strings.HasPrefix(mfirst, "_") {
				lastspace = i
			} else if strings.HasSuffix(mfirst, "_") {
				lastspace = i + 1
			}
			dsta = append(dsta, mfirst)
			break
		}
	}
	var src, dst string

	push := func() {
		if len(src)+len(dst) > 0 {
			m := make(map[uint32]string)
			hsh := murmur3hash(src + "\x00" + dst)
			if hsh == 0 {
				hsh++
			}
			m[uint32(hsh)] = dst
			m[0] = src
			ret = append(ret, m)
			src, dst = "", ""
		}
	}
	for i, v := range dsta {
		if v != "_" && strings.HasPrefix(v, "_") {
			push()
		}
		src += srca[i]
		dst += v
		if strings.HasSuffix(v, "_") {
			push()
		}
	}
	push()
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
