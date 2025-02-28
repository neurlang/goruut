package services

import (
	"github.com/neurlang/goruut/repo"
)
import . "github.com/martinarisk/di/dependency_injection"

type IPhonemizeWordService interface {
	PhonemizeWords(isReverse bool, lang, word string) []map[uint64]string
	CleanWord(isReverse bool, lang, word string) string
}

type PhonemizeWordService struct {
	repo *repo.IDictPhonemizerRepository
	ai   *repo.IHashtronPhonemizerRepository
	pre  *repo.IPrePhonWordStepsRepository
	cach *repo.IWordCachingRepository
}

func (p *PhonemizeWordService) PhonemizeWords(isReverse bool, lang, word string) (ret []map[uint64]string) {
	word = (*p.pre).PrePhonemizeWord(isReverse, lang, word)

	word = (*p.ai).CleanWord(isReverse, lang, word)
	if word == "" {
		return nil
	}
	ret = (*p.repo).LookupWords(isReverse, lang, word)
	if ret == nil {
		hash := (*p.cach).HashWord(isReverse, lang, word)
		r := (*p.cach).LoadWord(hash)
		if r == nil || len(r) == 0 {
			ret = (*p.ai).PhonemizeWords(isReverse, lang, word)
			for i, one := range ret {
				//rett := (*p.ai).PhonemizeWord(isReverse, lang, one[0])
				//if len(rett) > 0 {
				//	one = rett
				//	ret[i] = rett
				//}
				(*p.cach).StoreWord(one, hash+uint64(i))
			}
		} else {
			ret = append(ret, r)
			for i := uint64(1); true; i++ {
				r = (*p.cach).LoadWord(hash + i)
				if r == nil || len(r) == 0 {
					break
				}
				ret = append(ret, r)
			}
		}
	}
	return

}

func (p *PhonemizeWordService) CleanWord(isReverse bool, lang, word string) string {
	return (*p.ai).CleanWord(isReverse, lang, word)
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
