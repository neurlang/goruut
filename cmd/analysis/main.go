package main

import "github.com/neurlang/levenshtein"
import "github.com/neurlang/goruut/repo"

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)
import (
	"encoding/json"
	"flag"
	"io/ioutil"
)

func loop(filename string, do func(string, string)) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

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

		// Example: Print the columns
		do(column1, column2)

	}

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

type Language struct {
	Map            map[string][]string `json:"Map"`
	SrcMulti       []string            `json:"SrcMulti"`
	DstMulti       []string            `json:"DstMulti"`
	SrcMultiSuffix []string            `json:"SrcMultiSuffix"`
	DstMultiSuffix []string            `json:"DstMultiSuffix"`
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

func main() {

	langFile := flag.String("lang", "", "path to the JSON file containing language data")
	srcFile := flag.String("srcfile", "", "path to input TSV file containing source and target words dictionary")
	dstFile := flag.String("dstfile", "", "path to output TSV file containing source and target phone spaced words dictionary")
	hints := flag.Bool("hints", false, "display language file improvements hints")
	hitscnt := flag.Int("hits", 0, "count of hits to add to map")
	loss := flag.Bool("loss", false, "show edit distance sum (loss, error)")
	same := flag.Bool("same", false, "show same matrices")
	join := flag.Bool("join", false, "join letters")
	nostress := flag.Bool("nostress", false, "delete stress")
	nospaced := flag.Bool("nospaced", false, "delete spacing")
	matrices := flag.Bool("matrices", false, "show edit matrices")
	escapeunicode := flag.Bool("escapeunicode", false, "escape unicode when viewing")
	normalize := flag.String("normalize", "", "normalize unicode, for instance to NFC")
	flag.Parse()

	var lang *Language

	if *langFile != "" {
		var err error
		lang, err = LanguageNewFromFile(*langFile)
		if err != nil {
			return
		}

	}

	var dist float32

	var dict = make(map[string]struct{})
	if lang != nil {
		for k, v := range lang.Map {
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
	}

	var longSrcMulti int
	var longDstMulti int
	var longSrcMultiS int
	var longDstMultiS int

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

	loop(*srcFile, func(word1, word2 string) {

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

		var mat = levenshtein.MatrixTSlices[float32, string](srcslice([]rune(word1)), dstslice([]rune(word2)),
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

		var length = len(srcslice([]rune(word1))) + 1
		w1p := append(srcslice([]rune(word1)), "")
		w2p := append(dstslice([]rune(word2)), "")
		if d > 0 && matrices != nil && *matrices {
			if (same != nil && *same && len(w1p) == len(w2p)) || (same == nil) || (same != nil && !*same && len(w1p) != len(w2p)) {
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
			}
		}
		if d == 0 {
			tsvWriter.AddRow([]string{spacesep(srcslice([]rune(word1))), spacesep(dstslice([]rune(word2)))})
		}
		var final_from, final_to string
		if (same != nil && *same && len(w1p) == len(w2p)) || (same == nil) || (same != nil && !*same && len(w1p) != len(w2p)) {
			levenshtein.WalkVals(mat, uint(length), func(prev, this float32, x, y uint) bool {
				if _, ok := dict[w1p[x]+"\x00"+w2p[y]]; ok {
					return false
				}
				if hints != nil && *hints {

					if prev == this {
						return false
					}

					if escapeunicode != nil && *escapeunicode {
						final_from = ""
						for _, r := range w1p[x] {
							final_from += fmt.Sprintf("\\u%04X", r)
						}
					} else {
						final_from = w1p[x]
					}
					final_to = w2p[y]

				}
				return false
			})
		}

		if (join != nil) && (*join) {
			//println(word1, " ", word2)
			levenshtein.WalkVals(mat, uint(length), func(prev, this float32, x, y uint) bool {
				if prev != 0 && this == 0 && uint(len(w1p)) > x+1 && w2p[y] != "" {
					joined_from := w1p[x] + w1p[x+1]
					joined_to := w2p[y]
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
		if hints != nil && *hints && final_from != "" && final_to != "" {
			var found bool
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
		}
		dist += d
	})

	tsvWriter.Close()
	if (hints != nil && *hints) || (join != nil) && (*join) {
		data, err := json.Marshal(lang.Map)

		fmt.Println(string(data), err)
	}
	if loss != nil && *loss {
		fmt.Println("Edit distance is:", dist)
	}
}
