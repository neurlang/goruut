package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
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

	loop(*srcFile, 100, func(word1, word2 string) {
		sword1 := strings.Split(word1, " ")
		sword2 := strings.Split(word2, " ")

		if len(sword1) != len(sword2) {
			return
		}

		for i := range sword1 {
			if nodel != nil && *nodel && sword2[i] == "" {
				continue
			}
			mut.Lock()
			histogram[[2]string{sword1[i], sword2[i]}]++
			mut.Unlock()
		}
	})

	if dropFile != nil && *dropFile != "" {
		loop(*dropFile, 100, func(word1, word2 string) {
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

			loop(*srcFile, 100, func(word1, word2 string) {
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
	str := strings.ReplaceAll(string(bytes), "],", "],\n")
	fmt.Println(str)
}
