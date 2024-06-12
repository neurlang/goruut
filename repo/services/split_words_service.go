package services

import (
	"github.com/neurlang/goruut/repo"
)

import . "github.com/martinarisk/di/dependency_injection"

type ISplitWordsService interface {
	SplitWords(string, string) []string
}

type SplitWordsService struct {
	repo repo.ISpaceSplitterRepository
}

func (s *SplitWordsService) SplitWords(lang, sentence string) []string {

	switch lang {
	default:
		return s.repo.Split(sentence)
	}
}

func NewSplitWordsService(di *DependencyInjection) *SplitWordsService {
	repo := MustNeed(di, repo.NewSpaceSplitterRepository)

	return &SplitWordsService{
		repo: &repo,
	}
}

var _ ISplitWordsService = &SplitWordsService{}
