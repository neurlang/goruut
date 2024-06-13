package interfaces

type IpaFlavor interface {
	GetIpaFlavors() map[string]map[string]string
}
