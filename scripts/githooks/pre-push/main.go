package main

import (
	"githooks/engine"
	_ "githooks/plugins"
)

func main() {
	e := engine.NewEngine()
	e.FileResolver = engine.GetBranchChangedFiles

	// 1. Go Testing Task
	e.Register(engine.Task{
		Name:       "Go Backend Tests",
		Dir:        "server",
		Pattern:    ".go",
		PluginName: "go-test",
	})

	// 2. Frontend Testing Task
	e.Register(engine.Task{
		Name:       "Frontend Vue/JS Tests",
		Dir:        "app",
		Pattern:    ".js,.ts,.vue",
		PluginName: "vitest",
	})

	e.Run()
}
