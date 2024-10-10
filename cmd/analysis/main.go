package main

import "github.com/neurlang/levenshtein"
import "github.com/neurlang/goruut/repo"

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	randv2 "math/rand/v2"
	"os"
	"strings"
	//"time"
)
import (
	"encoding/json"
	"flag"
	"io/ioutil"
)
import (
	"unicode"
	"unicode/utf8"
)
import "sync"

func loop(filename string, group int, do func(string, string)) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	wg := sync.WaitGroup{}
	grp := group

	// Create a new scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		columns := strings.Split(line, "\t")

		// Check if we have exactly two or three columns
		if len(columns) != 2 && len(columns) != 3 {
			fmt.Println("Line does not have exactly two or three columns:", line)
			continue
		}

		// Process each column
		column1 := columns[0]
		column2 := columns[1]

		if grp == 0 {
			wg.Wait()
			grp = group
		} else {
			grp--
		}

		wg.Add(1)

		go func(column1, column2 string) {

			// Example: Print the columns
			do(column1, column2)

			wg.Done()
		}(column1, column2)

	}

	wg.Wait()

	// Check for any scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}

func spacesep(sli []string) (sep string) {
	for i, w := range sli {
		if i > 0 {
			sep += " "
		}
		sep += w
	}
	return sep
}
func nosep(sli []string) (sep string) {
	for _, w := range sli {
		sep += w
	}
	return sep
}

type Language struct {
	Map            map[string][]string `json:"Map"`
	SrcMulti       []string            `json:"SrcMulti"`
	DstMulti       []string            `json:"DstMulti"`
	SrcMultiSuffix []string            `json:"SrcMultiSuffix"`
	DstMultiSuffix []string            `json:"DstMultiSuffix"`
	DstMultiPrefix []string            `json:"DstMultiPrefix"`
	DropLast       []string            `json:"DropLast"`

	SplitBefore []string `json:"SplitBefore"`
	SplitAfter  []string `json:"SplitAfter"`

	PrePhonWordSteps interface{} `json:"PrePhonWordSteps"`
}

func LanguageNewFromFile(file string) (l *Language, err error) {
	// Read the JSON file
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return nil, err
	}

	// Parse the JSON data into the Language struct
	var lang Language
	err = json.Unmarshal(data, &lang)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return nil, err
	}
	return &lang, nil
}

func copyStrings(src []string) (dst []string) {
	for _, v := range src {
		dst = append(dst, v)
	}
	return
}

// isCombiner checks if a rune is a UTF-8 combining character.
func isCombiner(r rune) bool {
	return unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || unicode.Is(unicode.Mc, r)
}

// stringStartsWithCombiner checks if the string starts with a UTF-8 combining character.
func stringStartsWithCombiner(s string) bool {
	if s == "" {
		return false
	}

	r, _ := utf8.DecodeRuneInString(s)
	return isCombiner(r)
}

func init() {
	//rand.Seed(time.Now().UTC().UnixNano())
}

func main() {

	var mut sync.Mutex

	langFile := flag.String("lang", "", "path to the JSON file containing language data")
	save := flag.Bool("save", false, "write lang file at the end")
	srcFile := flag.String("srcfile", "", "path to input TSV file containing source and target words dictionary")
	dstFile := flag.String("dstfile", "", "path to output TSV file containing source and target phone spaced words dictionary")
	hitscnt := flag.Int("hits", 0, "count of hits to add to map")
	randomize := flag.Int("randomize", 0, "randomize dst word split")
	target := flag.Int("target", 0, "unknown words target")
	randinc := flag.Int("randinc", 0, "randomize increasing by making it less frequent using this integer")
	randsubs := flag.Int("randsubs", 0, "randomize word subset using this integer")
	loss := flag.Bool("loss", false, "show edit distance sum (loss, error)")
	spaceBackfit := flag.Bool("spacebackfit", false, "backfit space")
	same := flag.Bool("same", false, "show same matrices")
	wrong := flag.Bool("wrong", false, "print wrong words")
	threeway := flag.Bool("threeway", false, "threeway language extension algorithm")
	nostress := flag.Bool("nostress", false, "delete stress")
	noipadash := flag.Bool("noipadash", false, "delete dash from ipa")
	nospaced := flag.Bool("nospaced", false, "delete spacing")
	padspace := flag.Bool("padspace", false, "insert space to the end of target word in case of a spaceless written language")
	matrices := flag.Bool("matrices", false, "show edit matrices")
	escapeunicode := flag.Bool("escapeunicode", false, "escape unicode when viewing")
	normalize := flag.String("normalize", "", "normalize unicode, for instance to NFC")
	deleteval := flag.Bool("deleteval", false, "delete one value")
	flag.Parse()

	var lang *Language

	if *langFile != "" {
		var err error
		lang, err = LanguageNewFromFile(*langFile)
		if err != nil {
			return
		}

	}
	lang_orig_src_multi := copyStrings(lang.SrcMulti)
	lang_orig_dst_multi := copyStrings(lang.DstMulti)
	lang_orig_src_multi_suffix := copyStrings(lang.SrcMultiSuffix)
	lang_orig_dst_multi_suffix := copyStrings(lang.DstMultiSuffix)

	var dist, unknown uint64

	var dict = make(map[string]struct{})
	var droplast = make(map[string]struct{})
	var drop = make(map[string]struct{})
	var multiprefix = make(map[string]struct{})
	var multisuffix = make(map[string]struct{})
	if lang != nil {
		var deletedval string
		if deleteval != nil && *deleteval {
			var onlyone = randv2.IntN(2) == 0
			var n = randv2.IntN(len(lang.Map)+1) / 2
			for k, v := range lang.Map {
				n--
				if n < 0 {
					n = randv2.IntN(len(v)+1) / 2
					for i, w := range v {
						n--
						if n < 0 {
							if onlyone {
								v[i] = v[len(v)-1]
								v = v[:len(v)-1]
								lang.Map[k] = v
							} else {
								deletedval = w
							}
							break
						}
					}
					break
				}
			}
		}
		for k, v := range lang.Map {

			if len(v) == 1 && v[0] == "" {
				drop[k] = struct{}{}
				continue
			}

			for i := 0; i < len(v); {
				w := v[i]
				for deleteval != nil && *deleteval && w == deletedval {
					if i+1 == len(v) {
						v = v[:len(v)-1]
						lang.Map[k] = v
						break
					} else {
						// Swap with the last item and shrink the slice
						v[i] = v[len(v)-1]
						v = v[:len(v)-1]
						lang.Map[k] = v
						// Do not increment `i` in this case, as the new value at `v[i]` must be checked
					}
					w = v[i] // Update `w` after swapping
				}
				if w != deletedval {
					i++ // Only increment `i` if no deletion happened
				}
			}

			if len(v) == 0 {
				continue
			}

			if len([]rune(k)) > 1 {
				lang.SrcMulti = append(lang.SrcMulti, k)
			}
			for _, w := range v {
				if len([]rune(w)) > 1 {
					lang.DstMulti = append(lang.DstMulti, w)
				}
				dict[k+"\x00"+w] = struct{}{}
			}
		}

		for _, v := range lang.DropLast {
			droplast[v] = struct{}{}
			dict[v+"\x00"] = struct{}{}
		}
		for _, v := range lang.DstMultiPrefix {
			multiprefix[v] = struct{}{}
		}
		for _, v := range lang.DstMultiSuffix {
			multisuffix[v] = struct{}{}
		}
	}

	var longSrcMulti int
	var longDstMulti int
	var longSrcMultiS int
	var longDstMultiS int
	var longDst int

	for _, v := range lang.SrcMulti {
		if len(v) > longSrcMulti {
			longSrcMulti = len(v)
		}
	}
	for _, v := range lang.DstMulti {
		if len(v) > longDstMulti {
			longDstMulti = len(v)
		}
	}
	for _, v := range lang.SrcMultiSuffix {
		if len(v) > longSrcMultiS {
			longSrcMultiS = len(v)
		}
	}
	for _, v := range lang.DstMultiSuffix {
		if len(v) > longDstMultiS {
			longSrcMultiS = len(v)
		}
	}

	if longDstMulti > longDstMultiS {
		longDst = longDstMulti
	} else {
		longDst = longDstMultiS
	}

	srcslice := func(word []rune) (o []string) {
	outer:
		for i := 0; i < len(word); i++ {
			if lang != nil {
				for j := longSrcMulti; j > 0; j-- {
					for _, multi := range lang.SrcMulti {
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
				for j := longSrcMultiS; j > 0; j-- {
					for _, multi := range lang.SrcMultiSuffix {
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

	dstslice_det := func(srcword []string, word string) (o []string) {
		if lang == nil {
			return nil
		}
		o = append(o, "")
	outer:
		for _, srcpre := range srcword {
			if len(word) == 0 {
				break
			}
			mut.Lock()
			data := lang.Map[srcpre]
			mut.Unlock()
			for _, dstpre := range data {
				if _, ok := dict[srcpre+"\x00"+dstpre]; ok {
					if strings.HasPrefix(word, dstpre) {
						word = word[len(dstpre):]
						o = append(o, dstpre)
						continue outer
					}
				}
			}
			for j := longDst; j > 0; j-- {
				if j > len(word) {
					j = len(word)
				}
				if _, ok := multiprefix[word[:j]]; ok {
					continue
				}
				if _, ok := dict[srcpre+"\x00"+word[:j]]; ok {
					o = append(o, word[:j])
					word = word[j:]
					continue outer
				}
			}
		consume_suf:
			for len(word) > 0 {
				for len(word) > 0 && stringStartsWithCombiner(word) {
					o[len(o)-1] += string([]rune(word)[0])
					word = string([]rune(word)[1:])
				}
				for suf := range multisuffix {
					for strings.HasPrefix(word, suf) {
						o[len(o)-1] += suf
						word = word[len(suf):]
					}
				}
				for suf := range multisuffix {
					if strings.HasPrefix(word, suf) {
						continue consume_suf
					}
				}
				break
			}

			if len(word) == 0 {
				break
			}
			if _, ok := multiprefix[string([]rune(word)[0])]; ok && len([]rune(word)) >= 2 {
				o = append(o, string([]rune(word)[0:2]))
				word = string([]rune(word)[2:])
			} else {
				o = append(o, string([]rune(word)[0]))
				word = string([]rune(word)[1:])
			}
		consume_pref:
			for len(word) > 0 {
				for len(word) > 0 && stringStartsWithCombiner(word) {
					o[len(o)-1] += string([]rune(word)[0])
					word = string([]rune(word)[1:])
				}
				for pref := range multiprefix {
					for strings.HasPrefix(word, pref) {
						o[len(o)-1] += pref
						word = word[len(pref):]
					}
				}
				for pref := range multiprefix {
					if strings.HasPrefix(word, pref) {
						continue consume_pref
					}
				}
				break
			}
		}
		var aligno []string
		for len(word) > 0 {
		consume_suf2:
			for len(word) > 0 {
				for len(word) > 0 && stringStartsWithCombiner(word) {
					o[len(o)-1] += string([]rune(word)[0])
					word = string([]rune(word)[1:])
				}
				for suf := range multisuffix {
					for strings.HasPrefix(word, suf) {
						o[len(o)-1] += suf
						word = word[len(suf):]
					}
				}
				for suf := range multisuffix {
					if strings.HasPrefix(word, suf) {
						continue consume_suf2
					}
				}
				break
			}
			if aligno == nil {
				aligno = o
			}
			if len(word) == 0 {
				break
			}
			if _, ok := multiprefix[string([]rune(word)[0])]; ok && len([]rune(word)) >= 2 {
				o = append(o, string([]rune(word)[0:2]))
				word = string([]rune(word)[2:])
			} else {
				o = append(o, string([]rune(word)[0]))
				word = string([]rune(word)[1:])
			}
		consume_pref2:
			for len(word) > 0 {
				for len(word) > 0 && stringStartsWithCombiner(word) {
					o[len(o)-1] += string([]rune(word)[0])
					word = string([]rune(word)[1:])
				}
				for pref := range multiprefix {
					for strings.HasPrefix(word, pref) {
						o[len(o)-1] += pref
						word = word[len(pref):]
					}
				}
				for pref := range multiprefix {
					if strings.HasPrefix(word, pref) {
						continue consume_pref2
					}
				}
				break
			}
		}

		if o[0] == "" {
			o = o[1:]
		}
		if 1*len(o) > len(srcword)*5 {
			//println("WARN: ", nosep(srcword), nosep(o), " Significantly longer: ", nosep(aligno))
			return nil
		}
		//println(spacesep(o))
		return
	}

	dstslice := func(word []rune) (o []string) {
	outer:
		for i := 0; i < len(word); i++ {
			if lang != nil {
				for j := longDstMulti; j > 0; j-- {
					for _, multi := range lang.DstMulti {
						if len(multi) != j {
							continue
						}
						if _, ok := multiprefix[multi]; ok {
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
				for j := longDstMultiS; j > 0; j-- {
					for _, multi := range lang.DstMultiSuffix {
						if len(multi) != j {
							continue
						}
						if _, ok := multiprefix[multi]; ok {
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

	var tsvWriter TSVWriter
	if dstFile != nil && *dstFile != "" {
		err := tsvWriter.Open(*dstFile)
		if err != nil {
			panic(err.Error())
		}
	}

	var threeways = make(map[string]uint64)

	loop(*srcFile, 200, func(word1, word2 string) {

		if randsubs != nil && *randsubs != 0 {
			if rand.Intn(1+*randsubs) != 0 {
				return
			}
		}

		if normalize != nil && *normalize != "" {
			word1 = repo.NormalizeTo(word1, *normalize)
		}

		if nostress != nil && *nostress {
			word2 = strings.ReplaceAll(word2, "ˈ", "")
			word2 = strings.ReplaceAll(word2, "'", "")
			word2 = strings.ReplaceAll(word2, "ˌ", "")
		}

		if padspace != nil && *padspace {
			word2 = strings.ReplaceAll(word2, " ", "_")
			word2 += "_"
		}

		if nospaced != nil && *nospaced {
			word2 = strings.ReplaceAll(word2, " ", "")
			word1 = strings.ReplaceAll(word1, " ", "")
		}

		if noipadash != nil && *noipadash {
			word2 = strings.ReplaceAll(word2, "-", "")
		}

		srcword := srcslice([]rune(word1))

		var dstwordGreedy []string
		word2pref := word2

		var tokens = 0
		if randomize != nil {
			tokens = *randomize
		}

	outer:
		for i := 0; i < len(srcword); i++ {
			src := srcword[i]
			var j = randv2.IntN(1 + longDst)
			if j > len(word2pref) {
				continue
			}
			if _, ok := dict[src+"\x00"+word2pref[:j]]; ok {
				dstwordGreedy = append(dstwordGreedy, word2pref[:j])
				//println(src, word2pref[:j])
				word2pref = word2pref[j:]
				continue outer
			}
			if _, ok := drop[src]; ok {
				dstwordGreedy = append(dstwordGreedy, "")
				continue outer
			}
			if tokens > 0 {
				i = -1
				dstwordGreedy = nil
				word2pref = word2
				tokens--
			} else {
				break
			}
		}
		dstword := dstslice_det(srcword, word2)
		if dstword == nil {
			return
		}
		if len(dstword) == 0 {
			dstword = dstslice([]rune(word2))
		}
		if len(srcword) > 0 {
			if _, isDropLast := droplast[srcword[len(srcword)-1]]; isDropLast {
				if len(dstword)+1 == len(srcword) {
					dstword = append(dstword, "")
				}
				if len(dstwordGreedy)+1 == len(srcword) {
					dstwordGreedy = append(dstwordGreedy, "")
				}
			}
		}
		if len(srcword) == len(dstwordGreedy) {
			dstword = dstwordGreedy
		}
		var mat = levenshtein.MatrixSlices[uint64, string](srcword, dstword,
			func(i uint) *uint64 {
				if len(srcword) > int(i) {
					if _, ok := drop[srcword[i]]; ok {
						return nil
					}
				}
				var n uint64
				n = 1
				return &n
			}, nil, func(x *string, y *string) *uint64 {
				if _, ok := dict[*x+"\x00"+*y]; ok {
					return nil
				}
				if *y == "" {
					if _, ok := drop[*x]; ok {
						return nil
					}
				}
				//fmt.Println(*x, *y)
				var n uint64
				n = 1
				return &n
			}, nil)

		var d = *levenshtein.Distance(mat)

		var length = len(dstword) + 1
		w1p := append(srcword, "")
		w2p := append(dstword, "")

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
				if _, ok := multiprefix[threeway_to]; ok {
					if next_to == "" {
						return
					}
					if _, ok := multiprefix[next_to]; ok {
						return
					}
					//println(threeway_from, threeway_to, next_to)
					threeway_to += next_to
				}

				if threeway_to == "_" && padspace != nil && *padspace {
					// spacer character must be in initial grammar
					return
				}

				if _, ok := multisuffix[threeway_to]; ok || stringStartsWithCombiner(threeway_to) {
					return
				}
				if _, ok := dict[threeway_from+"\x00"+threeway_to]; ok {
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
				if hitscnt != nil && uint64(*hitscnt) == threeways[threeway_from+"\x00"+threeway_to] {
					//println(threeway_from, threeway_to)
					lang.Map[threeway_from] = append(lang.Map[threeway_from], threeway_to)
				} else if hitscnt != nil && uint64(*hitscnt) > threeways[threeway_from+"\x00"+threeway_to] {
					if randinc == nil || *randinc == 0 || randv2.IntN(*randinc) == 0 {
						threeways[threeway_from+"\x00"+threeway_to]++
					}
				}
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

		if d > 0 && matrices != nil && *matrices {
			if (same != nil && *same && len(w1p) == len(w2p)) || (same == nil) || (same != nil && !*same && len(w1p) != len(w2p)) {
				mut.Lock()
				if escapeunicode != nil && *escapeunicode {
					for _, rs := range w2p {
						for _, r := range rs {
							fmt.Fprintf(os.Stderr, "\\u%04X", r)
						}
						fmt.Fprint(os.Stderr, " ")
					}
				} else {
					for _, rs := range w2p {
						fmt.Fprintf(os.Stderr, "%s ", rs)
					}
				}
				fmt.Fprintln(os.Stderr)
				for i := 0; i+length <= len(mat); i += length {
					fmt.Fprintln(os.Stderr, w1p[i/length], mat[i:i+length])
				}
				fmt.Fprintln(os.Stderr, d)
				mut.Unlock()
			}
		}
		if d == 0 && (spaceBackfit == nil || !*spaceBackfit) {

			for _, v := range dstword {
				if strings.Contains(v, `"`) {
					panic(v)
				}
			}

			if padspace != nil && *padspace {

				var lasti = 0
				for i, w := range dstword {
					if strings.HasPrefix(w, "_") && i > lasti {
						mut.Lock()
						tsvWriter.AddRow([]string{spacesep(srcword[lasti:i]), spacesep(dstword[lasti:i])})
						mut.Unlock()
						lasti = i
					}
					if strings.HasSuffix(w, "_") && i+1 > lasti {
						mut.Lock()
						tsvWriter.AddRow([]string{spacesep(srcword[lasti : i+1]), spacesep(dstword[lasti : i+1])})
						mut.Unlock()
						lasti = i + 1
					}
				}
				if lasti != len(dstword) {
					mut.Lock()
					tsvWriter.AddRow([]string{spacesep(srcword[lasti:]), spacesep(dstword[lasti:])})
					mut.Unlock()
				}

			} else {

				mut.Lock()
				tsvWriter.AddRow([]string{spacesep(srcword), spacesep(dstword)})
				mut.Unlock()

			}
		} else if wrong != nil && (*wrong) {
			fmt.Println(word1, word2)

		}
		if d > 0 {
			mut.Lock()
			unknown++
			mut.Unlock()
		}

		mut.Lock()
		dist += d
		mut.Unlock()
	})

	tsvWriter.Close()
	if (threeway != nil) && (*threeway) || (deleteval != nil) && (*deleteval) {

		if hitscnt != nil && *hitscnt > 0 {
			var maxv uint64
			//var maxmapping string

			for _, v := range threeways {
				if v > maxv {
					maxv = v
					//maxmapping = k
				}
			}
			if maxv < uint64(*hitscnt) {
				//println("Decrease -hits to:", maxv, "adding best match to language:", maxmapping)
				fmt.Println("Decrease hits to:", maxv)
			}
		}

		if (save != nil) && *save {

			lang.SrcMulti = lang_orig_src_multi
			lang.DstMulti = lang_orig_dst_multi
			lang.SrcMultiSuffix = lang_orig_src_multi_suffix
			lang.DstMultiSuffix = lang_orig_dst_multi_suffix

			data, err := json.Marshal(lang)
			if err != nil {
				fmt.Println(err.Error())
			}

			data = bytes.ReplaceAll(data, []byte(`],"`), []byte("],\n\""))

			err = ioutil.WriteFile(*langFile, data, 0755)
			if err != nil {
				fmt.Println(err.Error())
			}

		} else {

			data, err := json.Marshal(lang.Map)

			if err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println(string(data))
			}
		}
	}
	if loss != nil && *loss {
		if randsubs != nil && *randsubs != 0 {
			fmt.Println("Edit distance is:", dist*(1+uint64(*randsubs)))
		} else {
			fmt.Println("Edit distance is:", dist)
		}
	}
	if target != nil && *target > 0 {
		unknown *= (1 + uint64(*randsubs))
		if unknown < uint64(*target) {
			fmt.Println("Unknown words:", unknown)
		}
	}
}
