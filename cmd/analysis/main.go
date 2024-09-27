package main

import "github.com/neurlang/levenshtein"
import "github.com/neurlang/goruut/repo"

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)
import (
	"encoding/json"
	"flag"
	"io/ioutil"
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

		// Check if we have exactly two columns
		if len(columns) != 2 {
			fmt.Println("Line does not have exactly two columns:", line)
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
	DropLast       []string            `json:"DropLast"`

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

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {

	var mut sync.Mutex

	langFile := flag.String("lang", "", "path to the JSON file containing language data")
	save := flag.Bool("save", false, "write lang file at the end")
	srcFile := flag.String("srcfile", "", "path to input TSV file containing source and target words dictionary")
	dstFile := flag.String("dstfile", "", "path to output TSV file containing source and target phone spaced words dictionary")
	hitscnt := flag.Int("hits", 0, "count of hits to add to map")
	randomize := flag.Int("randomize", 0, "randomize dst word split")
	randadd := flag.Int("randadd", 1, "randomize adding by making it less frequent using this integer")
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

	var dist uint64

	var dict = make(map[string]struct{})
	var droplast = make(map[string]struct{})
	var drop = make(map[string]struct{})
	if lang != nil {
		var deletedval string
		if deleteval != nil && *deleteval {
			var onlyone = rand.Intn(2) == 0
			var n = rand.Intn(len(lang.Map)+1) / 2
			for k, v := range lang.Map {
				n--
				if n < 0 {
					n = rand.Intn(len(v)+1) / 2
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

			for i, w := range v {
				for deleteval != nil && *deleteval && w == deletedval {
					if i+1 == len(v) {
						v = v[:len(v)-1]
						lang.Map[k] = v
						break
					} else {
						v[i] = v[len(v)-1]
						v = v[:len(v)-1]
						w = v[i]
						lang.Map[k] = v
					}
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
	dstslice := func(word []rune) (o []string) {
	outer:
		for i := 0; i < len(word); i++ {
			if lang != nil {
				for j := longDstMulti; j > 0; j-- {
					for _, multi := range lang.DstMulti {
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
				for j := longDstMultiS; j > 0; j-- {
					for _, multi := range lang.DstMultiSuffix {
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

	var tsvWriter TSVWriter
	if dstFile != nil && *dstFile != "" {
		err := tsvWriter.Open(*dstFile)
		if err != nil {
			panic(err.Error())
		}
	}

	var threeways = make(map[string]int)

	loop(*srcFile, 200, func(word1, word2 string) {

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
			var j = rand.Intn(1 + longDst)
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
		dstword := dstslice([]rune(word2))
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
			var dels = make([]bool, bin_length, bin_length)
			var swaps = make([]*string, bin_length, bin_length)
			levenshtein.Diff(mat, uint(length), func(is_skip, is_insert, is_delete, is_replace bool, x, y uint) bool {
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
			callback := func(threeway_from, threeway_to string) {

				if padspace != nil && *padspace && strings.Contains(strings.Trim(threeway_to, "_"), "_") {
					return
				}

				//println(threeway_from, threeway_to)
				mut.Lock()
				for _, w := range lang.Map[threeway_from] {
					if w == threeway_to {
						mut.Unlock()
						return
					}
				}
				threeways[threeway_from+"\x00"+threeway_to]++
				if hitscnt != nil && *hitscnt == threeways[threeway_from+"\x00"+threeway_to] {
					if randadd == nil || rand.Intn(*randadd) == 0 {
						lang.Map[threeway_from] = append(lang.Map[threeway_from], threeway_to)
					} else {
						delete(threeways, threeway_from+"\x00"+threeway_to)
					}
				}
				mut.Unlock()
			}
			var resultx, resulty string
			for x := range w1p {
				if len(bins[x]) == 0 && swaps[x] == nil && !dels[x] {
					if resultx != "" && resulty != "" {
						callback(resultx, resulty)
					}
					resultx, resulty = "", ""
				}

				for xx := range bins[x] {
					resulty += (string(bins[x][xx]))
				}
				if swaps[x] != nil {
					resultx += (string(w1p[x]))
					resulty += (string(*swaps[x]))
				} else if dels[x] {
					resultx += (string(w1p[x]))
				}
			}
			for x := len(w1p); x < bin_length; x++ {

				for xx := range bins[x] {
					resulty += (string(bins[x][xx]))
				}
			}
			if resultx != "" && resulty != "" {
				callback(resultx, resulty)
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

			mut.Lock()
			tsvWriter.AddRow([]string{spacesep(srcword), spacesep(dstword)})
			mut.Unlock()
		} else if wrong != nil && (*wrong) {
			fmt.Println(word1, word2)
		}

		mut.Lock()
		dist += d
		mut.Unlock()
	})

	tsvWriter.Close()
	if (threeway != nil) && (*threeway) || (deleteval != nil) && (*deleteval) {

		if hitscnt != nil && *hitscnt > 0 {
			var maxv int
			//var maxmapping string

			for _, v := range threeways {
				if v > maxv {
					maxv = v
					//maxmapping = k
				}
			}
			if maxv < *hitscnt {
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
		fmt.Println("Edit distance is:", dist)
	}
}
