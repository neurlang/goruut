package repo

import . "github.com/martinarisk/di/dependency_injection"

type IAutoTaggerRepository interface {
	TagWord(isReverse bool, lang, word1, word2 string) []string
	IsCrossDictWord(isReverse bool, lang, word string) bool
}
type AutoTaggerRepository struct {
	english_consonants *map[rune]struct{}
	english_vowels     *map[rune]struct{}
}

func (r *AutoTaggerRepository) IsCrossDictWord(isReverse bool, lang, word string) bool {
	if !isReverse && lang == "English" {
		return word == "the"
	}
	return false
}

func (r *AutoTaggerRepository) TagWord(isReverse bool, lang, word1, word2 string) []string {
	if !isReverse && lang == "English" {
		if word1 == "the" {
			if word2 == "ðə" {
				return []string{"the"}
			}
			if word2 == "ðɪ" || word2 == "ðˈi" {
				return []string{"thi"}
			}
		}
		for _, c := range []rune(word1) {
			if _, ok := (*r.english_consonants)[c]; ok {
				return []string{"consonant1st"}
			}
			if _, ok := (*r.english_vowels)[c]; ok {
				return []string{"vowel1st"}
			}
		}
	}
	return nil
}


func NewAutoTaggerRepository(di *DependencyInjection) *AutoTaggerRepository {
	var english_consonants = make(map[rune]struct{})
	var english_vowels = make(map[rune]struct{})
	for _, c := range [...]rune{'a', 'e', 'i', 'o', 'u'} {
		english_vowels[c] = struct{}{}
	}
	for _, c := range [...]rune{'b', 'c', 'd', 'f', 'g', 'h', 'j', 'k', 'l', 'm', 'n', 'p', 'q', 'r', 's', 't', 'v', 'w', 'x', 'y', 'z'} {
		english_consonants[c] = struct{}{}
	}

	return &AutoTaggerRepository{
		english_consonants: &english_consonants,
		english_vowels: &english_vowels,
	}
}

var _ IAutoTaggerRepository = &AutoTaggerRepository{}
