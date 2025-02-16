package main

import "strings"

import (
	"unicode"
	"unicode/utf8"
)

type SolutionEval struct {
	Map map[string]map[int]map[string]struct{}
	DstMultiPrefix map[string]struct{}
	DstMultiSuffix map[string]struct{}
	DropLast map[string]struct{}
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
	return &SolutionEval{
		Map: newMap,
		DstMultiPrefix: s.DstMultiPrefix,
		DstMultiSuffix: s.DstMultiSuffix,
		DropLast: s.DropLast,
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
		DstMultiPrefix: s.DstMultiPrefix,
		DstMultiSuffix: s.DstMultiSuffix,
		DropLast: s.DropLast,
	}
}

func (s *SolutionEval) With(src, dst string) *SolutionEval {
	newMap := make(map[string]map[int]map[string]struct{})
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
	return &SolutionEval{
		Map: newMap,
		DstMultiPrefix: s.DstMultiPrefix,
		DstMultiSuffix: s.DstMultiSuffix,
		DropLast: s.DropLast,
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
func (s *SolutionEval) IsDstMultiPrefix(str string) bool {
	_, ok := s.DstMultiPrefix[str]
	return ok
}
func (s *SolutionEval) IsDstMultiSuffix(str string) bool {
	_, ok := s.DstMultiSuffix[str]
	return ok
}

func (s *SolutionEval) Align(word1, word2 string, asymmetric bool) (ret *[2][]string) {
	if asymmetric {
		return s.AlignAsymmetric(word1, word2)
	}
	return s.AlignSymmetric(word1, word2)
}
func (s *SolutionEval) AlignSymmetric(word1, word2 string) (ret *[2][]string) {
	if len(word1) == 0 && len(word2) == 0 {
		return &[2][]string{}
	}
	for i := range word1 {
		word1k := word1[:len(word1)-i]
		if counts, ok := s.Map[strings.Trim(word1k, "_")]; ok {
			for j, vals := range counts {
				if j > len(word2) {
					continue
				}
				word2p := word2[:j]
				if _, ok := vals[word2p]; ok {
					ret = s.AlignSymmetric(word1[len(word1)-i:], word2[j:])
					if ret != nil {
						ret[0] = append([]string{word1k}, ret[0]...)
						ret[1] = append([]string{word2p}, ret[1]...)
						return
					}
				}
			}
		}
	}
	return
}

func (s *SolutionEval) AlignAsymmetric(word1, word2 string) *[2][]string {
	for i := range word1 {
		word1k := word1[:len(word1)-i]
		if counts, ok := s.Map[strings.Trim(word1k, "_")]; ok {
			for j := range word2 {
				word2p := word2[:len(word2)-j]
				if _, exists := counts[len(word2p)][word2p]; exists {
					// Allow partial alignment (not requiring full string processing)
					ret := &[2][]string{
						{word1k},
						{word2p},
					}

					// Try continuing alignment on the remaining substrings
					if next := s.AlignAsymmetric(word1[len(word1)-i:], word2[len(word2)-j:]); next != nil {
						ret[0] = append(ret[0], next[0]...)
						ret[1] = append(ret[1], next[1]...)
					}

					return ret // Return the first valid alignment found
				}
			}
		}
	}
	return nil // Allow partial matches; returning nil means no valid alignment found
}


// isCombiner checks if a rune is a UTF-8 combining character.
func isCombiner(r rune) bool {
	return unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || unicode.Is(unicode.Mc, r)
}

// stringStartsWithCombiner checks if the string starts with a UTF-8 combining character.
func stringStartsWithCombiner(s string) bool {
	if s == "" {
		return false
	}

	switch s {
	case "˧", "˥","˨", "˩", "˦", "ː":
		return true
	}

	r, _ := utf8.DecodeRuneInString(s)
	return isCombiner(r)
}
func (s *SolutionEval) AlignHybrid(word1, word2 string) *[2][]string {
    // Base case: if both words are empty, return nil
    if len(word1) == 0 && len(word2) == 0 {
        return nil
    }

    // Convert strings to rune slices for rune-safe manipulation
    runes1 := []rune(word1)
    runes2 := []rune(word2)

    // Try to find the longest prefix in word1 that exists in s.Map
    for i := len(runes1); i > 0; i-- {
        prefix1 := string(runes1[:i])
        if counts, ok := s.Map[strings.Trim(prefix1, "_")]; ok {
            // Try to find the corresponding prefix in word2
            for j := len(runes2); j > 0; j-- {
                prefix2 := string(runes2[:j])
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
        }
    }

    // If no valid prefix pair is found, fall back to single-rune processing
    // Ensure synchronization by processing both strings in a way that maintains alignment
    rune1 := firstRuneWithCombining(runes1)
    rune2 := firstRuneWithCombining(runes2)

    // If one string is empty, process the other string fully
    if len(rune1) == 0 {
        return &[2][]string{
            nil,
            splitIntoRunesWithCombining(runes2),
        }
    }
    if len(rune2) == 0 {
        return &[2][]string{
            splitIntoRunesWithCombining(runes1),
            nil,
        }
    }

    // Recursively align the remaining parts
    remaining := s.AlignHybrid(string(runes1[len(rune1):]), string(runes2[len(rune2):]))
    if remaining == nil {
        return &[2][]string{
            {string(rune1)},
            {string(rune2)},
        }
    }
    return &[2][]string{
        append([]string{string(rune1)}, remaining[0]...),
        append([]string{string(rune2)}, remaining[1]...),
    }
}

// Helper function to get the first rune (including combining characters) from a rune slice
func firstRuneWithCombining(runes []rune) []rune {
    if len(runes) == 0 {
        return nil
    }
    // Include combining characters that follow the first rune
    i := 1
    for i < len(runes) && isCombiningCharacter(runes[i]) {
        i++
    }
    return runes[:i]
}

// Helper function to check if a rune is a combining character
func isCombiningCharacter(r rune) bool {
    // Unicode combining characters are in the range 0x0300–0x036F
    return r >= 0x0300 && r <= 0x036F
}

// Helper function to split a rune slice into individual runes (including combining characters)
func splitIntoRunesWithCombining(runes []rune) []string {
    var result []string
    for i := 0; i < len(runes); {
        j := i + 1
        for j < len(runes) && isCombiningCharacter(runes[j]) {
            j++
        }
        result = append(result, string(runes[i:j]))
        i = j
    }
    return result
}
