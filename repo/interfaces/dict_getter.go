package interfaces

type DictGetter interface {
	GetDict(lang, filename string) ([]byte, error)
}
