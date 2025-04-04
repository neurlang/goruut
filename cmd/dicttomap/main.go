// Package dicttomap converts algned dictionary files into language.json phonetic mappings.
package main

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)
import (
	"encoding/json"
	"flag"
	"io/ioutil"
)
import "sync"

func nosep(sli []string) (sep string) {
	for _, w := range sli {
		sep += w
	}
	return sep
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

type baseLanguage struct {
	SrcMulti         interface{} `json:"SrcMulti"`
	DstMulti         interface{} `json:"DstMulti"`
	SrcMultiSuffix   interface{} `json:"SrcMultiSuffix"`
	DstMultiSuffix   interface{} `json:"DstMultiSuffix"`
	DropLast         interface{} `json:"DropLast"`
	DstMultiPrefix   interface{} `json:"DstMultiPrefix"`
	PrePhonWordSteps interface{} `json:"PrePhonWordSteps"`
	SplitBefore      interface{} `json:"SplitBefore"`
	SplitAfter       interface{} `json:"SplitAfter"`
	SplitAt          interface{} `json:"SplitAt"`
	IsDuplex         interface{} `json:"IsDuplex"`
	IsSrcSurround    interface{} `json:"IsSrcSurround"`
	SrcDuplicate     interface{} `json:"SrcDuplicate"`
}

type Language struct {
	Map map[string][]string `json:"Map"`
	baseLanguage
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

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {

	var mut sync.Mutex

	langFile := flag.String("lang", "", "path to the JSON file containing language data")
	//save := flag.Bool("save", false, "write lang file at the end")
	srcFile := flag.String("srcfile", "", "path to input TSV file containing source and target words dictionary")
	dropFile := flag.String("dropfile", "", "path to input TSV file containing dropped mappings")
	scanLast := flag.Int("scanlast", 0, "scan N last mappings")
	nodel := flag.Bool("nodel", false, "no delete rule")
	writeback := flag.Bool("writeback", false, "write result back to srcfile")
	flag.Parse()

	var lang *Language

	if langFile != nil && *langFile != "" {
		var err error
		lang, err = LanguageNewFromFile(*langFile)
		if err != nil {
			return
		}

	}
	_ = lang

	var histogram = make(map[[2]string]int)

	var srcData = load(*srcFile, 99999999)

	var writer TSVWriter
	writer.Open(*srcFile, nil)

	loop(srcData, 100, func(word1, word2 string) {
		sword1 := strings.Split(word1, " ")
		sword2 := strings.Split(word2, " ")

		var osword [2]string
		var histogram_current = make(map[[2]string]int)

		if len(sword1) != len(sword2) {
			return
		}

		for i := 1; i < len(sword1); i++ {
			key := nosep(sword1[:len(sword1)-i])
			val := nosep(sword2[:len(sword2)-i])
			if _, ok := lang.Map[key]; ok {

				if nodel != nil && *nodel && val == "" {
					continue
				}
				histogram_current[[2]string{key, val}]++
				osword[0] += " " + key
				osword[1] += " " + val

				//println(key, val, nosep(sword1), nosep(sword2))
				sword1 = sword1[len(sword1)-i:]
				sword2 = sword2[len(sword2)-i:]
				i = 0
			}
		}
		if len(sword1) > 0 {
			mut.Lock()
			for k, v := range histogram_current {
				histogram[k] += v
			}
			mut.Unlock()
			for i := range sword1 {
				if nodel != nil && *nodel && sword2[i] == "" {
					continue
				}
				mut.Lock()
				histogram[[2]string{sword1[i], sword2[i]}]++
				mut.Unlock()
			}
			osword[0] += " " + spacesep(sword1)
			osword[1] += " " + spacesep(sword2)
		}

		if len(osword[0]) > 0 && len(osword[1]) > 0 {
			//println(word1, word2, osword[0][1:], osword[1][1:])
			osword[0] = osword[0][1:]
			osword[1] = osword[1][1:]
			writer.AddRow(osword[:])
		}
	})
	writer.Close()
	if dropFile != nil && *dropFile != "" {
		loop(load(*dropFile, 99999999), 100, func(word1, word2 string) {
			mut.Lock()
			delete(histogram, [2]string{word1, word2})
			mut.Unlock()
		})
	}

	if scanLast != nil {
		for j := 0; j < *scanLast; j++ {

			var lowsrc, lowdst string
			var low = (1 << 31) - 1

			for k, v := range histogram {
				if v < low {
					lowsrc = k[0]
					lowdst = k[1]
					low = v
				}
			}

			delete(histogram, [2]string{lowsrc, lowdst})

			println(lowsrc + " (mapped to) " + lowdst)
			println()

			loop(srcData, 100, func(word1, word2 string) {
				sword1 := strings.Split(word1, " ")
				sword2 := strings.Split(word2, " ")

				if len(sword1) != len(sword2) {
					return
				}

				for i := range sword1 {
					if sword1[i] == lowsrc && sword2[i] == lowdst {
						mut.Lock()
						println(word1 + "\t" + word2)
						mut.Unlock()
						break
					}
				}
			})

			println()

		}
	}

	var data = make(map[string][]string)
	for k := range histogram {
		data[k[0]] = append(data[k[0]], k[1])
	}
	for k, sli := range data {
		sort.Slice(sli, func(i, j int) bool {
			return histogram[[2]string{k, sli[i]}] > histogram[[2]string{k, sli[j]}]
		})
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		println(err.Error())
		return
	}

	if writeback != nil && *writeback {
		var olang = lang
		olang.Map = data

		bytes, err := json.Marshal(olang)
		if err != nil {
			println(err.Error())
			return
		}
		str := strings.ReplaceAll(string(bytes), "],", "],\n")
		err = ioutil.WriteFile(*langFile, []byte(str), 0755)
		if err != nil {
			println(err.Error())
			return
		}
	} else {
		str := strings.ReplaceAll(string(bytes), "],", "],\n")
		fmt.Println(str)
	}
}
