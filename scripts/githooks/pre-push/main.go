package main

import (
	"errors"

	"githooks/engine"
)

func main() {
	e := engine.NewEngine()
	e.FileResolver = engine.GetBranchChangedFiles

	// 1. Go Testing Task
	e.Register(engine.Task{
		Name:    "Go Backend Tests",
		Dir:     "server",
		Pattern: ".go",
		Check: func(dir string, files []string) error {
			if err := e.RunCommand(dir, "go", "test", "./...", "-count=1"); err != nil {
				return errors.New("Go backend tests failed")
			}
			return nil
		},
	})

	// 2. Frontend Testing Task
	e.Register(engine.Task{
		Name:    "Frontend Vue/JS Tests",
		Dir:     "app",
		Pattern: ".js,.ts,.vue",
		Check: func(dir string, files []string) error {
			if err := e.RunCommand(dir, "pnpm", "exec", "vitest", "run"); err != nil {
				return errors.New("frontend tests failed")
			}
			return nil
		},
	})

	e.Run()
}
