package main

import (
	"encoding/json"
	"io/ioutil"
)

type Language struct {
	Map map[string][]string
}

func NewLanguage(filename string) (s *Language, err error) {
	// Read the JSON file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	s = &Language{}

	// Parse the JSON data into the Language struct
	err = json.Unmarshal(data, s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (l *Language) Transform(word string, solver func(problem, solution []string) bool) bool {
	for i := 0; i < len(word); i++ {
		var prefix = word[:len(word)-i]
		for _, choice := range l.Map[prefix] {
			if i == 0 && solver([]string{prefix}, []string{choice}) {
				return true
			}
			var suffix = word[len(word)-i:]
			if l.Transform(suffix, func(prob, sol []string) bool {
				prob = append([]string{prefix}, prob...)
				sol = append([]string{choice}, sol...)
				return solver(prob, sol)
			}) {
				return true
			}
		}
	}
	return false
}
