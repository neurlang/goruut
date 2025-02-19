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

func lambdaLoss(lambda *int, wordLoss, decisionComplexityLoss uint64) uint64 {
	if lambda == nil || *lambda < -64 || *lambda > 64 {
		return uint64(wordLoss)
	}
	
	if *lambda < 0 {
		l := uint64(-*lambda)
		return (uint64(wordLoss)) + (uint64(decisionComplexityLoss)<<l)
	} else {
		l := uint64(*lambda)
		return (uint64(wordLoss)<<l) + (uint64(decisionComplexityLoss))
	}
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
	hyperinit := flag.Int("hyperinit", 1, "hyperparameter initial parallelism (nonnegative, high values are slower)")
	hyper := flag.Int("hyper", 100, "hyperparameter parallelism (nonnegative, high values are slower)")
	lambda := flag.Int("rowlossimportance", 5, "hyperparameter row loss importance to reduce decision complexity loss (binary exponent, -64 - 64)")
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
	var written = make(map[[2]string]struct{})
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
			row := [2]string{spacesept(strs[0]), spacesep(strs[1])}
			mut.Lock()
			if _, ok := written[row]; !ok {
				written[row] = struct{}{}
				tsvWriter.AddRow(row[:])
			}
			mut.Unlock()
		}
	}

	lang_eval := lang.ToEval()

	var threeways map[[2]string]uint64
	var rowLoss, complexityLoss atomic.Uint64
	var bestLoss uint64
	var now_hyper int = *hyperinit

	slice := load(*srcFile, 999999999)

again:

	threeways = make(map[[2]string]uint64)
	rowLoss.Store(0)
	complexityLoss.Store(0)

	var delete_key = make([]string, 0, len(lang_eval.Map))
	var delete_langs = make([]*SolutionEval, 0, len(lang_eval.Map))
	var delete_loss = make([]atomic.Uint64, len(lang_eval.Map))
	for i := range delete_loss {
		delete_loss[i].Store(0)
	}
	var delete_complexity_loss = make([]atomic.Uint64, len(lang_eval.Map))
	for i := range delete_complexity_loss {
		delete_complexity_loss[i].Store(0)
	}
	if deleting != nil && *deleting {
		for k := range lang_eval.Map {
			delete_langs = append(delete_langs, lang_eval.WithoutKey(k))
			delete_key = append(delete_key, k)
			if len(delete_key) >= now_hyper {
				break
			}
		}
	}

	var remove_keys = lang_eval.GetValues()

	for len(remove_keys) > now_hyper {
		for k := range remove_keys {
			delete(remove_keys, k)
			break
		}
	}

	var remove_key = make([]string, 0, len(remove_keys))
	var remove_langs = make([]*SolutionEval, 0, len(remove_keys))
	var remove_loss = make([]atomic.Uint64, len(remove_keys))
	for i := range remove_loss {
		remove_loss[i].Store(0)
	}
	var remove_complexity_loss = make([]atomic.Uint64, len(remove_keys))
	for i := range remove_complexity_loss {
		remove_complexity_loss[i].Store(0)
	}
	if deleting != nil && *deleting {
		for k := range remove_keys {
			remove_langs = append(remove_langs, lang_eval.WithoutValue(k))
			remove_key = append(remove_key, k)
		}
	}

	loop(slice, 1000, func(word1, word2 string) {

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
			aligned, cplxloss := lang_single.Align(word1, word2, padspace != nil && *padspace)
			if aligned != nil {
				delete_complexity_loss[i].Add(cplxloss)
			} else {
				delete_loss[i].Add(1)
			}
		}
		// evaluate value removal languages
		for i, lang_single := range remove_langs {
			aligned, cplxloss := lang_single.Align(word1, word2, padspace != nil && *padspace)
			if aligned != nil {
				remove_complexity_loss[i].Add(cplxloss)
			} else {
				remove_loss[i].Add(1)
			}
		}

		// do export alignment
		aligned, cplxloss := lang_eval.Align(word1, word2, padspace != nil && *padspace)
		if aligned != nil {
			complexityLoss.Add(cplxloss)

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
			func(i uint) *uint64 {
				if len(aligned2[1]) > int(i) {
					if aligned2[1][i] == "" {
						return nil
					}
				}
				if len(aligned2[0]) > int(i) {
					if ok := lang_eval.IsDrop(aligned2[0][i]); ok {
						return nil
					}
				}
				//fmt.Println(*x, *y)
				var n uint64
				n = 1
				return &n
			}, nil, func(x *string, y *string) *uint64 {
				if ok := lang_eval.IsEdge(*x,*y); ok {
					return nil
				}
				if *y == "" {
					if ok := lang_eval.IsDrop(*x); ok {
						return nil
					}
				}
				//fmt.Println(*x, *y)
				var n uint64
				n = 1
				return &n
			}, nil)

		//var d = *levenshtein.Distance(mat)

		//fmt.Println(d, aligned2[0], aligned2[1])

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
					if next_to != "" {
						if ok := lang_eval.IsDstMultiPrefix(next_to); !ok {
							//println(threeway_from, threeway_to, next_to)
							threeway_to += next_to
						}
					}
				}

				if threeway_to == "_" && padspace != nil && *padspace {
					// spacer character must be in initial grammar
					return
				}

				if ok := lang_eval.IsDstMultiSuffix(threeway_to); ok || lang_eval.StringStartsWithCombiner(threeway_to) {
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

		fmt.Println("row loss: ",rowLoss.Load(),", complexity loss: ",complexityLoss.Load())

		return
	}
	var minKey, minValue string
	var maxLoss = lambdaLoss(lambda, rowLoss.Load(), complexityLoss.Load()) // we can't beat the loss, but we can meet it.

	fmt.Println(maxLoss, "[", rowLoss.Load(), complexityLoss.Load(), "]", now_hyper)

	if deleting != nil && *deleting {
		for i := range delete_langs {
			if lambdaLoss(lambda, delete_loss[i].Load(), delete_complexity_loss[i].Load()) <= maxLoss {
				maxLoss = lambdaLoss(lambda, delete_loss[i].Load(), delete_complexity_loss[i].Load())
				minKey = delete_key[i]
				minValue = ""
			}
		}
		for i := range remove_loss {
			if lambdaLoss(lambda, remove_loss[i].Load(), remove_complexity_loss[i].Load()) <= maxLoss {
				maxLoss = lambdaLoss(lambda, remove_loss[i].Load(), remove_complexity_loss[i].Load())
				minValue = remove_key[i]
				minKey = ""
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
		
		if len(threewayz) > now_hyper {
			threewayz = threewayz[:now_hyper]
		}
		const individual = false
		var threway_langs = make([]*SolutionEval, 2*len(threewayz), 2*len(threewayz))
		for i := range threewayz {
			if i == 0 {
				threway_langs[i] = lang_eval.With(threewayz[i][0], threewayz[i][1])

				l_del := lang_eval

				if minKey != "" {
					l_del = l_del.WithoutKey(minKey)
					//fmt.Println("removing key", minKey, maxLoss)
				}
				if minValue != "" {
					l_del = l_del.WithoutValue(minValue)
					//fmt.Println("removing value", minValue, maxLoss)
				}

				threway_langs[i+len(threewayz)] = l_del.With(threewayz[i][0], threewayz[i][1])
			} else if individual {
				threway_langs[i] = lang_eval.With(threewayz[i][0], threewayz[i][1])
				threway_langs[len(threewayz)+i] = lang_eval.With(threewayz[i][0], threewayz[i][1])
			} else {
				threway_langs[i] = threway_langs[i-1].With(threewayz[i][0], threewayz[i][1])
				threway_langs[len(threewayz)+i] = lang_eval.With(threewayz[i][0], threewayz[i][1])
			}
		}
		var threeway_loss = make([]atomic.Uint64, len(threway_langs), len(threway_langs))
		var threeway_complexity_loss = make([]atomic.Uint64, len(threway_langs), len(threway_langs))
		for i := range threeway_loss {
			threeway_loss[i].Store(0)
		}
		for i := range threeway_complexity_loss {
			threeway_complexity_loss[i].Store(0)
		}

		loop(slice, 1000, func(word1, word2 string) {

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
				if lang_single == nil {
					continue
				}

				aligned, cplxLoss := lang_single.Align(word1, word2, padspace != nil && *padspace)
				if aligned != nil {
					threeway_complexity_loss[i].Add(cplxLoss)
				} else {
					threeway_loss[i].Add(1)
				}
			}
		})

		var minLoss = maxLoss

		var minI = -1
		for i := range threway_langs {
			if threway_langs[i] == nil {
				continue
			}
			if lambdaLoss(lambda, threeway_loss[i].Load(), threeway_complexity_loss[i].Load()) < minLoss {
				minLoss = lambdaLoss(lambda, threeway_loss[i].Load(), threeway_complexity_loss[i].Load())
				minI = i
			}
		}
		//fmt.Println(minLoss, minI)
		if minI != -1 {
			if minI < len(threewayz) {
				fmt.Println(maxLoss, threewayz[:(minI+1)], minLoss)
				for i := 0; i <= minI; i++ {
					if individual {
						i = minI
					}
					lang_eval = lang_eval.With(threewayz[i][0], threewayz[i][1])
					if (save != nil) && *save && (langFile != nil) && (*langFile != "") {
						lang.With(threewayz[i][0], threewayz[i][1])
					}
				}
			} else {
				if minI == len(threewayz) {

					if minKey != "" {
						lang_eval = lang_eval.WithoutKey(minKey)
						if (save != nil) && *save && (langFile != nil) && (*langFile != "") {
							lang.WithoutKey(minKey)
							fmt.Println(maxLoss, "removing key:", minKey)
						}
					}
					if minValue != "" {
						lang_eval = lang_eval.WithoutValue(minValue)
						if (save != nil) && *save && (langFile != nil) && (*langFile != "") {
							lang.WithoutValue(minValue)
							fmt.Println(maxLoss, "removing value:", minValue)
						}
					}
				}
				//fmt.Println(maxLoss, minLoss, lambdaLoss(lambda,
				//	threeway_loss[minI].Load(),
				//	threeway_complexity_loss[minI].Load()))
				minI -= len(threewayz)
				fmt.Println(maxLoss, threewayz[minI], minLoss)
				lang_eval = lang_eval.With(threewayz[minI][0], threewayz[minI][1])
				if (save != nil) && *save && (langFile != nil) && (*langFile != "") {
					lang.With(threewayz[minI][0], threewayz[minI][1])
				}
			}
		}

		if (save != nil) && *save && (langFile != nil) && (*langFile != "") {
			lang.SaveToJson(*langFile)
		}

		if (bestLoss != 0 && bestLoss == minLoss) || minI == -1 {

			if hyper != nil {
				if now_hyper < *hyper {
					now_hyper <<= 1
				} else {
					return
				}
			} else {
				return
			}
		} else if minI != -1 && now_hyper > *hyperinit {
			now_hyper--
		}
		bestLoss = minLoss
	}
	goto again
}
