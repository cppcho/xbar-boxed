package boxed

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

// CommandRunner abstracts subprocess execution for testability.
type CommandRunner interface {
	// Run executes a command and waits for it to finish.
	Run(name string, args ...string) error
	// Start executes a command without waiting (fire-and-forget).
	Start(name string, args ...string) error
}

// RealRunner executes real subprocesses.
type RealRunner struct{}

func (r RealRunner) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

func (r RealRunner) Start(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Start()
}

// RecordingRunner captures command invocations for testing.
type RecordingRunner struct {
	mu     sync.Mutex
	Runs   []string
	Starts []string
}

func (r *RecordingRunner) Run(name string, args ...string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Runs = append(r.Runs, formatCmd(name, args))
	return nil
}

func (r *RecordingRunner) Start(name string, args ...string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Starts = append(r.Starts, formatCmd(name, args))
	return nil
}

func formatCmd(name string, args []string) string {
	parts := append([]string{name}, args...)
	return fmt.Sprintf("%s", strings.Join(parts, " "))
}
