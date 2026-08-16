package pkg

import (
	"fmt"
	"io/fs"
)

// LayeredFS is an fs.FS that overlays multiple filesystems, resolving each
// file lookup against the layers in order and returning the first match.
type LayeredFS struct {
	layers []fs.FS
}

// NewLayeredFS returns a LayeredFS that searches the given layers, in order,
// for each requested file.
func NewLayeredFS(layers ...fs.FS) *LayeredFS {
	return &LayeredFS{
		layers: layers,
	}
}

var _ fs.FS = (*LayeredFS)(nil)
var _ fs.GlobFS = (*LayeredFS)(nil)
var _ fs.ReadFileFS = (*LayeredFS)(nil)

// Open opens name by searching each layer in order, returning the first
// successful result or the last error encountered if no layer has the file.
func (l *LayeredFS) Open(name string) (fs.File, error) {
	var (
		lastErr error
		f       fs.File
	)
	for _, layer := range l.layers {
		f, lastErr = layer.Open(name)
		if lastErr == nil {
			return f, nil
		}
	}

	if lastErr != nil {
		lastErr = fmt.Errorf("open %q in any layer: %w", name, lastErr)
	}

	return nil, lastErr
}

// Glob always returns name as its only match; it does not perform real
// pattern matching against the underlying layers.
func (l *LayeredFS) Glob(name string) ([]string, error) {
	return []string{name}, nil
}

// ReadFile reads name by searching each layer in order, returning the first
// successful result or the last error encountered if no layer has the file.
func (l *LayeredFS) ReadFile(name string) ([]byte, error) {
	var (
		lastErr error
		b       []byte
	)
	for _, layer := range l.layers {
		b, lastErr = fs.ReadFile(layer, name)
		if lastErr == nil {
			return b, nil
		}
	}

	if lastErr != nil {
		lastErr = fmt.Errorf("read %q in any layer: %w", name, lastErr)
	}

	return nil, lastErr
}
