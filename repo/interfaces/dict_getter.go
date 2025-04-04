// Package interfaces defines repository contracts
package interfaces

type DictGetter interface {
	GetDict(lang, filename string) ([]byte, error)
	IsOldFormat(magic []byte) bool
	IsNewFormat(magic []byte) bool
}
