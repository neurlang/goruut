package services

import (
	"github.com/neurlang/goruut/repo"
)
import . "github.com/martinarisk/di/dependency_injection"

type IPhonemizeWordService interface {
	PhonemizeWord(string, string) (string, map[uint64]string)
}

type PhonemizeWordService struct {
	repo *repo.IDictPhonemizerRepository
	ai   *repo.IHashtronPhonemizerRepository
	pre  *repo.IPrePhonWordStepsRepository
	cach *repo.IWordCachingRepository
}

func (p *PhonemizeWordService) PhonemizeWord(lang, word string) (wrd string, ret map[uint64]string) {

	word = (*p.pre).PrePhonemizeWord(lang, word)

	wrd = (*p.ai).CleanWord(lang, word)
	hsh := (*p.cach).HashWord(lang, wrd)

	ret = (*p.repo).PhonemizeWord(lang, wrd)
	if ret == nil {
		ret = (*p.cach).LoadWord(hsh)

		if ret == nil || len(ret) == 0 {
			ret = (*p.ai).PhonemizeWord(lang, wrd)

			(*p.cach).StoreWord(ret, hsh)
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
