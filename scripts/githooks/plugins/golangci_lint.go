package plugins

import (
	"errors"

	"githooks/engine"
)

func init() {
	engine.RegisterPlugin("golangci-lint", func(e *engine.Engine, dir string, files []string) error {
		if err := e.RunCommand(dir, "go", "run", "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5", "run", "./..."); err != nil {
			return errors.New("Go linting failed")
		}
		return nil
	})
}
