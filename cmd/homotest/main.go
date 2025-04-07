package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/neurlang/goruut/dicts"
	"github.com/neurlang/goruut/lib"
	"github.com/neurlang/goruut/models/requests"
)

func main() {
	langname := flag.String("langname", "", "directory language name")
	inputFile := flag.String("input", "multi.tsv", "input TSV file")
	nostress := flag.Bool("nostress", false, "remove stress markers from phonemes")
	flag.Parse()

	if *langname == "" {
		fmt.Println("ERROR: langname flag is mandatory")
		return
	}

	file, err := os.Open(*inputFile)
	if err != nil {
		fmt.Println("Error opening input file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	p := lib.NewPhonemizer(nil)
	lang := dicts.LangName(*langname)

	var total, correct int
	var wrong_sentence string

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			fmt.Printf("Skipping invalid line: %s\n", line)
			continue
		}

		sentence := parts[0]
		expectedPhonemes := strings.Split(parts[1], " ")

		resp := p.Sentence(requests.PhonemizeSentence{
			Sentence:  sentence,
			Language:  lang,
			IsReverse: false,
		})

		if len(resp.Words) != len(expectedPhonemes) {
			fmt.Printf("Mismatched word count in line: %s (expected %d, got %d)\n",
				sentence, len(expectedPhonemes), len(resp.Words))
			continue
		}

		for i, word := range resp.Words {
			expected := expectedPhonemes[i]
			if expected == "_" {
				continue
			}

			generated := word.Phonetic
			if *nostress {
				generated = removeStress(generated)
				expected = removeStress(expected)
			}

			total++
			if generated == expected {
				correct++
			} else {
				fmt.Printf("Error: word '%s' expected '%s' got '%s'\n",
					word.CleanWord, expected, generated)
				wrong_sentence = sentence
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input file:", err)
	}

	if total > 0 {
		accuracy_1k_percent := 100000 * int64(correct) / int64(total)
		fmt.Printf("WER Accuracy: %d.%03d%% (%d/%d)\n", accuracy_1k_percent/1000, accuracy_1k_percent%1000, correct, total)
	} else {
		fmt.Println("No test cases processed")
	}
	fmt.Println("Last wrong sentence: ", wrong_sentence)
}

func removeStress(s string) string {
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "ˈ", "")
	s = strings.ReplaceAll(s, "ˌ", "")
	return s
}
