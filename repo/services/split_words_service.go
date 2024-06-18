package services

import (
	"github.com/neurlang/goruut/repo"
)

import . "github.com/martinarisk/di/dependency_injection"

type ISplitWordsService interface {
	SplitWords(string, string) []string
}

type SplitWordsService struct {
	repo1 repo.ISpaceSplitterRepository
	repo2 repo.ISpacerSplitterRepository
}

func (s *SplitWordsService) SplitWords(lang, sentence string) (out []string) {
	tmp := s.repo1.Split(sentence)
	for _, subsentence := range tmp {
		splitted := s.repo2.Split(lang, subsentence)
		out = append(out, splitted...)
	}
	return
}

func NewSplitWordsService(di *DependencyInjection) *SplitWordsService {
	repo1 := MustNeed(di, repo.NewSpaceSplitterRepository)
	repo2 := MustNeed(di, repo.NewSpacerSplitterRepository)

	return &SplitWordsService{
		repo1: &repo1,
		repo2: &repo2,
	}
}

var _ ISplitWordsService = &SplitWordsService{}
