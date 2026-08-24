package plugins

import (
	"errors"

	"githooks/engine"
)

func init() {
	engine.RegisterPlugin("go-test", func(e *engine.Engine, dir string, files []string) error {
		if err := e.RunCommand(dir, "go", "test", "./...", "-count=1"); err != nil {
			return errors.New("Go backend tests failed")
		}
		return nil
	})
}
