package services

import (
	"github.com/neurlang/goruut/repo"
	"strings"
)
import . "github.com/martinarisk/di/dependency_injection"

type IPhonemizeWordService interface {
	PhonemizeWord(string, string) map[uint64]string
}

type PhonemizeWordService struct {
	repo *repo.IDictPhonemizerRepository
	ai   *repo.IHashtronPhonemizerRepository
}

func (p *PhonemizeWordService) PhonemizeWord(lang, word string) (ret map[uint64]string) {

	word = strings.Trim(word, ".,")

	word = strings.ToLower(word)

	ret = (*p.repo).PhonemizeWord(lang, word)
	if ret == nil {
		ret = (*p.ai).PhonemizeWord(lang, word)
	}
	return
}

func NewPhonemizeWordService(di *DependencyInjection) *PhonemizeWordService {
	repository := MustNeed(di, repo.NewDictPhonemizerRepository)
	repoiface := (repo.IDictPhonemizerRepository)(&repository)
	ai_repo := MustNeed(di, repo.NewHashtronPhonemizerRepository)
	ai_repo_iface := (repo.IHashtronPhonemizerRepository)(&ai_repo)

	return &PhonemizeWordService{
		repo: &repoiface,
		ai:   &ai_repo_iface,
	}
}

var _ IPhonemizeWordService = &PhonemizeWordService{}
