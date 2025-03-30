package git

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type CommitInfo struct {
	SHA     string
	Subject string
}

// execute git commmand in a specific directory or cwd if dir is empty
func runGitCommand(dir string, args ...string) (string, string, error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	stdoutStr := strings.TrimSpace(stdout.String())
	stderrStr := strings.TrimSpace(stderr.String())

	if err != nil {
		return stdoutStr, stderrStr, fmt.Errorf(
			"git command %v failed in dir '%s': %w\n%s", args, dir, err, stderrStr,
		)
	}
	return stdoutStr, stderrStr, nil
}

// find the root directory of the git repo containing a given path
func GetRepoRoot(path string) (string, error) {
	// if path is empty, use cwd
	targetDir := path
	if targetDir == "" {
		var err error
		targetDir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current working directory: %w", err)
		}
	}

	// ensure the path exists and is a directory if specified
	if path != "" {
		info, err := os.Stat(path)
		if err != nil {
			return "", fmt.Errorf("failed to stat path %s: %w", path, err)
		}
		if !info.IsDir() {
			// if its a file, use its containing directory
			targetDir = filepath.Dir(path)
		}
	}
	stdout, _, err := runGitCommand(targetDir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("failed to find git repository root for %s: %w", targetDir, err)
	}
	absPath, err := filepath.Abs(stdout)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for repo root %s: %w", stdout, err)
	}
	return absPath, nil
}

func GetCommitLog(repoPath, startRef string, count int) ([]CommitInfo, error) {
	args := []string{"log", fmt.Sprintf("--format=format:%%H %%s"), fmt.Sprintf("-n %d", count)}
	if startRef != "" {
		args = append(args, startRef)
	}

	stdout, _, err := runGitCommand(repoPath, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	lines := strings.Split(stdout, "\n")
	commits := make([]CommitInfo, 0, len(lines))
	for _, line := range lines {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			commits = append(commits, CommitInfo{SHA: parts[0], Subject: parts[1]})
		}
	}
	return commits, nil
}

// GetDiffForCommit retrieves the diff associated with a specific commit SHA.
func GetDiffForCommit(repoPath, sha string) (string, error) {
	// `git show` includes the diff. `--pretty=""` suppresses commit info.
	// `--unified=0` can minimize context lines, but let's keep default context for better LLM understanding.
	// Using `<sha>^!` shows changes *in* the commit vs its parent. Handles merge commits better than `show`.
	// Need special handling for the very first commit (no parent). `git show` works here.
	// Let's try `git show` first for simplicity.
	// stdout, _, err := runGitCommand(repoPath, "show", "--pretty=format:%b", sha) // %b = body (includes diff)
	// A potentially cleaner way: diff against parent. Handles initial commit via magic SHA.
	emptyTreeSHA := "4b825dc642cb6eb9a060e54bf8d69288fbee4904" // Git's magic empty tree hash
	parentRef := sha + "^"

	// Check if the commit has a parent
	_, _, err := runGitCommand(repoPath, "rev-parse", "--verify", parentRef)
	diffTarget := parentRef
	if err != nil {
		// Likely the initial commit, diff against the empty tree
		// Check if the error indicates no parent
		if strings.Contains(err.Error(), "unknown revision") || strings.Contains(err.Error(), "bad revision") {
			fmt.Fprintf(os.Stderr, "info: commit %s appears to be the initial commit, diffing against empty tree\n", sha[:7])
			diffTarget = emptyTreeSHA
		} else {
			// Different error, propagate it
			return "", fmt.Errorf("failed to check parent for commit %s: %w", sha, err)
		}
	}

	// Get the diff between the commit and its determined parent/empty tree
	stdout, _, err := runGitCommand(repoPath, "diff", diffTarget, sha)
	if err != nil {
		// Fallback or error? Maybe try `git show` if `diff` fails? For now, error out.
		return "", fmt.Errorf("failed to get diff for commit %s: %w", sha, err)
	}

	return stdout, nil
}

func GetCommitMessage(repoPath, sha string) (string, error) {
	stdout, _, err := runGitCommand(repoPath, "log", "-n", "1", "--pretty=format:%B", sha)
	if err != nil {
		return "", fmt.Errorf("failed to get commit message for %s: %w", sha, err)
	}
	return stdout, nil
}

// GetStagedDiff (existing function - no changes needed for now)
func GetStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Check if it's just "no changes staged"
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 0 && stdout.Len() == 0 && stderr.Len() == 0 {
			return "", nil // No staged changes is not an error for us
		}
		// Check if stderr indicates no changes (less reliable)
		if stderr.Len() == 0 && stdout.Len() == 0 {
			return "", nil // Assume no changes if output is empty
		}
		// Otherwise, it's a real error
		return "", fmt.Errorf("git diff --staged failed: %w\nstderr: %s", err, stderr.String())
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
	err = cmd.Start()
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
