package main

import "github.com/neurlang/levenshtein"
import "github.com/neurlang/goruut/lib"
import "github.com/neurlang/goruut/dicts"
import "github.com/neurlang/goruut/models/requests"
import "os"
import "fmt"
import "github.com/neurlang/classifier/parallel"
import "bufio"
import "flag"
import "strings"
import "sync/atomic"
import "math/rand"

func loop(filename string, top, group int, do func(string, string)) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	var slice [][2]string

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

		slice = append(slice, [2]string{column1, column2})
	}

	// Check for any scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}

	rand.Shuffle(len(slice), func(i, j int) { slice[i], slice[j] = slice[j], slice[i] })
	if len(slice) > top {
		slice = slice[:top]
	}

	parallel.ForEach(len(slice), group, func(n int) {
		// Process each column
		column1 := slice[n][0]
		column2 := slice[n][1]
		// Example: Print the columns
		do(column1, column2)
	})

}

func main() {
	langname := flag.String("langname", "", "directory language name")
	isreverse := flag.Bool("reverse", false, "is reverse")
	nostress := flag.Bool("nostress", false, "no stress")
	flag.Parse()

	var coolname string
	var srcfile string

	if langname != nil {
		coolname = dicts.LangName(*langname)
		srcfile = "../../dicts/" + *langname + "/dirty.tsv"
	}

	p := lib.NewPhonemizer(nil)

	var percent, errsum, total atomic.Uint64
	loop(srcfile, 10000, 1000, func(word1, word2 string) {
		total.Add(1)
		if nostress != nil && *nostress {
			word2 = strings.ReplaceAll(word2, "'", "")
			word2 = strings.ReplaceAll(word2, "ˈ", "")
			word2 = strings.ReplaceAll(word2, "ˌ", "")
		}

		if isreverse != nil && *isreverse {
			word1, word2 = word2, word1
		}
		resp := p.Sentence(requests.PhonemizeSentence{
			Sentence: word1,
			Language: coolname,
			IsReverse: isreverse != nil && *isreverse,
		})

		var target string
		for i := range resp.Words {
			target += resp.Words[i].Phonetic + " "
		}
		target = strings.Trim(target, " ")
		if nostress != nil && *nostress {
			target = strings.ReplaceAll(target, "'", "")
			target = strings.ReplaceAll(target, "ˈ", "")
			target = strings.ReplaceAll(target, "ˌ", "")
		}

		var mat = levenshtein.Matrix[uint64](uint(len([]rune(target))), uint(len([]rune(word2))),
			nil, nil,
			levenshtein.OneSlice[rune, uint64]([]rune(target), []rune(word2)), nil)
		var dist = *levenshtein.Distance(mat)
		errsum.Add(dist)
		if target == word2 {
			percent.Add(1)
		}



		//success := 100 * percent.Load() / total.Load()
		//println("[success rate]", success, "%", "with", errsum.Load(), "errors", percent.Load(), "successes", "for", *langname)
	})
	{
		success := 100 * percent.Load() / total.Load()
		println("[success rate]", success, "%", "with", errsum.Load(), "errors", percent.Load(), "successes", "for", *langname)
	}
}
