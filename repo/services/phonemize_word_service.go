package services

import (
	"github.com/neurlang/goruut/repo"
)
import . "github.com/martinarisk/di/dependency_injection"

type IPhonemizeWordService interface {
	PhonemizeWord(string, string) (string, map[uint64]string)
	PhonemizeWordCJK(string, string) (string, map[uint64][2]string)
	CleanWord(lang, word string) string
}

type PhonemizeWordService struct {
	repo *repo.IDictPhonemizerRepository
	ai   *repo.IHashtronPhonemizerRepository
	pre  *repo.IPrePhonWordStepsRepository
	cach *repo.IWordCachingRepository
}

func (p *PhonemizeWordService) PhonemizeWord(lang, word string) (wrd string, ret map[uint64]string) {
	wrd, r := p.PhonemizeWordCJK(lang, word)
	ret = make(map[uint64]string)
	for k, v := range r {
		ret[k] = v[0]
	}
	return
}

func (p *PhonemizeWordService) CleanWord(lang, word string) string {
	return (*p.ai).CleanWord(lang, word)
}

func (p *PhonemizeWordService) PhonemizeWordCJK(lang, word string) (wrd string, ret map[uint64][2]string) {

	word = (*p.pre).PrePhonemizeWord(lang, word)

	wrd = (*p.ai).CleanWord(lang, word)
	if wrd == "" {
		ret = make(map[uint64][2]string)
		ret[0] = [2]string{"", ""}
		return
	}
	hsh := (*p.cach).HashWord(lang, wrd)

	ret = (*p.repo).PhonemizeWordCJK(lang, wrd)
	if ret == nil {
		ret = (*p.cach).LoadWordCJK(hsh)
		if ret != nil {
			for k, ipa := range ret {
				if !(*p.ai).CheckWord(lang, wrd, ipa[0]) {
					delete(ret, k)
				}
			}
		}

		if ret == nil || len(ret) == 0 {
			ret = (*p.ai).PhonemizeWordCJK(lang, wrd)

			(*p.cach).StoreWordCJK(ret, hsh)
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
