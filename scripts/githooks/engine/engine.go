package engine

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// TaskStatus represents the current execution state of a task.
type TaskStatus int

const (
	StatusPending TaskStatus = iota
	StatusSuccess
	StatusFailed
	StatusSkipped // Task matched no staged/changed files
	StatusBlocked // Task skipped because its dependencies failed
)

// Task defines a check to run.
type Task struct {
	Name      string
	Dir       string   // Directory from which this task runs (e.g. "server" or "app"). Empty means root.
	Pattern   string   // Comma-separated list of file extensions/suffixes to match
	DependsOn []string // List of Task Names this task depends on
	Skip      bool     // If true, skip this task but allow dependent tasks to run
	Check     func(dir string, files []string) error
}

// Engine manages and executes tasks.
type Engine struct {
	Tasks        []Task
	FileResolver func() ([]string, error) // Returns list of files to check (staged files vs branch files)
	logFile      *os.File
}

// NewEngine creates a new Engine instance.
func NewEngine() *Engine {
	return &Engine{}
}

// Register adds a task to the engine.
func (e *Engine) Register(task Task) {
	e.Tasks = append(e.Tasks, task)
}

// Logf prints a formatted message to both standard output and the log file.
func (e *Engine) Logf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Print(msg)
	if e.logFile != nil {
		_, _ = e.logFile.Write([]byte(msg))
	}
}

// Logln prints a message and a newline to both standard output and the log file.
func (e *Engine) Logln(a ...interface{}) {
	msg := fmt.Sprintln(a...)
	fmt.Print(msg)
	if e.logFile != nil {
		_, _ = e.logFile.Write([]byte(msg))
	}
}

// Run executes all tasks matching files resolved by FileResolver, in dependency order.
func (e *Engine) Run() {
	// Dynamically resolve log file path (e.g., .githooks/pre-commit.log)
	logPath := os.Args[0] + ".log"
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err == nil {
		e.logFile = file
		defer e.logFile.Close()
	} else {
		fmt.Printf("⚠️ Warning: Could not create log file %s: %v. Running checks without file logging.\n", logPath, err)
	}

	if e.FileResolver == nil {
		e.Logln("❌ Error: FileResolver is not configured in engine.")
		if e.logFile != nil {
			_ = e.logFile.Close()
		}
		os.Exit(1)
	}

	files, err := e.FileResolver()
	if err != nil {
		e.Logf("❌ Error resolving files: %v\n", err)
		if e.logFile != nil {
			_ = e.logFile.Close()
			fmt.Printf("📝 Full logs written to: %s. You can review it anytime!\n", logPath)
		}
		os.Exit(1)
	}

	if len(files) == 0 {
		e.Logln("ℹ️ No changed files found. Skipping checks.")
		if e.logFile != nil {
			_ = e.logFile.Close()
			fmt.Printf("📝 Full logs written to: %s. You can review it anytime!\n", logPath)
		}
		os.Exit(0)
	}

	// Initialize status map for all registered tasks
	status := make(map[string]TaskStatus)
	for _, task := range e.Tasks {
		status[task.Name] = StatusPending
	}

	var hasFailures bool

	for {
		progress := false
		pendingCount := 0

		for _, task := range e.Tasks {
			if status[task.Name] != StatusPending {
				continue
			}
			pendingCount++

			// Check if dependencies are resolved
			depFailed := false
			depNotResolved := false

			for _, depName := range task.DependsOn {
				depStatus, exists := status[depName]
				if !exists {
					// Reference to a non-existent task is considered resolved
					continue
				}
				switch depStatus {
				case StatusPending:
					depNotResolved = true
				case StatusFailed, StatusBlocked:
					depFailed = true
				}
			}

			if depNotResolved {
				// Wait for dependency task to finish first
				continue
			}

			if depFailed {
				e.Logf("⏭️ Skipping task: %s (a dependency failed)\n", task.Name)
				e.Logln(strings.Repeat("-", 50))
				status[task.Name] = StatusBlocked
				progress = true
				continue
			}

			if task.Skip {
				e.Logf("⏭️ Skipping task: %s (explicitly skipped in configuration)\n", task.Name)
				e.Logln(strings.Repeat("-", 50))
				status[task.Name] = StatusSkipped
				progress = true
				continue
			}

			// Gather files matching this task's filters
			var matched []string
			for _, file := range files {
				if task.Dir != "" {
					prefix := task.Dir + "/"
					if !strings.HasPrefix(file, prefix) {
						continue
					}
				}

				if MatchPattern(file, task.Pattern) {
					relPath := file
					if task.Dir != "" {
						relPath = strings.TrimPrefix(file, task.Dir+"/")
					}
					matched = append(matched, relPath)
				}
			}

			// If no files match, skip the task
			if len(matched) == 0 {
				status[task.Name] = StatusSkipped
				progress = true
				continue
			}

			dirName := task.Dir
			if dirName == "" {
				dirName = "root"
			}
			e.Logf("🏃 Running task: %s (%d files matched in '%s')...\n", task.Name, len(matched), dirName)

			dirPath := task.Dir
			if dirPath == "" {
				dirPath = "."
			}

			if err := task.Check(dirPath, matched); err != nil {
				e.Logf("❌ Task '%s' failed: %v\n", task.Name, err)
				status[task.Name] = StatusFailed
				hasFailures = true
			} else {
				e.Logf("✅ Task '%s' passed!\n", task.Name)
				status[task.Name] = StatusSuccess
			}
			e.Logln(strings.Repeat("-", 50))
			progress = true
		}

		if pendingCount == 0 {
			break
		}

		// If no task moved from Pending state, we have a cycle or unresolved dependency configuration
		if !progress {
			e.Logln("❌ Error: Circular or unresolved dependencies detected in tasks configuration!")
			if e.logFile != nil {
				_ = e.logFile.Close()
				fmt.Printf("📝 Full logs written to: %s. You can review it anytime!\n", logPath)
			}
			os.Exit(1)
		}
	}

	if hasFailures {
		e.Logln("❌ Hook checks failed! Please fix the errors above.")
		if e.logFile != nil {
			_ = e.logFile.Close()
			fmt.Printf("📝 Full logs written to: %s. You can review it anytime!\n", logPath)
		}
		os.Exit(1)
	}

	e.Logln("🎉 All checks passed successfully!")
	if e.logFile != nil {
		_ = e.logFile.Close()
		fmt.Printf("📝 Full logs written to: %s. You can review it anytime!\n", logPath)
	}
	os.Exit(0)
}

// GetStagedFiles: Query staged files using git (useful for pre-commit)
func GetStagedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only", "--diff-filter=ACM")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return parseLines(stdout.String()), nil
}

// GetBranchChangedFiles: Query files changed in current branch compared to remote tracking branch (useful for pre-push)
func GetBranchChangedFiles() ([]string, error) {
	// 1. Try to diff against upstream branch
	cmd := exec.Command("git", "diff", "--name-only", "@{u}...HEAD")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err == nil {
		return parseLines(stdout.String()), nil
	}

	// 2. Fallback: diff against local main branch
	cmd = exec.Command("git", "diff", "--name-only", "main...HEAD")
	stdout.Reset()
	cmd.Stdout = &stdout
	if err := cmd.Run(); err == nil {
		return parseLines(stdout.String()), nil
	}

	// 3. Fallback: diff against origin/main
	cmd = exec.Command("git", "diff", "--name-only", "origin/main...HEAD")
	stdout.Reset()
	cmd.Stdout = &stdout
	if err := cmd.Run(); err == nil {
		return parseLines(stdout.String()), nil
	}

	// 4. Fallback: list all tracked files in repository to force checking everything
	cmd = exec.Command("git", "ls-files")
	stdout.Reset()
	cmd.Stdout = &stdout
	if err := cmd.Run(); err == nil {
		return parseLines(stdout.String()), nil
	}

	return nil, fmt.Errorf("failed to get files list from git")
}

// MatchPattern: Match file path against comma-separated extensions
func MatchPattern(file, pattern string) bool {
	if pattern == "" {
		return true
	}
	suffixes := strings.Split(pattern, ",")
	for _, suffix := range suffixes {
		if strings.HasSuffix(file, strings.TrimSpace(suffix)) {
			return true
		}
	}
	return false
}

// RunCommand: Run command with cleaned Git environment variables to avoid VCS resolution conflicts
func (e *Engine) RunCommand(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	// Clean Git local env vars
	gitVars := getGitLocalEnvVars()
	var env []string
	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2)
		isGitVar := false
		for _, gv := range gitVars {
			if parts[0] == gv {
				isGitVar = true
				break
			}
		}
		if !isGitVar {
			env = append(env, envVar)
		}
	}
	cmd.Env = env

	// Use MultiWriter to redirect execution output to both stdout/stderr and the log file
	if e.logFile != nil {
		cmd.Stdout = io.MultiWriter(os.Stdout, e.logFile)
		cmd.Stderr = io.MultiWriter(os.Stderr, e.logFile)
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

func parseLines(output string) []string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var files []string
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			files = append(files, trimmed)
		}
	}
	return files
}

func getGitLocalEnvVars() []string {
	return []string{
		"GIT_ALTERNATE_OBJECT_DIRECTORIES",
		"GIT_CONFIG",
		"GIT_CONFIG_PARAMETERS",
		"GIT_CONFIG_COUNT",
		"GIT_OBJECT_DIRECTORY",
		"GIT_DIR",
		"GIT_WORK_TREE",
		"GIT_IMPLICIT_WORK_TREE",
		"GIT_GRAFT_FILE",
		"GIT_INDEX_FILE",
		"GIT_NO_REPLACE_OBJECTS",
		"GIT_REPLACE_REF_BASE",
		"GIT_PREFIX",
		"GIT_SHALLOW_FILE",
		"GIT_COMMON_DIR",
	}
}
