package services

import (
	"github.com/neurlang/goruut/repo"
)
import . "github.com/martinarisk/di/dependency_injection"

type IPhonemizeWordService interface {
	PhonemizeWord(string, string) map[uint64]string
}

type PhonemizeWordService struct {
	repo *repo.IDictPhonemizerRepository
	ai   *repo.IHashtronPhonemizerRepository
	pre  *repo.IPrePhonWordStepsRepository
	cach *repo.IWordCachingRepository
}

func (p *PhonemizeWordService) PhonemizeWord(lang, word string) (ret map[uint64]string) {

	word = (*p.pre).PrePhonemizeWord(lang, word)

	ret = (*p.repo).PhonemizeWord(lang, word)
	if ret == nil {
		ret = (*p.cach).LoadWord((*p.cach).HashWord(lang, word))

		if ret == nil || len(ret) == 0 {
			ret = (*p.ai).PhonemizeWord(lang, word)

			(*p.cach).StoreWord(ret, (*p.cach).HashWord(lang, word))
		}
	}
	return
}

func NewPhonemizeWordService(di *DependencyInjection) *PhonemizeWordService {
	repository := MustNeed(di, repo.NewDictPhonemizerRepository)
	repoiface := (repo.IDictPhonemizerRepository)(&repository)
	ai_repo := MustNeed(di, repo.NewHashtronPhonemizerRepository)
	ai_repo_iface := (repo.IHashtronPhonemizerRepository)(&ai_repo)
	pre_repo := MustNeed(di, repo.NewPrePhonWordStepsRepository)
	pre_repo_iface := (repo.IPrePhonWordStepsRepository)(&pre_repo)
	cach_repo := MustNeed(di, repo.NewWordCachingRepository)
	cach_repo_iface := (repo.IWordCachingRepository)(&cach_repo)

	return &PhonemizeWordService{
		repo: &repoiface,
		ai:   &ai_repo_iface,
		pre:  &pre_repo_iface,
		cach: &cach_repo_iface,
	}
}

var _ IPhonemizeWordService = &PhonemizeWordService{}
