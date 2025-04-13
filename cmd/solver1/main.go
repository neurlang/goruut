package main

import (
	"flag"
	"fmt"
)
import "strings"

func join(strs []string) string {
	return strings.Join(strs, "")
}

func main() {
	langname := flag.String("langname", "", "directory language name")
	wordlist := flag.String("wordlist", "", "wordlist.txt word list to solve word pronounces")

	flag.Parse()

	if *langname == "" {
		fmt.Println("ERROR: langname flag is mandatory")
		return
	}
	if *wordlist == "" {
		fmt.Println("ERROR: wordlist flag is mandatory")
		return
	}
	forward_file := "../../dicts/" + *langname + "/language.json"
	reverse_file := "../../dicts/" + *langname + "/language_reverse.json"

	var clean_forward, clean_reverse TSVWriter
	(&clean_forward).Open("clean.tsv", nil)
	(&clean_reverse).Open("clean_reverse.tsv", nil)

	var l1, err1 = NewLanguage(forward_file)
	if err1 != nil {
		panic(err1.Error())
	}
	var l2, err2 = NewLanguage(reverse_file)
	if err2 != nil {
		panic(err2.Error())
	}

	loop(load(*wordlist, 0), 1000, func(word string) {
		word = strings.Trim(word)
		if word == "" {
			return
		}
		var sols byte
		var done_prob, done_sol, done_phon, done_wrd []string
		// phonemize
		if !l1.Transform(word, func(wrd, phon []string) bool {
			// dephonemize
			return l2.Transform(join(phon), func(prob, sol []string) bool {
				// detect solution
				if join(sol) == word {
					if sols <= byte(len(word))&1 {
						done_prob = prob
						done_sol = sol
						done_phon = phon
						done_wrd = wrd
					}
					sols++
					return sols >= 3
				}
				return false
			})
		}) {
			if sols >= 1 && sols <= 2 {
				clean_forward.AddRow([]string{strings.Join(done_wrd, " "), strings.Join(done_phon, " ")})
				clean_reverse.AddRow([]string{strings.Join(done_prob, " "), strings.Join(done_sol, " ")})
			}
		}
	})
}
