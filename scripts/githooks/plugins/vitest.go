package plugins

import (
	"errors"

	"githooks/engine"
)

func init() {
	engine.RegisterPlugin("vitest", func(e *engine.Engine, dir string, files []string) error {
		if err := e.RunCommand(dir, "pnpm", "exec", "vitest", "run"); err != nil {
			return errors.New("frontend tests failed")
		}
		return nil
	})
}
