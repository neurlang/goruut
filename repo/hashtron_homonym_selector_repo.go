package repo

import (
	"bytes"
	"fmt"
	"github.com/neurlang/classifier/layer/crossattention"
	"github.com/neurlang/classifier/layer/sochastic"
	"github.com/neurlang/classifier/layer/sum"
	"github.com/neurlang/classifier/net/feedforward"
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/repo/interfaces"
	"sort"
	"strconv"
	"strings"
	"sync"
)
import "github.com/neurlang/classifier/datasets/phonemizer_multi"
import "github.com/neurlang/classifier/hash"
import . "github.com/martinarisk/di/dependency_injection"

type IHashtronHomonymSelectorRepository interface {
	Select(isReverse bool, lang string, sentence []map[string][2]uint32) (ret [][4]uint32)
}

type HashtronHomonymSelectorRepository struct {
	getter *interfaces.DictGetter

	mut   *sync.RWMutex
	hlang *hlanguages
	nets  *map[string]*feedforward.FeedforwardNetwork
}

type hlanguages map[string]*hlanguage
type hlanguage struct {
}

func (r *HashtronHomonymSelectorRepository) LoadLanguage(isReverse bool, lang string) {
	var reverse string
	if isReverse {
		reverse = "_reverse"
	}

	// First check with read lock to avoid unnecessary write lock
	r.mut.RLock()
	if r.nets != nil && (*r.nets)[lang+reverse] != nil {
		log.Now().Debugf("Language %s already loaded", lang)
		r.mut.RUnlock()
		return
	}
	r.mut.RUnlock()

	// Double-checked locking: acquire write lock and check again
	r.mut.Lock()
	defer r.mut.Unlock()

	nets := r.nets
	if nets == nil {
		netss := make(map[string]*feedforward.FeedforwardNetwork)
		nets = &netss
		r.nets = &netss
		log.Now().Debugf("Language %s made map of nets", lang)
	}
	if (*nets)[lang+reverse] != nil {
		log.Now().Debugf("Language %s already loaded", lang)
		return
	}

	var files = []string{
		"weights7" + reverse + ".json.zlib",
		//"weights5" + reverse + ".json.zlib",
	}
	for i, file := range files {
		compressedData := log.Error1((*r.getter).GetDict(lang, file))

		if compressedData == nil {
			continue
		}
		if (*r.getter).IsNewFormat(compressedData) {
			bytesReader := bytes.NewReader(compressedData)

			switch i {
			case 0:
				const fanout1 = 24
				const fanout2 = 1
				const fanout3 = 4
				const fanout4 = 32

				var net feedforward.FeedforwardNetwork
				net.NewLayer(fanout1*fanout2, 0)
				for i := 0; i < fanout3; i++ {
					net.NewCombiner(crossattention.MustNew3(fanout1, fanout2))
					net.NewLayerPI(fanout1*fanout2, 0, 0)
					net.NewCombiner(sochastic.MustNew(fanout1*fanout2, fanout4-8*byte(i), uint32(i)))
					net.NewLayerPI(fanout1*fanout2, 0, 0)
				}
				net.NewCombiner(sochastic.MustNew(fanout1*fanout2, fanout4, fanout3))
				net.NewLayer(fanout1*fanout2, 0)
				net.NewCombiner(sum.MustNew([]uint{fanout1 * fanout2}, 0))
				net.NewLayer(1, 0)

				(*r.nets)[lang+reverse] = &net


			}
			err := (*r.nets)[lang+reverse].ReadZlibWeights(bytesReader)
			log.Error0(err)

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
	net := (*r.nets)[lang+reverse]
	r.mut.RUnlock()

	if net == nil {
		return
	}

	var ai_sentence = phonemizer_multi.Sample{
		Sentence: []phonemizer_multi.Token{},
	}
	for i, mapping := range sentence {
		log.Now().Debugf("Sentence %d: %v", i, mapping)
		var origword string
		var strkey [][3]string
		for v, k := range mapping {
			if k[0] == 0 {
				origword = strings.TrimRight(v, " ")
				continue
			}
			strkey = append(strkey, [3]string{v, fmt.Sprint(k[0]), fmt.Sprint(k[1])})
		}
		log.Now().Debugf("Origword %d: %s", i, origword)
		sort.SliceStable(strkey, func(i, j int) bool {
			return strkey[i][0] < strkey[j][0]
		})
		var choices [][2]uint32
		for _, v := range strkey {
			num, _ := strconv.Atoi(v[1])
			improvised, _ := strconv.Atoi(v[2])
			if improvised == 1 {
				choices = append(choices, [2]uint32{0, uint32(num)})
			} else {
				choices = append(choices, [2]uint32{hash.StringHash(0, v[0]), uint32(num)})
			}
		}
		sort.SliceStable(choices, func(i, j int) bool {
			return choices[i][0] < choices[j][0]
		})
		var sol uint32
		if len(choices) > 0 {
			sol = choices[0][0]
		}
		//sol = 0
		ai_sentence.Sentence = append(ai_sentence.Sentence, phonemizer_multi.Token{
			Homograph: hash.StringHash(0, origword),
			Choices:   choices,
			Solution:  sol,
		})
	}
	const fanout1 = 24
	for i := range ai_sentence.Sentence {
		var sample = ai_sentence.V2(fanout1, i)
		if sample.Len() <= 1 {
			// no choice
			continue
		}
		var unchosed, chosed [2]uint32
		var accept bool
		for j := 0; !accept && j < sample.Len(); j++ {
			ai_sentence.Sentence[i].Solution = ai_sentence.Sentence[i].Choices[j][0]
			var pred uint32
			if false {
				// old models here
			} else { // newest model
				log.Now().Debugf("Sample IO %d %d: %v", i, j, sample.IO(j).SampleSentence.Sample.Sentence)
				//for feat := 0; feat < fanout1; feat++ {
				//	fmt.Printf("Sample IO %d %d: %d\n", i, j, sample.IO(j).Feature(feat))
				//}
				r.mut.RLock()
				pred = uint32(net.Infer2(sample.IO(j)))
				r.mut.RUnlock()
				log.Now().Debugf("Sample IO pred %d %d: %d", i, j, pred)
			}
			if pred == 1 && !accept {
				accept = true
				chosed = ai_sentence.Sentence[i].Choices[j]
			} else if j == 0 {
				unchosed = ai_sentence.Sentence[i].Choices[j]
			}
		}
		var pred uint32
		if !accept {
			ai_sentence.Sentence[i].Solution = unchosed[0]
			pred = unchosed[1]
		} else {
			ai_sentence.Sentence[i].Solution = chosed[0]
			pred = chosed[1]
		}
		ret = append(ret, [4]uint32{uint32(i), ai_sentence.Sentence[i].Solution, pred, 1})
	}
	if len(ai_sentence.Sentence) > 0 {
		var sample = ai_sentence.V2(fanout1, len(ai_sentence.Sentence)-1)
		log.Now().Debugf("Sample IO final: %v", (&sample).IO(0).SampleSentence.Sample.Sentence)
	}
	return
}

func NewHashtronHomonymSelectorRepository(di *DependencyInjection) *HashtronHomonymSelectorRepository {
	getter := MustAny[interfaces.DictGetter](di)
	hlangs := make(hlanguages)
	return &HashtronHomonymSelectorRepository{
		getter: &getter,
		hlang:  &hlangs,
		mut:    &sync.RWMutex{},
	}
}

var _ IHashtronHomonymSelectorRepository = &HashtronHomonymSelectorRepository{}
