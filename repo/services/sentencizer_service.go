package services

import (
	"github.com/neurlang/goruut/repo"
	"github.com/sentencizer/sentencizer"
)

import . "github.com/martinarisk/di/dependency_injection"

type ISentencizerService interface {
	Split(string, string) []string
}

type SentencizerService struct {
	repo1 *repo.ISpaceSplitterRepository
}

func (s *SentencizerService) Split(lang, text string) (out []string) {
	switch lang {
	case "Hebrew", "Hebrew2", "Hebrew3":
		segmenter := sentencizer.NewSegmenter("he")
		return segmenter.Segment(text)
	case "English", "EnglishAmerican", "EnglishBritish":
		segmenter := sentencizer.NewSegmenter("en")
		return segmenter.Segment(text)
	}
	return []string{text}
}

func NewSentencizerService(di *DependencyInjection) *SentencizerService {

	return &SentencizerService{}
}

var _ ISentencizerService = &SentencizerService{}
