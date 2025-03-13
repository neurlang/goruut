package requests

type PhonemizeSentence struct {
	IpaFlavors []string
	Language   string
	Languages  []string
	Sentence   string
	IsReverse  bool
}

func (p *PhonemizeSentence) Init() {
	if p.Language == "" && len(p.Languages) > 0 {
		p.Language = p.Languages[0]
	}
}
