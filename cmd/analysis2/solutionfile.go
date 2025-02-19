package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"bytes"
)

type SolutionFile struct {
	Map map[string][]string            `json:"Map"`
	SrcMulti       []string            `json:"SrcMulti"`
	DstMulti       []string            `json:"DstMulti"`
	SrcMultiSuffix []string            `json:"SrcMultiSuffix"`
	DstMultiSuffix []string            `json:"DstMultiSuffix"`
	DstMultiPrefix []string            `json:"DstMultiPrefix"`
	DropLast       []string            `json:"DropLast"`

	SplitBefore []string `json:"SplitBefore"`
	SplitAfter  []string `json:"SplitAfter"`

	PrePhonWordSteps interface{} `json:"PrePhonWordSteps"`

	UseCombining  bool `json:"UseCombining"`
}

func (s *SolutionFile) WithoutKey(key string) {
	//fmt.Println("WITHOUT KEY", key)
	delete(s.Map, key)
}
func (s *SolutionFile) WithoutValue(value string) {
	//fmt.Println("WITHOUT VALUE", value)
	for k, v := range s.Map {
		for i, val := range v {
			if val == value {
				v[i] = v[len(v)-1]
				v = v[:len(v)-1]
			}
		}
		s.Map[k] = v
	}
}
func (s *SolutionFile) With(src, dst string) {
	//fmt.Println("WITH", src, dst)
	if arr, ok := s.Map[src]; ok {
		for _, val := range arr {
			if val == dst {
				return
			}
		}
		s.Map[src] = append(s.Map[src], dst)
	} else {
		s.Map[src] = []string{dst}
	}
}
func (s *SolutionFile) LoadFromJson(file string) error {
	// Read the JSON file
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return err
	}

	// Parse the JSON data into the Language struct
	err = json.Unmarshal(data, s)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return err
	}
	return nil
}

func (s *SolutionFile) SaveToJson(file string) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	data = bytes.ReplaceAll(data, []byte(`],"`), []byte("],\n\""))

	err = ioutil.WriteFile(file, data, 0755)
	if err != nil {
		return err
	}
	return nil
}

func (s *SolutionFile) ToEval() (e *SolutionEval) {
	e = &SolutionEval{
		Map: make(map[string]map[int]map[string]struct{}),
		Drop: make(map[string]struct{}),
		DstMultiPrefix: make(map[string]struct{}),
		DstMultiSuffix: make(map[string]struct{}),
		DropLast: make(map[string]struct{}),
		UseCombining: s.UseCombining,
	}
	for k, val := range s.Map {
		if k == "" {
			continue
		}
		if len(val) == 0 {
			continue
		}
		if len(val) == 1 && val[0] == "" {
			e.Drop[k] = struct{}{}
			continue
		}
		e.Map[k] = make(map[int]map[string]struct{})
		for _, v := range val {
			if v == "" {
				e.Drop[k] = struct{}{}
				continue
			}
			if e.Map[k][len(v)] == nil {
				e.Map[k][len(v)] = make(map[string]struct{})
			}
			e.Map[k][len(v)][v] = struct{}{}
		}
	}
	for _, val := range s.DstMultiPrefix {
		e.DstMultiPrefix[val] = struct{}{}
	}
	for _, val := range s.DstMultiSuffix {
		e.DstMultiSuffix[val] = struct{}{}
	}
	for _, val := range s.DropLast {
		e.DropLast[val] = struct{}{}
	}
	return
}
