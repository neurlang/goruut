package main

import (
	"flag"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/neurlang/goruut/dicts"
	"github.com/neurlang/goruut/lib"
	"github.com/neurlang/goruut/models/requests"
)

func main() {
	langname := flag.String("langname", "", "directory language name")
	nostress := flag.Bool("nostress", false, "remove stress markers from phonemes")
	eval := flag.Bool("eval", true, "eval set")
	flag.Parse()

	if *langname == "" {
		fmt.Println("ERROR: langname flag is mandatory")
		return
	}
	srcfile := "../../dicts/" + *langname + "/multi.tsv"
	if eval != nil && *eval {
		srcfile = "../../dicts/" + *langname + "/multi_eval.tsv"
	}

	p := lib.NewPhonemizer(nil)
	lang := dicts.LangName(*langname)

	var total, correct atomic.Uint64
	var wrong_mutex sync.Mutex
	var wrong_sentence string
	var wrong_graphemes, wrong_phonemes, wrong_words []string

	loop(load(srcfile, 0), 1000, func(sentence, ipa string) {
		expectedPhonemes := strings.Split(ipa, " ")
		expectedGraphemes := strings.Split(sentence, " ")
		var words []string

		resp := p.Sentence(requests.PhonemizeSentence{
			Sentence:  sentence,
			Language:  lang,
			IsReverse: false,
		})

		if len(resp.Words) < len(expectedPhonemes) {
			fmt.Printf("Mismatched word count in line: %s (expected %d, got %d)\n",
				sentence, len(expectedPhonemes), len(resp.Words))
			return
		}
		if len(resp.Words) > len(expectedPhonemes) {
			for i := 0; i < len(resp.Words); i++ {
				w := resp.Words[i]
				if len(resp.Words) <= len(expectedPhonemes) {
					break
				}
				for i < len(resp.Words) && i < len(expectedGraphemes) && w.CleanWord != expectedGraphemes[i] {
					expectedGraphemes = slices.Insert(expectedGraphemes, i, "_")
					expectedPhonemes = slices.Insert(expectedPhonemes, i, "_")
					i++
				}
			}
		}
		for _, word := range resp.Words {
			words = append(words, word.Phonetic)
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

			total.Add(1)
			if generated == expected {
				correct.Add(1)
			} else {
				//fmt.Printf("Error: word '%s' expected '%s' got '%s'\n",
				//	word.CleanWord, expected, generated)

				wrong_mutex.Lock()

				wrong_sentence = sentence
				wrong_graphemes = expectedGraphemes
				wrong_phonemes = expectedPhonemes
				wrong_words = words

				wrong_mutex.Unlock()
			}
		}
	})

	if total.Load() > 0 {
		accuracy_1k_percent := 100000 * int64(correct.Load()) / int64(total.Load())
		fmt.Printf("WER Accuracy: %d.%03d%% (%d/%d)\n", accuracy_1k_percent/1000, accuracy_1k_percent%1000, correct.Load(), total.Load())
	} else {
		fmt.Println("No test cases processed")
	}
	fmt.Println("Last wrong sentence: ", wrong_sentence)
	fmt.Println("Last wrong graphemes: ", wrong_graphemes)
	fmt.Println("Last wrong phonemes: ", wrong_phonemes)
	fmt.Println("Last wrong words: ", wrong_words)
}

func removeStress(s string) string {
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "ˈ", "")
	s = strings.ReplaceAll(s, "ˌ", "")
	return s
}
