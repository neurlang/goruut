package services

import (
	"github.com/neurlang/goruut/repo"
)

import . "github.com/martinarisk/di/dependency_injection"

type ISplitWordsService interface {
	SplitWords(string, string) []string
}

type SplitWordsService struct {
	repo1 *repo.ISpaceSplitterRepository
}

func (s *SplitWordsService) SplitWords(lang, sentence string) (out []string) {
	return (*s.repo1).SplitLang(lang, sentence)
}

func NewSplitWordsService(di *DependencyInjection) *SplitWordsService {
	repo1 := (repo.ISpaceSplitterRepository)(Ptr(MustNeed(di, repo.NewSpaceSplitterRepository)))

	return &SplitWordsService{
		repo1: &repo1,
	}
}

var _ ISplitWordsService = &SplitWordsService{}
