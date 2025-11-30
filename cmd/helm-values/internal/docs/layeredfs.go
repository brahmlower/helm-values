package docs

import (
	"io/fs"
)

func NewLayeredFS(layers ...fs.FS) *LayeredFS {
	return &LayeredFS{
		layers: layers,
	}
}

type LayeredFS struct {
	layers []fs.FS
}

var _ fs.FS = (*LayeredFS)(nil)
var _ fs.GlobFS = (*LayeredFS)(nil)
var _ fs.ReadFileFS = (*LayeredFS)(nil)

func (l *LayeredFS) Open(name string) (fs.File, error) {
	var lastErr error
	var f fs.File
	for _, layer := range l.layers {
		f, lastErr = layer.Open(name)
		if lastErr == nil {
			return f, nil
		}
	}

	return nil, lastErr
}

func (l *LayeredFS) Glob(name string) ([]string, error) {
	return []string{name}, nil
}

func (l *LayeredFS) ReadFile(name string) ([]byte, error) {
	var lastErr error
	var b []byte
	for _, layer := range l.layers {
		b, lastErr = fs.ReadFile(layer, name)
		if lastErr == nil {
			return b, nil
		}
	}

	return nil, lastErr
}
