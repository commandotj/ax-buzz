package brew

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

func RunPackageAction(ctx context.Context, action, name string, isCask bool, progress func(string)) error {
	if name == "" {
		return fmt.Errorf("package name is required")
	}
	args := []string{action}
	if isCask {
		args = append(args, "--cask")
	}
	args = append(args, name)
	return runBrew(ctx, args, progress)
}

func runBrew(ctx context.Context, args []string, progress func(string)) error {
	if brewExecutableErr != nil {
		return brewExecutableErr
	}
	cmd := exec.CommandContext(ctx, brewExecutable, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("open Homebrew stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("open Homebrew stderr: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start Homebrew: %w", err)
	}

	lines := make(chan string, 64)
	var readers sync.WaitGroup
	readers.Add(2)
	readLines := func(pipe interface{ Read([]byte) (int, error) }) {
		defer readers.Done()
		scanner := bufio.NewScanner(pipe)
		scanner.Buffer(make([]byte, 4096), 1024*1024)
		for scanner.Scan() {
			lines <- strings.TrimSpace(scanner.Text())
		}
	}
	go readLines(stdout)
	go readLines(stderr)
	finished := make(chan error, 1)
	go func() {
		readers.Wait()
		close(lines)
		finished <- cmd.Wait()
	}()
	for line := range lines {
		if line != "" && progress != nil {
			progress(line)
		}
	}
	if err := <-finished; err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("brew %s: %w", strings.Join(args, " "), err)
	}
	return nil
}
