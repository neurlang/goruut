package interfaces

type ModelStorage interface {
	GetDict(lang, filename string) ([]byte, error)
	HaveLang(lang string) bool
}

type LoadModels interface {
	GetLoadModels() []*struct {
		Lang string
		File string
		Size int64
	}
}
