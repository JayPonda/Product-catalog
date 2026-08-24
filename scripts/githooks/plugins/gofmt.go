package plugins

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"githooks/engine"
)

func init() {
	engine.RegisterPlugin("gofmt", func(e *engine.Engine, dir string, files []string) error {
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
			e.Logln("❌ The following Go files are not formatted:")
			e.Logln(unformatted)
			return errors.New("unformatted Go files found; please run 'make format-backend'")
		}
		return nil
	})
}
