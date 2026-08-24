package main

import (
	"githooks/engine"
	_ "githooks/plugins"
)

func main() {
	e := engine.NewEngine()
	e.FileResolver = engine.GetStagedFiles

	// 1. Go Formatting Check (gofmt)
	e.Register(engine.Task{
		Name:       "Go Formatting Check",
		Dir:        "server",
		Pattern:    ".go",
		PluginName: "gofmt",
	})

	// 2. Go Linting Task (golangci-lint)
	e.Register(engine.Task{
		Name:       "Go Linting",
		Dir:        "server",
		Pattern:    ".go",
		DependsOn:  []string{"Go Formatting Check"},
		PluginName: "golangci-lint",
	})

	// 3. Frontend Formatting Task (prettier)
	e.Register(engine.Task{
		Name:       "Frontend Formatting Check",
		Dir:        "app",
		Pattern:    ".js,.ts,.vue,.css,.json",
		PluginName: "prettier",
	})

	// 4. Frontend Linting Task (eslint-oxlint)
	e.Register(engine.Task{
		Name:       "Frontend Linting",
		Dir:        "app",
		Pattern:    ".js,.ts,.vue",
		DependsOn:  []string{"Frontend Formatting Check"},
		PluginName: "eslint-oxlint",
	})

	e.Run()
}
