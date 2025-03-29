package git

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func GetStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// ignore code = 1 if stderr is empty -> no staged changes
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 && stderr.Len() == 0 {
			return "", nil
		}
		return "", fmt.Errorf("git diff failed: %w\n%s", err, stderr.String())
	}
	return stdout.String(), nil
}

func Commit(msg string) error {
	cmd := exec.Command("git", "commit", "-eF", "-")

	// get a pipe to stdin
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe for git commit: %w", err)
	}

	/// connect command stdout & stderr to parent process so user can interact with the editor
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// start command asynchronously
	err =  cmd.Start()
	if err != nil {
		stdinPipe.Close()
		return fmt.Errorf("failed to start git command: %w", err)
	}

	// write commit message to stdin
	writeErrChan := make(chan error, 1)
	go func() {
		_, writeErr := io.WriteString(stdinPipe, msg)
		closeErr := stdinPipe.Close()
		if writeErr != nil {
			writeErrChan <- fmt.Errorf("failed to write commit message to git stdin: %w", writeErr)
		} else if closeErr != nil {
			writeErrChan <- fmt.Errorf("failed to close git stdin pipe: %w", closeErr)
		} else {
			writeErrChan <- nil
		}
	}()

	// wait for write to finish
	writeErr := <-writeErrChan
	if writeErr != nil {
		return writeErr
	}

	// wait for the command & editor session to complete
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("git commit command failed: %w", err)
	}

	return nil
}
