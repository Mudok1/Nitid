package cli

import (
	"path/filepath"

	"nitid/internal/core"
)

func newCoreService() (*core.Service, error) {
	root, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}
	return core.New(root), nil
}
