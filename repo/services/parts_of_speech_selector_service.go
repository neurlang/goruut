package services

import (
	"github.com/neurlang/classifier/hash"
	"github.com/neurlang/goruut/helpers"
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/repo"
	"strings"
)

import . "github.com/martinarisk/di/dependency_injection"

type IPartsOfSpeechSelectorService interface {
	Select(isReverse bool, lang string, sentence []map[string]uint32, languages []string) (ret [][3]string)
}

type PartsOfSpeechSelectorService struct {
	repo   *repo.IDictPhonemizerRepository
	repoa  *repo.IAutoTaggerRepository
	repoai *repo.IHashtronHomonymSelectorRepository
}

func next_continue_english(lang string, isReverse bool, my_tags, next_tags map[string]bool) bool {
	if isReverse {
		return false
	}
	if !strings.HasPrefix(lang, "English") {
		return false
	}
	if my_tags["thi"] && next_tags["consonant1st"] {
		return true
	}
	if my_tags["the"] && next_tags["vowel1st"] {
		return true
	}
	return false
}

func (p *PartsOfSpeechSelectorService) Select(isReverse bool, lang string, sentence []map[string]uint32, languages []string) (ret [][3]string) {

	var input []map[string][2]uint32

	for _, words := range sentence {
		var orig string
		for word, k := range words {
			if k == 0 {
				orig = strings.TrimRight(word, " ")
				break
			}
		}
		var inputmap = make(map[string][2]uint32)
		inputmap[orig+" "] = [2]uint32{0, 0}
		for word, k := range words {
			if k == 0 {
				continue
			}
			var tags = (*p.repo).LookupTags(isReverse, lang, orig, word)
			var hasdict bool
			for _, tag := range log.Error1(helpers.ParseJson[[]string]([]byte(tags))) {
				if tag == "dict" {
					hasdict = true
				}
			}
			log.Now().Debugf("PreSelect Orig: %s, Word: %s, WordsTags: %s, HasDict: %v", orig, word, tags, hasdict)
			if hasdict {
				inputmap[word] = [2]uint32{k, 0}
			} else {
				// set dict flag
				//inputmap[word] = [2]uint32{k ^ hash.StringHash(0, "dict"), 1}
			}
		}
		log.Now().Debugf("PreSelect Word: %s Now: %v", orig, inputmap)
		input = append(input, inputmap)
	}

	var preferred = (*p.repoai).Select(isReverse, lang, input)

	log.Now().Debugf("Preferred: %v", preferred)

	var intermediate []map[[2]string][]string
	for i, words := range sentence {
		var last_preferred, hash_preferred uint32
		for _, row := range preferred {
			if row[0] != uint32(i) {
				continue
			}
			if row[3] == 0 {
				continue
			}
			last_preferred = row[2]
			hash_preferred = row[1]
			break
		}
		log.Now().Debugf("Preferred: %d %d", last_preferred, hash_preferred)
		var orig string
		for word, k := range words {
			if k == 0 {
				orig = strings.TrimRight(word, " ")
				break
			}
		}
		var wordstags = make(map[[2]string][]string)
		for word, k := range words {
			if k == 0 {
				continue
			}
			var tags = (*p.repo).LookupTags(isReverse, lang, orig, word)
			log.Now().Debugf("Orig: %s, Word: %s, WordsTags: %s", orig, word, tags)
			var tags_parsed = (*p.repoa).TagWord(isReverse, lang, orig, word)
			if tags != "" && tags != "[]" {
				tags_parsed = append(tags_parsed, log.Error1(helpers.ParseJson[[]string]([]byte(tags)))...)
			}
			for _, lang := range languages {
				var tags = (*p.repo).LookupTags(isReverse, lang, orig, word)
				log.Now().Debugf("Orig: %s, Word: %s, WordsTags: %s", orig, word, tags)
				tags_parsed = append(tags_parsed, (*p.repoa).TagWord(isReverse, lang, orig, word)...)
				if tags != "" && tags != "[]" {
					tags_parsed = append(tags_parsed, log.Error1(helpers.ParseJson[[]string]([]byte(tags)))...)
				}
			}
			if last_preferred != 0 && last_preferred == k || hash_preferred == hash.StringHash(0, word) {
				log.Now().Debugf("Preferred: %v, row: %d %d", word, last_preferred, hash_preferred)
				tags_parsed = append(tags_parsed, "preferred")
			}
			log.Now().Debugf("Orig: %v, Dest: %v, Tags: %v", orig, word, tags_parsed)
			wordstags[[2]string{orig, word}] = tags_parsed
		}
		log.Now().Debugf("WordsTags: %v, Words: %v", wordstags, words)
		intermediate = append(intermediate, wordstags)
	}

outer:
	for i, mapping := range intermediate {
		var next_tags = make(map[string]bool)
		if i+1 < len(intermediate) {
			var next_mapping = intermediate[i+1]
			for _, tags := range next_mapping {
				for _, tag := range tags {
					next_tags[tag] = true
				}
			}
		}
		for words, tags := range mapping {
			var my_tags = make(map[string]bool)
			for _, tag := range tags {
				my_tags[tag] = true
			}
			if next_continue_english(lang, isReverse, my_tags, next_tags) {
				continue
			}
			if !my_tags["preferred"] {
				continue
			}
			if !my_tags["dict"] {
				continue
			}
			var orig = words[0]
			var dest = words[1]
			ret = append(ret, [3]string{orig, dest, string(log.Error1(helpers.SerializeJson(tags)))})
			continue outer
		}
		for words, tags := range mapping {
			var my_tags = make(map[string]bool)
			for _, tag := range tags {
				my_tags[tag] = true
			}
			if next_continue_english(lang, isReverse, my_tags, next_tags) {
				continue
			}
			if !my_tags["dict"] {
				continue
			}
			var orig = words[0]
			var dest = words[1]
			ret = append(ret, [3]string{orig, dest, string(log.Error1(helpers.SerializeJson(tags)))})
			continue outer
		}
		for words, tags := range mapping {
			var my_tags = make(map[string]bool)
			for _, tag := range tags {
				my_tags[tag] = true
			}
			if next_continue_english(lang, isReverse, my_tags, next_tags) {
				continue
			}
			var orig = words[0]
			var dest = words[1]
			ret = append(ret, [3]string{orig, dest, string(log.Error1(helpers.SerializeJson(tags)))})
			continue outer
		}
		// in case of other bug push an empty word to keep punctuation algined
		ret = append(ret, [3]string{"", "", "[]"})
	}
	return
}

func NewPartsOfSpeechSelectorService(di *DependencyInjection) *PartsOfSpeechSelectorService {
	repoiface := (repo.IDictPhonemizerRepository)(Ptr(MustNeed(di, repo.NewDictPhonemizerRepository)))
	arepoiface := (repo.IAutoTaggerRepository)(Ptr(MustNeed(di, repo.NewAutoTaggerRepository)))
	airepoiface := (repo.IHashtronHomonymSelectorRepository)(Ptr(MustNeed(di, repo.NewHashtronHomonymSelectorRepository)))
	return &PartsOfSpeechSelectorService{
		repo:   &repoiface,
		repoa:  &arepoiface,
		repoai: &airepoiface,
	}
}

var _ IPartsOfSpeechSelectorService = &PartsOfSpeechSelectorService{}
