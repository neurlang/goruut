package main

import "github.com/neurlang/levenshtein"
import "github.com/neurlang/goruut/repo"

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"regexp"
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

type Spacer struct {
	Spacer []struct {
		LeftRegexp  string `json:"LeftRegexp"`
		RightRegexp string `json:"RightRegexp"`
		left, right *regexp.Regexp
	}
	List []string
}

func SpacerNewFromFile(file string) (l *Spacer, err error) {
	// Read the JSON file
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return nil, err
	}

	// Parse the JSON data into the Language struct
	var spacer Spacer
	err = json.Unmarshal(data, &spacer)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return nil, err
	}
	for i := range spacer.Spacer {
		l, err := regexp.Compile(spacer.Spacer[i].LeftRegexp)
		if err != nil {
			fmt.Printf("Error compiling regexp: %v\n", err)
			return nil, err
		}
		r, err := regexp.Compile(spacer.Spacer[i].RightRegexp)
		if err != nil {
			fmt.Printf("Error compiling regexp: %v\n", err)
			return nil, err
		}
		spacer.Spacer[i].left, spacer.Spacer[i].right = l, r
	}
	return &spacer, nil
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {

	var mut sync.Mutex

	langFile := flag.String("lang", "", "path to the JSON file containing language data")
	save := flag.Bool("save", false, "write lang file at the end")
	spacerFile := flag.String("spacerfile", "", "path to the JSON file containing spacer data")
	srcFile := flag.String("srcfile", "", "path to input TSV file containing source and target words dictionary")
	dstFile := flag.String("dstfile", "", "path to output TSV file containing source and target phone spaced words dictionary")
	hints := flag.Bool("hints", false, "display language file improvements hints")
	hitscnt := flag.Int("hits", 0, "count of hits to add to map")
	randomize := flag.Int("randomize", 0, "randomize dst word split")
	loss := flag.Bool("loss", false, "show edit distance sum (loss, error)")
	spaceBackfit := flag.Bool("spacebackfit", false, "backfit space")
	same := flag.Bool("same", false, "show same matrices")
	join := flag.Bool("join", false, "join letters")
	wrong := flag.Bool("wrong", false, "print wrong words")
	prolong := flag.Bool("prolong", false, "prolong mistaken prefix")
	threeway := flag.Bool("threeway", false, "threeway language extension algorithm")
	nostress := flag.Bool("nostress", false, "delete stress")
	noipadash := flag.Bool("noipadash", false, "delete dash from ipa")
	nospaced := flag.Bool("nospaced", false, "delete spacing")
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
	lang_orig_src_multi := lang.SrcMulti
	lang_orig_dst_multi := lang.DstMulti

	var spacer *Spacer
	if *spacerFile != "" {
		var err error
		spacer, err = SpacerNewFromFile(*spacerFile)
		if err != nil {
			return
		}

	}

	var dist float32

	var dict = make(map[string]struct{})
	var droplast = make(map[string]struct{})
	if lang != nil {
		var deletedval string
		if deleteval != nil && *deleteval {
			var n = rand.Intn(len(lang.Map)+1) / 2
			for _, v := range lang.Map {
				n--
				if n < 0 {
					n = rand.Intn(len(v)+1) / 2
					for _, w := range v {
						n--
						if n < 0 {
							deletedval = w
							break
						}
					}
					break
				}
			}
		}
		for k, v := range lang.Map {

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
	if dstFile != nil {
		tsvWriter.Open(*dstFile, nil)
	}

	var hits = make(map[string]int)
	var joins = make(map[string]int)
	var prolongs = make(map[string]int)
	var threeways = make(map[string]int)

	loop(*srcFile, 100, func(word1, word2 string) {

		if normalize != nil && *normalize != "" {
			word1 = repo.NormalizeTo(word1, *normalize)
		}

		if nostress != nil && *nostress {
			word2 = strings.ReplaceAll(word2, "ˈ", "")
			word2 = strings.ReplaceAll(word2, "'", "")
			word2 = strings.ReplaceAll(word2, "ˌ", "")
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
			var j = 1 + rand.Intn(longDst)
			if j > len(word2pref) {
				continue
			}
			if _, ok := dict[src+"\x00"+word2pref[:j]]; ok {
				dstwordGreedy = append(dstwordGreedy, word2pref[:j])
				//println(src, word2pref[:j])
				word2pref = word2pref[j:]
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
		var mat = levenshtein.MatrixTSlices[float32, string](srcword, dstword,
			nil, nil, func(x *string, y *string) *float32 {
				if _, ok := dict[*x+"\x00"+*y]; ok {
					return nil
				}

				//fmt.Println(*x, *y)
				var n float32
				n = 1
				return &n
			}, nil)

		var d = *levenshtein.Distance(mat)

		var length = len(srcword) + 1
		w1p := append(srcword, "")
		w2p := append(dstword, "")

		if d == 1 && threeway != nil && *threeway {
			var lastx, lasty uint
			levenshtein.WalkVals(mat, uint(length), func(prev, this float32, x, y uint) bool {

				if x == 0 || y == 0 {
					return false
				}
				//println(x, y, this,  w1p[x-1],  w2p[y-1])

				if this == 0 {
					if lastx == 0 {
						lastx = x
					}
					if lasty == 0 {
						lasty = y
					}
					var threeway_from, threeway_to string
					var lendiff = 1 + uint(len(srcword)) - uint(len(dstword))
					code := (lastx - x) | (lasty-y)<<1 | lendiff<<2
					switch code {
					// merging ipa
					case 0:
						threeway_from = w1p[x-1] + w1p[x]
						if y >= 2 {
							threeway_to = w2p[y-2] + w2p[y-1] + w2p[y]
						} else {
							threeway_to = w2p[y-1] + w2p[y]
						}
					case 1:
						threeway_from = w1p[x]
						threeway_to = w2p[y-1] + w2p[y]
					case 2:
						threeway_from = w1p[x-1] + w1p[x]
						threeway_to = w2p[y] + w2p[y+1]
					case 3:
						threeway_from = w1p[x-1] + w1p[x]
						threeway_to = w2p[y-1] + w2p[y]
					// neutral
					case 4:
						threeway_from = w1p[x-1] + w1p[x]
						threeway_to = w2p[y-1] + w2p[y]
					case 5:
						threeway_from = w1p[x-1] + w1p[x]
						threeway_to = w2p[y-1] + w2p[y]
					case 6:
						threeway_from = w1p[x-1] + w1p[x]
						threeway_to = w2p[y-1] + w2p[y]
					case 7:
						threeway_from = w1p[x-1] + w1p[x]
						threeway_to = w2p[y-1] + w2p[y]
					//  merging text
					case 8:
						if x >= 2 {
							threeway_from = w1p[x-2] + w1p[x-1] + w1p[x]
						} else {
							threeway_from = w1p[x-1] + w1p[x]
						}
						threeway_to = w2p[y-1] + w2p[y]
					case 9:
						threeway_from = w1p[x-1] + w1p[x]
						threeway_to = w2p[y-1] + w2p[y]
					case 10:
						threeway_from = w1p[x-1] + w1p[x]
						threeway_to = w2p[y-1]
					case 11:
						threeway_from = w1p[x-1] + w1p[x]
						threeway_to = w2p[y-1] + w2p[y]
					}

					//fmt.Println(code, word1, word2, "|", spacesep(srcword), "|", spacesep(dstword), "|",  x, y, lastx, lasty, threeway_from, threeway_to)

					mut.Lock()
					for _, w := range lang.Map[threeway_from] {
						if w == threeway_to {
							mut.Unlock()
							return true
						}
					}
					threeways[threeway_from+"\x00"+threeway_to]++
					if hitscnt != nil && *hitscnt < threeways[threeway_from+"\x00"+threeway_to] {
						lang.Map[threeway_from] = append(lang.Map[threeway_from], threeway_to)
					}
					mut.Unlock()
				} else {
					lastx, lasty = x, y
				}

				return this == 0
			})
		}

		if d > 0 && prolong != nil && *prolong {
			levenshtein.WalkVals(mat, uint(length), func(prev, this float32, x, y uint) bool {
				if x == 0 || y == 0 {
					return false
				}
				prolonged_from := w1p[x-1] + w1p[x]
				prolonged_to := w2p[y-1]
				if prev != 0 && this == 0 {
					var w1after string
					for _, after := range w1p[x+1:] {
						w1after += after
					}
					var w2after string
					for _, after := range w2p[y:] {
						w2after += after
					}

					srcwordp := srcslice([]rune(w1after))
					dstwordp := dstslice([]rune(w2after))
					var matp = levenshtein.MatrixTSlices[float32, string](srcwordp, dstwordp,
						nil, nil, func(x *string, y *string) *float32 {
							if _, ok := dict[*x+"\x00"+*y]; ok {
								return nil
							}

							//fmt.Println(*x, *y)
							var n float32
							n = 1
							return &n
						}, nil)

					var dp = *levenshtein.Distance(matp)

					if dp == 0 {
						for _, w := range lang.Map[prolonged_from] {
							if w == prolonged_to {
								return true
							}
						}
						prolongs[prolonged_from+"\x00"+prolonged_to]++
						if hitscnt != nil && *hitscnt < prolongs[prolonged_from+"\x00"+prolonged_to] {
							lang.Map[prolonged_from] = append(lang.Map[prolonged_from], prolonged_to)
						}
						return true
					}
				}
				return !(prev == 0 && this == 0)
			})
		}

		if d > 0 && matrices != nil && *matrices {
			if (same != nil && *same && len(w1p) == len(w2p)) || (same == nil) || (same != nil && !*same && len(w1p) != len(w2p)) {
				mut.Lock()
				if escapeunicode != nil && *escapeunicode {
					for _, rs := range w1p {
						for _, r := range rs {
							fmt.Printf("\\u%04X", r)
						}
						fmt.Print(" ")
					}
				} else {
					for _, rs := range w1p {
						fmt.Printf("%s ", rs)
					}
				}
				fmt.Println()
				for i := 0; i+length-1 < len(mat); i += length {
					fmt.Println(w2p[i/length], mat[i:i+length])
				}
				fmt.Println(d)
				mut.Unlock()
			}
		}
		if d == 0 && (spaceBackfit == nil || !*spaceBackfit) {
			mut.Lock()
			tsvWriter.AddRow([]string{spacesep(srcword), spacesep(dstword)})
			mut.Unlock()
		} else if wrong != nil && (*wrong) {
			fmt.Println(word1, word2)
		}
		var final_froms, final_tos []string
		if (same != nil && *same && len(w1p) == len(w2p)) || (same == nil) || (same != nil && !*same && len(w1p) != len(w2p)) {
			levenshtein.WalkVals(mat, uint(length), func(prev, this float32, x, y uint) bool {
				if x == 0 || y == 0 {
					return true
				}

				if _, ok := dict[w1p[x]+"\x00"+w2p[y]]; ok {
					return false
				}
				if hints != nil && *hints {

					if prev == this {
						return false
					}

					// cross add

					final_froms = append(final_froms, w1p[x])
					final_tos = append(final_tos, w2p[y])
					final_froms = append(final_froms, w1p[x-1])
					final_tos = append(final_tos, w2p[y])
					final_froms = append(final_froms, w1p[x])
					final_tos = append(final_tos, w2p[y-1])
					final_froms = append(final_froms, w1p[x-1])
					final_tos = append(final_tos, w2p[y-1])

				}
				return false
			})
		}

		if d > 0 && (join != nil) && (*join) {
			//println(word1, " ", word2)
			levenshtein.WalkVals(mat, uint(length), func(prev, this float32, x, y uint) bool {
				if prev != 0 && this == 0 && uint(len(w1p)) > x+1 && uint(len(w2p)) > y+1 {
					joined_from := w1p[x] + w1p[x+1]
					joined_to := w2p[y] + w2p[y+1]

					for _, w := range lang.Map[joined_from] {
						if w == joined_to {
							return false
						}
					}
					joins[joined_from+"\x00"+joined_to]++
					if hitscnt != nil && *hitscnt < joins[joined_from+"\x00"+joined_to] {
						lang.Map[joined_from] = append(lang.Map[joined_from], joined_to)
					}
					return true
				}

				return false
			})
		}

		if spaceBackfit != nil && *spaceBackfit {
			var output []string
			var q = 0
			for i := range w1p {
				left, right := string(nosep(w1p[q:i])), string(nosep(w1p[i:]))
				for j := 0; j < len(spacer.Spacer); j++ {
					l := spacer.Spacer[j].left
					r := spacer.Spacer[j].right
					lmatch := l != nil && l.MatchString(left) || l == nil
					rmatch := r != nil && r.MatchString(right) || r == nil
					if lmatch && rmatch {
						output = append(output, left)
						q = i
					}
				}
			}
			output = append(output, string(nosep(w1p[q:])))
		outerr:
			for i := 0; i < len(output); i++ {
				for j := len([]rune(output[i])) - 1; j > 0; j-- {
					left := string([]rune(output[i])[:j])
					right := string([]rune(output[i])[j:])
					for j := range spacer.List {
						w := strings.Split(spacer.List[j], " ")
						if left != w[0] {
							continue
						}
						prefix := w[1]
						if prefix != "" && strings.HasPrefix(right, prefix) {
							output = append(output[:i+1], output[i:]...)
							output[i] = left
							output[i+1] = right
							i--
							continue outerr
						}
					}
				}
			}

			var backx = uint(len(w1p))
			var backfitted string
			levenshtein.WalkVals(mat, uint(length), func(prev, this float32, x, y uint) bool {

				if x < backx {
					//println(w1p[x], w2p[y])

					backfitted = w1p[x] + backfitted
					backx = x
				}

				if w2p[y] == " " && backfitted != "" && backfitted[0] != ' ' {
					backfitted = " " + backfitted
				}

				return false
			})
			mut.Lock()
			tsvWriter.AddRow([]string{backfitted, spacesep(output)})
			mut.Unlock()
		}

		if hints != nil && *hints && len(final_froms) != 0 && len(final_tos) != 0 {

			for ff, final_from := range final_froms {
				final_to := final_tos[ff]
				if final_from == "" {
					continue
				}
				if final_to == "" {
					continue
				}

				var found bool
				mut.Lock()
				for _, w := range lang.Map[final_from] {
					if w == final_to {
						found = true
						break
					}
				}
				if !found {
					hits[final_from+"\x00"+final_to]++
					if hitscnt != nil && *hitscnt < hits[final_from+"\x00"+final_to] {
						lang.Map[final_from] = append(lang.Map[final_from], final_to)
					}
				}
				mut.Unlock()

			}
		}
		mut.Lock()
		dist += d
		mut.Unlock()
	})

	tsvWriter.Close()
	if (hints != nil && *hints) || (join != nil) && (*join) || (prolong != nil) && (*prolong) ||
		(threeway != nil) && (*threeway) || (deleteval != nil) && (*deleteval) {

		if (save != nil) && *save {

			lang.SrcMulti = lang_orig_src_multi
			lang.DstMulti = lang_orig_dst_multi

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
