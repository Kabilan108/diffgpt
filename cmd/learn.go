// cmd/learn.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kabilan108/diffgpt/internal/config"
	"github.com/kabilan108/diffgpt/internal/git"
	"github.com/spf13/cobra"
)

var (
	learnGlobal bool
	learnStart  string
	learnClear  bool
	learnCount  int
)

var learnCmd = &cobra.Command{
	Use:   "learn [path]",
	Short: "learn commit style from a repository",
	Long: `scans a git repository's history to learn its commit message style.

fetches recent commits and their associated diffs and stores them as examples.

examples can be stored globally or per-repository and are used for in-context learning
during commit message generation.

if [repo-path] is omitted, learns from the current repository.`,
	Args: cobra.MaximumNArgs(1), // 0 or 1 argument for repo path
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Determine repository path
		repoPathArg := ""
		if len(args) > 0 {
			repoPathArg = args[0]
		}
		repoRoot, err := git.GetRepoRoot(repoPathArg)
		if err != nil {
			return fmt.Errorf("failed to determine repository root: %w", err)
		}
		absRepoRoot, err := filepath.Abs(repoRoot) // Use absolute path for consistency
		if err != nil {
			return fmt.Errorf("failed to get absolute path for repository root: %w", err)
		}

		// Determine storage key
		storageKey := absRepoRoot
		if learnGlobal {
			storageKey = "global"
		}

		// Handle --clear flag
		if learnClear {
			if _, exists := cfg.Examples[storageKey]; exists {
				delete(cfg.Examples, storageKey)
				if err := config.SaveConfig(cfg); err != nil {
					return fmt.Errorf("failed to save cleared configuration: %w", err)
				}
				fmt.Printf("Cleared examples for '%s'\n", storageKey)
			} else {
				fmt.Printf("No examples found for '%s' to clear.\n", storageKey)
			}
			return nil // Done after clearing
		}

		// Fetch commit log
		fmt.Printf("Fetching last %d commits from '%s'...\n", learnCount, repoRoot)
		commits, err := git.GetCommitLog(repoRoot, learnStart, learnCount)
		if err != nil {
			return fmt.Errorf("failed to fetch commit log: %w", err)
		}
		if len(commits) == 0 {
			fmt.Println("No commits found matching criteria.")
			return nil
		}

		// Fetch diffs and full messages
		fmt.Println("Processing commits to extract diffs and messages...")
		learnedExamples := make([]config.Example, 0, len(commits))
		for i, commit := range commits {
			fmt.Printf("  [%d/%d] Processing commit %s (%s)\n", i+1, len(commits), commit.SHA[:7], commit.Subject)
			diff, err := git.GetDiffForCommit(repoRoot, commit.SHA)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to get diff for commit %s: %v\n", commit.SHA, err)
				continue // Skip this commit if diff fails
			}
			// Skip empty diffs (e.g., merge commits without changes)
			if strings.TrimSpace(diff) == "" {
				fmt.Printf("  Skipping commit %s: empty diff\n", commit.SHA[:7])
				continue
			}

			fullMessage, err := git.GetCommitMessage(repoRoot, commit.SHA)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to get message for commit %s: %v\n", commit.SHA, err)
				continue // Skip this commit if message fails
			}

			learnedExamples = append(learnedExamples, config.Example{
				Diff:    diff,
				Message: fullMessage,
			})
		}

		// Store examples
		cfg.Examples[storageKey] = learnedExamples
		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save learned examples: %w", err)
		}

		fmt.Printf("Successfully learned and saved %d examples for '%s'.\n", len(learnedExamples), storageKey)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(learnCmd)

	learnCmd.Flags().BoolVarP(&learnGlobal, "global", "g", false, "Store examples globally instead of per-repository")
	learnCmd.Flags().StringVarP(&learnStart, "start", "s", "", "Commit SHA or ref to start learning from (newest commit)")
	learnCmd.Flags().BoolVarP(&learnClear, "clear", "c", false, "Clear existing examples for the target (repo or global)")
	learnCmd.Flags().IntVarP(&learnCount, "count", "n", 10, "Number of recent commits to learn from")
}