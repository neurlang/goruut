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

	langMap := langStruct.Map

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
		// Split word using longest prefix rule
		segments := splitWordLongestPrefix(word, langMap)
		
		// Generate IPA for each segment and find best combination
		inferredIPA := generateInferredIPA(segments, langMap, ipa)
		
		// Calculate edit distance
		editDist := calculateEditDistance(inferredIPA, ipa)

		if editDist == 0 {
			return
		}
		
		// Print results
		result := fmt.Sprintf("%d\t%s\t%s\t%s", editDist, word, inferredIPA, ipa)
		//fmt.Println(result)
		
		if dumpWriter != nil {
			fmt.Fprintln(dumpWriter, result)
		}
	})
}

// splitWordLongestPrefix splits a word using longest prefix matching
func splitWordLongestPrefix(word string, langMap map[string][]string) []string {
	var segments []string
	i := 0
	
	for i < len(word) {
		longestMatch := ""
		
		// Find the longest prefix match
		for j := i + 1; j <= len(word); j++ {
			prefix := word[i:j]
			if _, exists := langMap[prefix]; exists {
				longestMatch = prefix
			}
		}
		
		if longestMatch != "" {
			segments = append(segments, longestMatch)
			i += len(longestMatch)
		} else {
			// If no match found, take single character
			segments = append(segments, string(word[i]))
			i++
		}
	}
	
	return segments
}

// generateInferredIPA generates IPA from segments using greedy approach
// For each segment, pick the phoneme option that gives lowest edit distance so far
func generateInferredIPA(segments []string, langMap map[string][]string, targetIPA string) string {
	if len(segments) == 0 {
		return ""
	}
	
	targetRunes := []rune(strings.ToLower(targetIPA))
	currentIPA := ""
	
	// Process each segment greedily
	for _, segment := range segments {
		var options []string
		
		if phonemes, exists := langMap[segment]; exists && len(phonemes) > 0 {
			options = phonemes
		} else {
			options = []string{segment}
		}
		
		// Find the best option for this segment
		bestOption := options[0]
		minDistance := uint64(^uint64(0)) // Max uint64
		
		for _, option := range options {
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
	
	return currentIPA
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