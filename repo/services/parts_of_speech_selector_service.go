package services

import (
	"github.com/neurlang/goruut/helpers"
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/repo"
	"strings"
)
import . "github.com/martinarisk/di/dependency_injection"

type IPartsOfSpeechSelectorService interface {
	Select(isReverse bool, lang string, sentence []map[uint32]string) (ret [][3]string)
}

type PartsOfSpeechSelectorService struct {
	repo *repo.IDictPhonemizerRepository
	repoa *repo.IAutoTaggerRepository
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

func (p *PartsOfSpeechSelectorService) Select(isReverse bool, lang string, sentence []map[uint32]string) (ret [][3]string) {
	var intermediate []map[[2]string][]string
	for _, words := range sentence {
		var orig = words[0]
		var wordstags = make(map[[2]string][]string)
		for k, word := range words {
			if k == 0 {
				orig = word
				continue
			}
			var tags = (*p.repo).LookupTags(isReverse, lang, orig, word)
			log.Now().Debugf("Orig: %s, Word: %s, WordsTags: %s", orig, word, tags)
			var tags_parsed = (*p.repoa).TagWord(isReverse, lang, orig, word)
			if tags != "" && tags != "[]" {
				tags_parsed = append(tags_parsed, log.Error1(helpers.ParseJson[[]string]([]byte(tags)))...)
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
		if i + 1 < len(intermediate) {
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
			if !my_tags["dict"] {continue;}
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
	}
	return
}

func NewPartsOfSpeechSelectorService(di *DependencyInjection) *PartsOfSpeechSelectorService {
	repoiface := (repo.IDictPhonemizerRepository)(Ptr(MustNeed(di, repo.NewDictPhonemizerRepository)))
	arepoiface := (repo.IAutoTaggerRepository)(Ptr(MustNeed(di, repo.NewAutoTaggerRepository)))
	return &PartsOfSpeechSelectorService{
		repo: &repoiface,
		repoa: &arepoiface,
	}
}

var _ IPartsOfSpeechSelectorService = &PartsOfSpeechSelectorService{}
