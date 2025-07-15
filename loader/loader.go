// Package loader implements an external language model loader based on zip files
package loader

import (
	"archive/zip"
	"github.com/neurlang/goruut/helpers/log"
	"github.com/neurlang/goruut/repo/interfaces"
	"io"
	"os"
)
import . "github.com/martinarisk/di/dependency_injection"

type ILoader interface {
	interfaces.DictGetter
}

type Loader struct {
	g interfaces.DictGetter
	d *map[string]string
}

func (l *Loader) GetDict(lang, file string) ([]byte, error) {
	if zipfile, ok := (*l.d)[lang]; ok {
		reader := log.Error1(zip.OpenReader(zipfile))
		if reader != nil {
			f := log.Error1(reader.Open(file))
			if f != nil {
				data := log.Error1(io.ReadAll(f))
				if len(data) > 0 {
					log.Now().Infof("Loader used file: %s %s %s", lang, zipfile, file)
					return data, nil
				}
			}
		}
	}

	return l.g.GetDict(lang, file)
}
func (l *Loader) IsNewFormat(data []byte) bool {
	return l.g.IsNewFormat(data)
}
func (l *Loader) IsOldFormat(data []byte) bool {
	return l.g.IsNewFormat(data)
}

func NewLoader(di *DependencyInjection) *Loader {
	getter := MustAny[interfaces.DictGetter](di)
	loadModels := MustAny[interfaces.LoadModels](di)
	data := make(map[string]string)

	for _, m := range loadModels.GetLoadModels() {
		if m == nil {
			continue
		}
		if m.Lang == "" {
			continue
		}
		if m.File == "" {
			continue
		}
		if m.Size != 0 {
			fi := log.Error1(os.Stat(m.File))
			realSize := fi.Size()
			if realSize != m.Size {
				continue
			}
		}
		data[m.Lang] = m.File
		log.Now().Infof("Loaded language %s as %s", m.Lang, m.File)
	}
	// remove the previous getter
	di.Remove(getter)

	return &Loader{
		g: getter,
		d: &data,
	}
}

var _ ILoader = &Loader{}
