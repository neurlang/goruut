package main

import "github.com/neurlang/levenshtein"

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
	loss := flag.Bool("loss", false, "show edit distance sum (loss, error)")
	same := flag.Bool("same", false, "show same matrices")
	nostress := flag.Bool("nostress", false, "delete stress")
	nospaced := flag.Bool("nospaced", false, "delete spacing")
	matrices := flag.Bool("matrices", false, "show edit matrices")
	escapeunicode := flag.Bool("escapeunicode", false, "escape unicode")
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

	srcslice := func(word []rune) (o []string) {
	outer:
		for i := 0; i < len(word); i++ {
			if lang != nil {
				for _, multi := range lang.SrcMulti {
					if strings.HasPrefix(string(word[i:]), multi) {
						o = append(o, multi)
						i += len([]rune(multi)) - 1
						if i >= len(word) {
							return
						}
						continue outer
					}
				}
				for _, multi := range lang.SrcMultiSuffix {
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
	dstslice := func(word []rune) (o []string) {
	outer:
		for i := 0; i < len(word); i++ {
			if lang != nil {
				for _, multi := range lang.DstMulti {
					if strings.HasPrefix(string(word[i:]), multi) {
						o = append(o, multi)
						i += len([]rune(multi)) - 1
						if i >= len(word) {
							return
						}
						continue outer
					}
				}
				for _, multi := range lang.DstMultiSuffix {
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

	var tsvWriter TSVWriter
	if dstFile != nil {
		tsvWriter.Open(*dstFile, nil)
	}

	loop(*srcFile, func(word1, word2 string) {

		if nostress != nil && *nostress {
			word2 = strings.ReplaceAll(word2, "ˈ", "")
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
				for _, rs := range w1p {
					for _, r := range rs {
						fmt.Printf("\\u%04X", r)
					}
					fmt.Print(" ")
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
		levenshtein.Walk(mat, uint(length), func(x, y uint) {
			if _, ok := dict[w1p[x]+"\x00"+w2p[y]]; ok {
				return
			}
			if hints != nil && *hints {
				if escapeunicode != nil && *escapeunicode {
					for _, r := range w1p[x] {
						fmt.Printf("\\u%04X", r)
					}
				} else {
					fmt.Print(w1p[x])
				}
				fmt.Println("", w2p[y])

			}
		})
		dist += d
	})

	tsvWriter.Close()
	if loss != nil && *loss {
		fmt.Println("Edit distance is:", dist)
	}
}
