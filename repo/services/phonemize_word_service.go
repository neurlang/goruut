package services

import (
	"github.com/neurlang/goruut/repo"
)
import . "github.com/martinarisk/di/dependency_injection"

type IPhonemizeWordService interface {
	PhonemizeWords(string, string) []map[uint64]string
	CleanWord(lang, word string) string
}

type PhonemizeWordService struct {
	repo *repo.IDictPhonemizerRepository
	ai   *repo.IHashtronPhonemizerRepository
	pre  *repo.IPrePhonWordStepsRepository
	cach *repo.IWordCachingRepository
}

func (p *PhonemizeWordService) PhonemizeWords(lang, word string) (ret []map[uint64]string) {
	word = (*p.pre).PrePhonemizeWord(lang, word)

	word = (*p.ai).CleanWord(lang, word)
	if word == "" {
		return nil
	}
	ret = (*p.repo).LookupWords(lang, word)
	if ret == nil {
		hash := (*p.cach).HashWord(lang, word)
		r := (*p.cach).LoadWord(hash)
		if r == nil || len(r) == 0 {
			ret = (*p.ai).PhonemizeWords(lang, word)
			for _, one := range ret {
				(*p.cach).StoreWord(one, hash)
			}
		} else {
			ret = append(ret, r)
		}
	}
	return

}

func (p *PhonemizeWordService) CleanWord(lang, word string) string {
	return (*p.ai).CleanWord(lang, word)
}

func NewPhonemizeWordService(di *DependencyInjection) *PhonemizeWordService {
	repoiface := (repo.IDictPhonemizerRepository)(Ptr(MustNeed(di, repo.NewDictPhonemizerRepository)))
	ai_repo_iface := (repo.IHashtronPhonemizerRepository)(Ptr(MustNeed(di, repo.NewHashtronPhonemizerRepository)))
	pre_repo_iface := (repo.IPrePhonWordStepsRepository)(Ptr(MustNeed(di, repo.NewPrePhonWordStepsRepository)))
	cach_repo_iface := (repo.IWordCachingRepository)(Ptr(MustNeed(di, repo.NewWordCachingRepository)))

	return &PhonemizeWordService{
		repo: &repoiface,
		ai:   &ai_repo_iface,
		pre:  &pre_repo_iface,
		cach: &cach_repo_iface,
	}
}

var _ IPhonemizeWordService = &PhonemizeWordService{}
