package main

import "strings"
import "sort"
import (
	"unicode"
	"unicode/utf8"
)

type SolutionEval struct {
	Map map[string]map[int]map[string]struct{}
	Drop map[string]struct{}
	DstMultiPrefix map[string]struct{}
	DstMultiSuffix map[string]struct{}
	DropLast map[string]struct{}
	UseCombining bool
}

func (s *SolutionEval) GetValues() (ret map[string]struct{}) {
	ret = make(map[string]struct{})
	for _, v := range s.Map {
		for _, v2 := range v {
			for k3 := range v2 {
				ret[k3] = struct{}{}
			}
		}
	}
	return
}
func (s *SolutionEval) WithoutValue(value string) *SolutionEval {
	newMap := make(map[string]map[int]map[string]struct{})
	newDrop := make(map[string]struct{})
	for k, v := range s.Map {
		if _, ok := v[len(value)][value]; !ok {
			newMap[k] = v
			continue
		}
		newMap[k] = make(map[int]map[string]struct{})
		for k2, v2 := range v {
			if k2 == len(value) {
				newMap[k][k2] = make(map[string]struct{})
				for k3 := range v2 {
					if k3 != value {
						newMap[k][k2][k3] = struct{}{}
					}
				}
				// Only keep the map if it has elements
				if len(newMap[k][k2]) == 0 {
					delete(newMap[k], k2)  // Avoid keeping an empty map
				}
			} else {
				newMap[k][k2] = v2
			}
		}
	}
	if value != "" {
		newDrop = s.Drop
	}
	return &SolutionEval{
		Drop: newDrop,
		Map: newMap,
		DstMultiPrefix: s.DstMultiPrefix,
		DstMultiSuffix: s.DstMultiSuffix,
		DropLast: s.DropLast,
		UseCombining: s.UseCombining,
	}
}

func (s *SolutionEval) WithoutKey(key string) *SolutionEval {
	newMap := make(map[string]map[int]map[string]struct{})
	for k, v := range s.Map {
		if k != key {
			newMap[k] = v
		}
	}
	return &SolutionEval{
		Map: newMap,
		Drop: s.Drop,
		DstMultiPrefix: s.DstMultiPrefix,
		DstMultiSuffix: s.DstMultiSuffix,
		DropLast: s.DropLast,
		UseCombining: s.UseCombining,
	}
}

func (s *SolutionEval) With(src, dst string) *SolutionEval {
	newMap := make(map[string]map[int]map[string]struct{})
	newDrop := make(map[string]struct{})
	if dst == "" {
		newMap = s.Map
		for k := range s.Drop {
			newDrop[k] = struct{}{}
		}
		newDrop[src] = struct{}{}
	} else {
		newDrop = s.Drop
		for k, v := range s.Map {
			if k != src {
				newMap[k] = v
			} else {
				newMap[k] = make(map[int]map[string]struct{})
				for k2, v2 := range v {
					if k2 != len(dst) {
						newMap[k][k2] = v2
					} else {
						newMap[k][k2] = make(map[string]struct{})
						for k3 := range v2 {
							newMap[k][k2][k3] = struct{}{}
						}
					}
				}
			}
		}
		if newMap[src] == nil {
			newMap[src] = make(map[int]map[string]struct{})
		}
		if newMap[src][len(dst)] == nil {
			newMap[src][len(dst)] = make(map[string]struct{})
		}
		newMap[src][len(dst)][dst] = struct{}{}
	}
	return &SolutionEval{
		Drop: newDrop,
		Map: newMap,
		DstMultiPrefix: s.DstMultiPrefix,
		DstMultiSuffix: s.DstMultiSuffix,
		DropLast: s.DropLast,
		UseCombining: s.UseCombining,
	}
}

func (s *SolutionEval) IsEdge(src, dst string) bool {
	data, ok := s.Map[src]
	if !ok {
		return false
	}
	_, ok2 := data[len(dst)][dst]
	return ok2
}
func (s *SolutionEval) IsDropLast(str string) bool {
	_, ok := s.DropLast[str]
	return ok
}
func (s *SolutionEval) IsDrop(str string) bool {
	_, ok := s.Drop[str]
	return ok
}
func (s *SolutionEval) IsDstMultiPrefix(str string) bool {
	_, ok := s.DstMultiPrefix[str]
	if ok {
		//println("prefix", str)
		return true
	}
	for k := range s.DstMultiPrefix {
		if strings.HasSuffix(str, k) {
			//println("prefix", str, k)
			return true
		}
	}
	return ok
}
func (s *SolutionEval) IsDstMultiSuffix(str string) bool {
	_, ok := s.DstMultiSuffix[str]
	if ok {
		//println("suffix", str)
		return true
	}
	for k := range s.DstMultiSuffix {
		if strings.HasPrefix(str, k) {
			//println("suffix", str, k)
			return true
		}
	}
	return ok
}
func (s *SolutionEval) ComplexityLoss(aligned1 []string) (ret uint64) {
	for _, v := range aligned1 {
		ret += uint64(len(s.Map[v]))
	}
	if ret > uint64(len(aligned1)) {
		ret -= uint64(len(aligned1))
	} else {
		return 0
	}
	return
}
func (s *SolutionEval) Align(word1, word2 string, asymmetric, isCleaning bool) (ret *[2][]string, cplxLoss uint64) {
	if asymmetric {
		ret = s.AlignAsymmetric(word1, word2, isCleaning)
	} else {
		ret = s.AlignSymmetric(word1, word2, isCleaning)
	}
	if ret != nil {
		cplxLoss = s.ComplexityLoss(ret[0])
	}
	return
}
func (s *SolutionEval) AlignSymmetric(word1, word2 string, isCleaning bool) (ret *[2][]string) {
	if len(word1) == 0 && len(word2) == 0 {
		return &[2][]string{}
	}
	for i := range word1 {
		word1k := word1[:len(word1)-i]
		key := strings.Trim(word1k, "_")
		if counts, ok := s.Map[key]; ok {
			// Extract and sort lengths in descending order
			lengths := make([]int, 0, len(counts))
			for l := range counts {
				lengths = append(lengths, l)
			}
			if s.IsDrop(key) {
				lengths = append(lengths, 0)
			}
			sort.Sort(sort.Reverse(sort.IntSlice(lengths))) // Sort from big to small

			// Iterate over sorted lengths
			for _, j := range lengths {
				if j > len(word2) {
					continue
				}
				word2p := word2[:j]
				if _, ok := counts[j][word2p]; ok {
					ret = s.AlignSymmetric(word1[len(word1)-i:], word2[j:], isCleaning)
					if ret != nil {
						ret[0] = append([]string{word1k}, ret[0]...)
						ret[1] = append([]string{word2p}, ret[1]...)
						return
					}
				}
			}
			if isCleaning {
				break
			}
		}
	}
	return
}

func (s *SolutionEval) AlignAsymmetric(word1, word2 string, isCleaning bool) (ret *[2][]string) {
	for i := range word1 {
		word1k := word1[:len(word1)-i]
		key := strings.Trim(word1k, "_")
		if counts, ok := s.Map[key]; ok {
			// Extract and sort lengths in descending order
			lengths := make([]int, 0, len(counts))
			for l := range counts {
				lengths = append(lengths, l)
			}
			if s.IsDrop(key) {
				lengths = append(lengths, 0)
			}
			sort.Sort(sort.Reverse(sort.IntSlice(lengths))) // Sort from big to small

			// Iterate over sorted lengths
			for _, j := range lengths {
				if j > len(word2) {
					continue
				}
				word2p := word2[:j]
				if _, exists := counts[j][word2p]; exists {
					retok := &[2][]string{
						{word1k},
						{word2p},
					}
					// Allow partial alignment (not requiring full string processing)
					if i == 0 || len(word2) == j {
						return retok // return end match
					}
					// Try continuing alignment on the remaining substrings
					if next := s.AlignAsymmetric(word1[len(word1)-i:], word2[j:], isCleaning); next != nil {
						
						retok[0] = append(retok[0], next[0]...)
						retok[1] = append(retok[1], next[1]...)
						if ret == nil {
							ret = retok
						} else if len(ret[0]) < len(retok[0]) {
							ret = retok
						}
					}
					if ret == nil {
						ret = retok
					}
				}
			}
			if isCleaning {
				break
			}
		}
	}
	return ret // Allow partial matches; returning nil means no valid alignment found
}


// isCombiner checks if a rune is a UTF-8 combining character.
func isCombiner(r rune) bool {
	return unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || unicode.Is(unicode.Mc, r)
}

// StringStartsWithCombiner checks if the string starts with a UTF-8 combining character.
func (s *SolutionEval) StringStartsWithCombiner(str string) bool {
	if str == "" {
		return false
	}
	if s.UseCombining {
		return false
	}

	r, _ := utf8.DecodeRuneInString(str)


	switch r {
	case '˧', '˥','˨', '˩', '˦', 'ː':
		return true
	}

	return isCombiner(r)
}
func (s *SolutionEval) AlignHybrid(word1, word2 string) *[2][]string {
    // Base case: if some words are empty, return nil
    if len(word1) == 0 && len(word2) == 0 {
        return nil
    }

    // Convert strings to rune slices for rune-safe manipulation
    runes1 := []rune(word1)
    runes2 := []rune(word2)
    var possibleprefix string

    if len(word2) != 0 || len(word1) != 0 {
	    // Try to find the longest prefix in word1 that exists in s.Map
	    for i := len(runes1); i > 0; i-- {
		prefix1 := string(runes1[:i])
		key1 := strings.Trim(prefix1, "_")
		if counts, ok := s.Map[key1]; ok {
		    // Try to find the corresponding prefix in word2
		    for j := 1; j <= len(runes2); j++ {
		        prefix2 := string(runes2[:j])
			if (s.IsDstMultiSuffix(prefix2) || s.IsDstMultiPrefix(prefix2)) {
				continue
			}
		        if _, exists := counts[len(prefix2)][prefix2]; exists {
		            // Found a valid pair, recursively align the remaining parts
		            remaining := s.AlignHybrid(string(runes1[i:]), string(runes2[j:]))
		            if remaining == nil {
		                return &[2][]string{
		                    {prefix1},
		                    {prefix2},
		                }
		            }
		            return &[2][]string{
		                append([]string{prefix1}, remaining[0]...),
		                append([]string{prefix2}, remaining[1]...),
		            }
		        }
		    }
		    // cjt mode
		    possibleprefix = prefix1
   		   // println(prefix1, word1, word2)
		}
	    }
    }



    
    // If no valid prefix pair is found, fall back to single-rune processing
    // Ensure synchronization by processing both strings in a way that maintains alignment
    rune1 := s.firstRuneWithCombining(runes1, true)
    rune2 := s.firstRuneWithCombining(runes2, false)

    // If one string is empty, process the other string fully
    if len(rune1) == 0 {
        return &[2][]string{
            nil,
            s.splitIntoRunesWithCombining(runes2, false),
        }
    }
    if len(rune2) == 0 {
        return &[2][]string{
            s.splitIntoRunesWithCombining(runes1, true),
            nil,
        }
    }
    
    // Recursively align the remaining parts
    remaining := s.AlignHybrid(string(runes1[len(rune1):]), string(runes2[len(rune2):]))
    if remaining == nil {

        remaining = s.AlignHybrid(word1[len(possibleprefix):], string(runes2[len(rune2):]))
        if remaining == nil {
		return &[2][]string{
		    {string(rune1)},
		    {string(rune2)},
		}
        }
        return &[2][]string{
		append([]string{possibleprefix}, remaining[0]...),
		append([]string{string(rune2)}, remaining[1]...),
	}
    }
    return &[2][]string{
        append([]string{string(rune1)}, remaining[0]...),
        append([]string{string(rune2)}, remaining[1]...),
    }
}

// Helper function to get the first rune (including combining characters) from a rune slice
func (s *SolutionEval) firstRuneWithCombining(runes []rune, is_src bool) []rune {
    if len(runes) == 0 {
        return nil
    }
    // Include combining characters that follow the first rune
    i := 1
    for i < len(runes) && (s.StringStartsWithCombiner(string(runes[i])) || (!is_src && (s.IsDstMultiSuffix(string(runes[i])) || s.IsDstMultiPrefix(string(runes[i]))))) {
        i++
    }
    return runes[:i]
}


// Helper function to split a rune slice into individual runes (including combining characters)
func (s *SolutionEval) splitIntoRunesWithCombining(runes []rune, is_src bool) []string {
    var result []string
    for i := 0; i < len(runes); {
        j := i + 1
   	for j < len(runes) && (s.StringStartsWithCombiner(string(runes[i])) || (!is_src && (s.IsDstMultiSuffix(string(runes[j])) || s.IsDstMultiPrefix(string(runes[i:j]))))) {
	    j++
	}
        result = append(result, string(runes[i:j]))
	//if !is_src {
		//result = append(result, "")
	//}
        i = j
    }
    return result
}
