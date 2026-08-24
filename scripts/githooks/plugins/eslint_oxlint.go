package plugins

import (
	"errors"

	"githooks/engine"
)

func init() {
	engine.RegisterPlugin("eslint-oxlint", func(e *engine.Engine, dir string, files []string) error {
		oxlintArgs := append([]string{"./node_modules/.bin/oxlint"}, files...)
		if err := e.RunCommand(dir, "node", oxlintArgs...); err != nil {
			return errors.New("oxlint linting failed")
		}

		eslintArgs := append([]string{"./node_modules/.bin/eslint"}, files...)
		if err := e.RunCommand(dir, "node", eslintArgs...); err != nil {
			return errors.New("eslint linting failed")
		}

		return nil
	})
}
