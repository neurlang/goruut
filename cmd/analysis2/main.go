package main

import "fmt"
import "flag"
import "strings"
import "sync"
import "sync/atomic"
import "github.com/neurlang/levenshtein"
import "sort"

func spacesep(sli []string) (sep string) {
	for i, w := range sli {
		if i > 0 {
			sep += " "
		}
		sep += w
	}
	return sep
}
func spacesept(sli []string) (sep string) {
	for i, w := range sli {
		if i > 0 {
			sep += " "
		}
		sep += strings.Trim(w, "_")
	}
	return sep
}

func nosep(sli []string) (sep string) {
	for _, w := range sli {
		sep += w
	}
	return sep
}

func main() {
	langFile := flag.String("lang", "", "path to the JSON file containing language data")
	srcFile := flag.String("srcfile", "", "path to input TSV file containing source and target words dictionary")
	dstFile := flag.String("dstfile", "", "path to output TSV file containing source and target phone spaced words dictionary")
	padspace := flag.Bool("padspace", false, "insert space to the end of target word in case of a spaceless written language")
	reverse := flag.Bool("reverse", false, "reverse translation (swap source and target languages)")
	threeway := flag.Bool("threeway", false, "threeway language extension algorithm")
	save := flag.Bool("save", false, "write lang file at the end")
	deleting := flag.Bool("deleting", false, "deleting columns / rows")
	hyper := flag.Int("hyper", 100, "hyperparameter parallelism")
	flag.Parse()

	_ = dstFile

	var lang SolutionFile

	if *langFile != "" {
		err := lang.LoadFromJson(*langFile)
		if err != nil {
			println(err.Error())
			return
		}

	}
	var tsvWriter TSVWriter
	var mut sync.Mutex
	if dstFile != nil && *dstFile != "" {
		err := tsvWriter.Open(*dstFile)
		if err != nil {
			panic(err.Error())
		}
	}

	tsvWrite := func(strs [2][]string) {
		if reverse != nil && *reverse {
			strs[0], strs[1] = strs[1], strs[0]
		}
		if dstFile != nil && *dstFile != "" {
			mut.Lock()
			tsvWriter.AddRow([]string{spacesept(strs[0]), spacesep(strs[1])})
			mut.Unlock()
		}
	}

	lang_eval := lang.ToEval()

	var threeways map[[2]string]uint64
	var rowLoss atomic.Uint64
	var bestLoss uint64
again:

	if bestLoss != 0 && bestLoss == rowLoss.Load() {
		return
	}
	bestLoss = rowLoss.Load()

	threeways = make(map[[2]string]uint64)
	rowLoss.Store(0)

	var delete_key = make([]string, 0, len(lang_eval.Map))
	var delete_langs = make([]*SolutionEval, 0, len(lang_eval.Map))
	var delete_loss = make([]atomic.Uint64, len(lang_eval.Map))
	if deleting != nil && *deleting {
		for k := range lang_eval.Map {
			delete_langs = append(delete_langs, lang_eval.WithoutKey(k))
			delete_key = append(delete_key, k)
			if hyper != nil {
				if len(delete_key) >= *hyper {
					break
				}
			}
		}
	}

	var remove_keys = lang_eval.GetValues()

	if hyper != nil {
		for len(remove_keys) > *hyper {
			for k := range remove_keys {
				delete(remove_keys, k)
				break
			}
		}
	}

	var remove_key = make([]string, 0, len(remove_keys))
	var remove_langs = make([]*SolutionEval, 0, len(remove_keys))
	var remove_loss = make([]atomic.Uint64, len(remove_keys))
	if deleting != nil && *deleting {
		for k := range remove_keys {
			remove_langs = append(remove_langs, lang_eval.WithoutValue(k))
			remove_key = append(remove_key, k)
		}
	}

	loop(*srcFile, 999999999, 1000, func(word1, word2 string) {

		if padspace != nil && *padspace {
			word2 = strings.ReplaceAll(word2, " ", "_")
			word2 += "_"
			for strings.Contains(word2, "__") {
				word2 = strings.ReplaceAll(word2, "__", "_")
			}
		}
		if reverse != nil && *reverse {
			word1, word2 = word2, word1
		}

		// evaluate key deletion languages
		for i, lang_single := range delete_langs {
			aligned := lang_single.Align(word1, word2, padspace != nil && *padspace)
			if aligned != nil {
				continue
			}
			delete_loss[i].Add(1)
		}
		// evaluate value removal languages
		for i, lang_single := range remove_langs {
			aligned := lang_single.Align(word1, word2, padspace != nil && *padspace)
			if aligned != nil {
				continue
			}
			remove_loss[i].Add(1)
		}

		// do export alignment
		aligned := lang_eval.Align(word1, word2, padspace != nil && *padspace)
		if aligned != nil {
			if padspace != nil && *padspace {
				var j = 0
				var k = 0
				if reverse != nil && *reverse {
					k = 1
				}
				for i, val := range (*aligned)[1-k] {
					if i < len((*aligned)[k]) && i < len((*aligned)[1-k]) && strings.HasPrefix(val, "_") {
						toprint := [2][]string{(*aligned)[k][j:i], (*aligned)[1-k][j:i]}
						j = i
						tsvWrite(toprint)
					} else if i+1 < len((*aligned)[k]) && i+1 < len((*aligned)[1-k]) && strings.HasSuffix(val, "_") {
						toprint := [2][]string{(*aligned)[k][j:i+1], (*aligned)[1-k][j:i+1]}
						j = i+1
						tsvWrite(toprint)
					}
				}
				if j < len((*aligned)[k]) && j < len((*aligned)[1-k]) {
					toprint := [2][]string{(*aligned)[k][j:], (*aligned)[1-k][j:]}
					if len(toprint[0]) < len(toprint[1]) {
						toprint[1] = toprint[1][:len(toprint[0])]
					} else {
						toprint[0] = toprint[0][:len(toprint[1])]
					}


					tsvWrite(toprint)
				}
				//if len((*aligned)[0]) == len((*aligned)[1]) {
				return
				//}
				//fmt.Println(len((*aligned)[0]), len((*aligned)[1]))

			} else {
				tsvWrite(*aligned)
				return
			}
		}

		// count error
		rowLoss.Add(1)
		//analyze it
		aligned2 := lang_eval.AlignHybrid(word1, word2)

		var mat = levenshtein.MatrixSlices[uint64, string](aligned2[0], aligned2[1],
			nil, nil, func(x *string, y *string) *uint64 {
				if ok := lang_eval.IsEdge(*x,*y); ok {
					return nil
				}
				if *y == "" {
					if ok := lang_eval.IsDropLast(*x); ok {
						return nil
					}
				}
				//fmt.Println(*x, *y)
				var n uint64
				n = 1
				return &n
			}, nil)

		//var d = *levenshtein.Distance(mat)

		var length = len(aligned2[1]) + 1
		w1p := append(aligned2[0], "")
		w2p := append(aligned2[1], "")

		if threeway != nil && *threeway {
			var bin_length = len(w1p)
			if len(w2p) > bin_length {
				bin_length = len(w2p)
			}

			var bins = make([][]string, bin_length, bin_length)
			var miny = make(map[uint]uint)
			var maxy = make(map[uint]uint)
			var dels = make([]bool, bin_length, bin_length)
			var swaps = make([]*string, bin_length, bin_length)
			levenshtein.Diff(mat, uint(length), func(is_skip, is_insert, is_delete, is_replace bool, x, y uint) bool {

				miny[x] = y
				if _, ok := maxy[x]; ok {
					if maxy[x] < y {
						maxy[x] = y
					}
				} else {
					maxy[x] = y
				}

				if is_skip {

					return true
				}

				if is_replace {
					swaps[x] = &w2p[y]
				}
				if is_insert {
					bins[x] = append(bins[x], w2p[y])
				}
				if is_delete {

					dels[x] = true
				}

				return true
			})
			callback := func(threeway_from, threeway_to, next_to string) {

				if padspace != nil && *padspace && strings.Contains(strings.Trim(threeway_to, "_"), "_") {
					return
				}
				if ok := lang_eval.IsDstMultiPrefix(threeway_to); ok {
					if next_to == "" {
						return
					}
					if ok := lang_eval.IsDstMultiPrefix(next_to); ok {
						return
					}
					//println(threeway_from, threeway_to, next_to)
					threeway_to += next_to
				}

				if threeway_to == "_" && padspace != nil && *padspace {
					// spacer character must be in initial grammar
					return
				}

				if ok := lang_eval.IsDstMultiSuffix(threeway_to); ok || stringStartsWithCombiner(threeway_to) {
					return
				}
				if ok := lang_eval.IsEdge(threeway_from, threeway_to); ok {
					return
				}
				//println(threeway_from, threeway_to)
				mut.Lock()
				for _, w := range lang.Map[threeway_from] {
					if w == threeway_to {
						mut.Unlock()
						return
					}

					if strings.Trim(threeway_to, "_") != strings.Trim(w, "_") &&
						strings.HasSuffix(strings.Trim(threeway_to, "_"), strings.Trim(w, "_")) {
						mut.Unlock()
						return
					}
				}
				threeways[[2]string{threeway_from,threeway_to}]++
				mut.Unlock()
			}
			var resultx, resulty string
			for x := range w1p {
				var bin string
				for xx := range bins[x] {
					bin += (string(bins[x][xx]))
				}

				lookahead := bin
				if len(bin) == 0 {
					minn := miny[uint(x)]
					maxx := maxy[uint(x)]
					if minn > 0 {
						minn--
					}

					lookahead = (string(nosep(w2p[minn:maxx])))
				}
				if len(bins[x]) == 0 && swaps[x] == nil && !dels[x] {
					if resultx != "" && resulty != "" {
						//println(word1, word2, resultx, resulty, lookahead)
						callback(resultx, resulty, lookahead)
					} else if resultx != "" && lookahead != "" {
						//println(word1, word2, resultx, resulty, lookahead)
						callback(resultx, resulty+lookahead, "")
					}
				}

				resulty += (bin)
				//longresulty += (bin)
				if swaps[x] != nil {
					resultx += (string(w1p[x]))
					resulty += (string(*swaps[x]))
				} else if dels[x] {
					resultx += (string(w1p[x]))
				}
				if len(bins[x]) == 0 && swaps[x] == nil && !dels[x] {
					if resultx != "" && resulty != "" {
						callback(resultx, resulty, lookahead)
					} else if resultx != "" && lookahead != "" {
						//println(word1, word2, resultx, resulty, lookahead)
						callback(resultx, resulty+lookahead, "")
					}
					resultx, resulty = "", ""
				}
			}
			for x := len(w1p); x < bin_length; x++ {

				for xx := range bins[x] {
					resulty += (string(bins[x][xx]))
				}
				if resultx != "" && resulty != "" {
					callback(resultx, resulty, "")
				}
			}

		}
	})

	if threeway == nil || false == *threeway {
		return
	}

	var maxLoss = rowLoss.Load() // we can't beat the loss, but we can meet it.
	if deleting != nil && *deleting {
		for i := range delete_langs {
			if delete_loss[i].Load() == maxLoss {
				var minKey = delete_key[i]
				lang_eval = lang_eval.WithoutKey(minKey)
				lang.WithoutKey(minKey)
				fmt.Println("removing key", minKey, maxLoss)
			}
		}
		for i := range remove_loss {
			if remove_loss[i].Load() == maxLoss {
				var minValue = remove_key[i]
				lang_eval = lang_eval.WithoutValue(minValue)
				lang.WithoutValue(minValue)
				fmt.Println("removing value", minValue, maxLoss)
			}
		}

	}

	if len(threeways) == 0 {
		if (save != nil) && *save && (langFile != nil) && (*langFile != "") {
			lang.SaveToJson(*langFile)
		}
		return
	}

	{


		var threewayz [][3]string
		for k, cnt := range threeways {
			threewayz = append(threewayz, [3]string{k[0], k[1], fmt.Sprint(cnt)})
		}
		sort.Slice(threewayz, func(i, j int) bool {
			if len(threewayz[i][2]) == len(threewayz[j][2]) {
				for k := range threewayz[i][2] {
					if threewayz[i][2][k] == threewayz[j][2][k] {
						continue
					}
					return threewayz[i][2][k] > threewayz[j][2][k]
				}
			}
			return len(threewayz[i][2]) > len(threewayz[j][2])
		})
		
		if hyper != nil {
			if len(threewayz) > *hyper {
				threewayz = threewayz[:*hyper]
			}
		}

		var threway_langs = make([]*SolutionEval, len(threewayz), 2*len(threewayz))
		for i := range threewayz {
			if i == 0 {
				threway_langs[i] = lang_eval.With(threewayz[i][0], threewayz[i][1])
			} else if false {
				threway_langs[i] = lang_eval.With(threewayz[i][0], threewayz[i][1])
			} else {
				threway_langs[i] = threway_langs[i-1].With(threewayz[i][0], threewayz[i][1])
				threway_langs = append(threway_langs, lang_eval.With(threewayz[i][0], threewayz[i][1]))
			}
		}

		var threeway_loss = make([]atomic.Uint64, len(threway_langs), len(threway_langs))

		loop(*srcFile, 999999999, 1000, func(word1, word2 string) {

			if padspace != nil && *padspace {
				word2 = strings.ReplaceAll(word2, " ", "_")
				word2 += "_"
				for strings.Contains(word2, "__") {
					word2 = strings.ReplaceAll(word2, "__", "_")
				}
			}
			if reverse != nil && *reverse {
				word1, word2 = word2, word1
			}
			for i, lang_single := range threway_langs {
				aligned := lang_single.Align(word1, word2, padspace != nil && *padspace)
				if aligned != nil {
					continue
				}
				threeway_loss[i].Add(1)
			}
		})

		var minLoss = threeway_loss[0].Load()
		var minI int
		for i := range threeway_loss {
			if threeway_loss[i].Load() < minLoss {
				minLoss = threeway_loss[i].Load()
				minI = i
			}
		}


		if minI < len(threewayz) {
			fmt.Println(maxLoss, threewayz[:(minI+1)], minLoss)
			for i := 0; i <= minI; i++ {
				lang_eval = lang_eval.With(threewayz[i][0], threewayz[i][1])
				if (save != nil) && *save && (langFile != nil) && (*langFile != "") {
					lang.With(threewayz[i][0], threewayz[i][1])
				}
			}
		} else {
			minI -= len(threewayz)
			fmt.Println(maxLoss, threewayz[minI], minLoss)
			lang_eval = lang_eval.With(threewayz[minI][0], threewayz[minI][1])
			if (save != nil) && *save && (langFile != nil) && (*langFile != "") {
				lang.With(threewayz[minI][0], threewayz[minI][1])
			}
		}

		if (save != nil) && *save && (langFile != nil) && (*langFile != "") {
			lang.SaveToJson(*langFile)
		}
	}
	goto again
}
