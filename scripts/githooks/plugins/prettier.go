package plugins

import (
	"errors"

	"githooks/engine"
)

func init() {
	engine.RegisterPlugin("prettier", func(e *engine.Engine, dir string, files []string) error {
		args := append([]string{"./node_modules/.bin/prettier", "--check"}, files...)
		if err := e.RunCommand(dir, "node", args...); err != nil {
			return errors.New("frontend formatting check failed; please run 'pnpm --dir app run format'")
		}
		return nil
	})
}
