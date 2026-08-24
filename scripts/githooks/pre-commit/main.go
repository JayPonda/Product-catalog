package main

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"githooks/engine"
)

func main() {
	e := engine.NewEngine()
	e.FileResolver = engine.GetStagedFiles

	// 1. Go Formatting Check (gofmt)
	e.Register(engine.Task{
		Name:    "Go Formatting Check",
		Dir:     "server",
		Pattern: ".go",
		Check: func(dir string, files []string) error {
			args := append([]string{"-l"}, files...)
			cmd := exec.Command("gofmt", args...)
			cmd.Dir = dir

			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to execute gofmt: %w", err)
			}

			unformatted := strings.TrimSpace(stdout.String())
			if unformatted != "" {
				fmt.Println("❌ The following Go files are not formatted:")
				fmt.Println(unformatted)
				return errors.New("unformatted Go files found; please run 'make format-backend'")
			}
			return nil
		},
	})

	// 2. Go Linting Task (golangci-lint)
	e.Register(engine.Task{
		Name:      "Go Linting",
		Dir:       "server",
		Pattern:   ".go",
		DependsOn: []string{"Go Formatting Check"},
		Check: func(dir string, files []string) error {
			if err := e.RunCommand(dir, "go", "run", "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5", "run", "./..."); err != nil {
				return errors.New("Go linting failed")
			}
			return nil
		},
	})

	// 3. Frontend Formatting Task (prettier)
	e.Register(engine.Task{
		Name:    "Frontend Formatting Check",
		Dir:     "app",
		Pattern: ".js,.ts,.vue,.css,.json",
		Check: func(dir string, files []string) error {
			args := append([]string{"./node_modules/.bin/prettier", "--check"}, files...)
			if err := e.RunCommand(dir, "node", args...); err != nil {
				return errors.New("frontend formatting check failed; please run 'pnpm --dir app run format'")
			}
			return nil
		},
	})

	// 4. Frontend Linting Task (oxlint & eslint)
	e.Register(engine.Task{
		Name:      "Frontend Linting",
		Dir:       "app",
		Pattern:   ".js,.ts,.vue",
		DependsOn: []string{"Frontend Formatting Check"},
		Check: func(dir string, files []string) error {
			oxlintArgs := append([]string{"./node_modules/.bin/oxlint"}, files...)
			if err := e.RunCommand(dir, "node", oxlintArgs...); err != nil {
				return errors.New("oxlint linting failed")
			}

			eslintArgs := append([]string{"./node_modules/.bin/eslint"}, files...)
			if err := e.RunCommand(dir, "node", eslintArgs...); err != nil {
				return errors.New("eslint linting failed")
			}

			return nil
		},
	})

	e.Run()
}
