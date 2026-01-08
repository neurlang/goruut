package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/neurlang/goruut/dicts"
	"github.com/neurlang/levenshtein"
	di "github.com/martinarisk/di/dependency_injection"
	"github.com/neurlang/goruut/repo/interfaces"
)

type dummy struct{}

func (dummy) GetIpaFlavors() map[string]map[string]string {
	return make(map[string]map[string]string)
}
func (dummy) GetPolicyMaxWords() int {
	return 99999999999
}

func main() {
	langFile := flag.String("lang", "", "path to the language directory (e.g., 'english')")
	dumpFile := flag.String("dump", "", "file where to dump outliers")
	flag.Parse()

	if *langFile == "" {
		fmt.Println("Error: -lang parameter is required")
		os.Exit(1)
	}

	// Get the language name from directory
	langName := dicts.LangName(*langFile)
	if langName == "" {
		fmt.Printf("Error: unsupported language directory: %s\n", *langFile)
		os.Exit(1)
	}

	// Create dependency injection container
	dictGetter := dicts.DictGetter{}
	di := di.NewDependencyInjection()
	di.Add((interfaces.DictGetter)(dictGetter))
	di.Add((interfaces.IpaFlavor)(dummy{}))
	di.Add((interfaces.PolicyMaxWords)(dummy{}))

	// Load language.json
	langData, err := dictGetter.GetDict(langName, "language.json")
	if err != nil {
		fmt.Printf("Error loading language.json: %v\n", err)
		os.Exit(1)
	}

	var langStruct struct {
		Map map[string][]string `json:"Map"`
	}
	if err := json.Unmarshal(langData, &langStruct); err != nil {
		fmt.Printf("Error parsing language.json: %v\n", err)
		os.Exit(1)
	}

	// Load language_reverse.json
	langReverseData, err := dictGetter.GetDict(langName, "language_reverse.json")
	if err != nil {
		fmt.Printf("Error loading language_reverse.json: %v\n", err)
		os.Exit(1)
	}

	var langReverseStruct struct {
		Map map[string][]string `json:"Map"`
	}
	if err := json.Unmarshal(langReverseData, &langReverseStruct); err != nil {
		fmt.Printf("Error parsing language_reverse.json: %v\n", err)
		os.Exit(1)
	}

	langMap := langStruct.Map
	langReverseMap := langReverseStruct.Map

	// Load lexicon.tsv
	lexiconPath := filepath.Join("dicts", *langFile, "lexicon.tsv")
	lexiconData := load(lexiconPath, -1) // Load all entries

	if len(lexiconData) == 0 {
		fmt.Printf("Error: no data loaded from %s\n", lexiconPath)
		os.Exit(1)
	}

	var dumpWriter *os.File
	if *dumpFile != "" {
		dumpWriter, err = os.Create(*dumpFile)
		if err != nil {
			fmt.Printf("Error creating dump file: %v\n", err)
			os.Exit(1)
		}
		defer dumpWriter.Close()
	}

	// Process each word-IPA pair
	loop(lexiconData, 1000, func(word, ipa string) {
		// Forward direction: word -> IPA
		segments := splitWordLongestPrefix(word, langMap)
		inferredIPA := generateInferredIPA(segments, langMap, ipa)
		editDistForward := calculateEditDistance(inferredIPA, ipa)

		// Reverse direction: IPA -> word
		ipaSegments := splitWordLongestPrefix(ipa, langReverseMap)
		inferredWord := generateInferredWord(ipaSegments, langReverseMap, word)
		editDistReverse := calculateEditDistance(inferredWord, word)

		// Only print if there are errors in either direction
		if editDistForward > 0 || editDistReverse > 0 {
			result := fmt.Sprintf("%d\t%s\t%s\t%s\t%s", 
				editDistForward + editDistReverse, word, ipa, inferredWord, inferredIPA)
			fmt.Println(result)
			
			if dumpWriter != nil {
				fmt.Fprintln(dumpWriter, result)
			}
		}
	})
}

// splitWordLongestPrefix splits a word using longest prefix matching based on runes
func splitWordLongestPrefix(word string, langMap map[string][]string) []string {
	var segments []string
	runes := []rune(word)
	i := 0
	
	for i < len(runes) {
		longestMatch := ""
		longestMatchLen := 0
		
		// Find the longest prefix match by testing all possible lengths
		for j := i + 1; j <= len(runes); j++ {
			prefix := string(runes[i:j])
			if _, exists := langMap[prefix]; exists {
				longestMatch = prefix
				longestMatchLen = j - i
			}
		}
		
		if longestMatch != "" {
			segments = append(segments, longestMatch)
			i += longestMatchLen
		} else {
			// If no match found, take single rune
			segments = append(segments, string(runes[i]))
			i++
		}
	}
	
	return segments
}

// generateInferredIPA generates IPA from segments using greedy approach
// For each segment, pick the phoneme option that gives lowest edit distance so far
// Skip segments that have no mapping in langMap
func generateInferredIPA(segments []string, langMap map[string][]string, targetIPA string) string {
	if len(segments) == 0 {
		return ""
	}
	
	targetRunes := []rune(strings.ToLower(targetIPA))
	currentIPA := ""
	
	// Process each segment greedily
	for _, segment := range segments {
		// Skip segments that have no mapping
		if phonemes, exists := langMap[segment]; exists && len(phonemes) > 0 {
			// Find the best option for this segment
			bestOption := phonemes[0]
			minDistance := uint64(^uint64(0)) // Max uint64
			
			for _, option := range phonemes {
				// Try this option and calculate edit distance
				testIPA := currentIPA + option
				testRunes := []rune(strings.ToLower(testIPA))
				
				// Calculate edit distance between partial IPA and corresponding part of target
				targetLen := len(targetRunes)
				testLen := len(testRunes)
				compareLen := testLen
				if compareLen > targetLen {
					compareLen = targetLen
				}
				
				var targetPart []rune
				if compareLen > 0 {
					targetPart = targetRunes[:compareLen]
				}
				
				mat := levenshtein.Matrix[uint64](uint(len(testRunes)), uint(len(targetPart)),
					nil, nil,
					levenshtein.OneSlice[rune, uint64](testRunes, targetPart), nil)
				
				distance := *levenshtein.Distance(mat)
				
				if distance < minDistance {
					minDistance = distance
					bestOption = option
				}
			}
			
			// Add the best option to current IPA
			currentIPA += bestOption
		}
		// If no mapping exists, skip this segment (don't add anything)
	}
	
	return currentIPA
}

// generateInferredWord generates word from IPA segments using greedy approach
// For each IPA segment, pick the grapheme option that gives lowest edit distance so far
// Skip segments that have no mapping in langReverseMap
func generateInferredWord(ipaSegments []string, langReverseMap map[string][]string, targetWord string) string {
	if len(ipaSegments) == 0 {
		return ""
	}
	
	targetRunes := []rune(strings.ToLower(targetWord))
	currentWord := ""
	
	// Process each IPA segment greedily
	for _, ipaSegment := range ipaSegments {
		// Skip segments that have no mapping
		if graphemes, exists := langReverseMap[ipaSegment]; exists && len(graphemes) > 0 {
			// Find the best option for this IPA segment
			bestOption := graphemes[0]
			minDistance := uint64(^uint64(0)) // Max uint64
			
			for _, option := range graphemes {
				// Try this option and calculate edit distance
				testWord := currentWord + option
				testRunes := []rune(strings.ToLower(testWord))
				
				// Calculate edit distance between partial word and corresponding part of target
				targetLen := len(targetRunes)
				testLen := len(testRunes)
				compareLen := testLen
				if compareLen > targetLen {
					compareLen = targetLen
				}
				
				var targetPart []rune
				if compareLen > 0 {
					targetPart = targetRunes[:compareLen]
				}
				
				mat := levenshtein.Matrix[uint64](uint(len(testRunes)), uint(len(targetPart)),
					nil, nil,
					levenshtein.OneSlice[rune, uint64](testRunes, targetPart), nil)
				
				distance := *levenshtein.Distance(mat)
				
				if distance < minDistance {
					minDistance = distance
					bestOption = option
				}
			}
			
			// Add the best option to current word
			currentWord += bestOption
		}
		// If no mapping exists, skip this segment (don't add anything)
	}
	
	return currentWord
}

// calculateEditDistance calculates Levenshtein distance between two strings
func calculateEditDistance(s1, s2 string) uint64 {
	runes1 := []rune(strings.ToLower(s1))
	runes2 := []rune(strings.ToLower(s2))
	
	mat := levenshtein.Matrix[uint64](uint(len(runes1)), uint(len(runes2)),
		nil, nil,
		levenshtein.OneSlice[rune, uint64](runes1, runes2), nil)
	
	return *levenshtein.Distance(mat)
}