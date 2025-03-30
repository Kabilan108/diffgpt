package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kabilan108/diffgpt/internal/config"
	"github.com/kabilan108/diffgpt/internal/git"
	"github.com/kabilan108/diffgpt/internal/llm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Options struct {
	baseUrl  string
	apiKey   string
	model    string
	detailed bool
}

var o = Options{}

var rootCmd = &cobra.Command{
	Use:   "diffgpt",
	Short: "Generate commit messages based on your diffs.",
	Long: `generate commit messages from diffs

uses models from open router and supports any openai-compatible llm provider.

set the following environment variables to use a different provider.
  DIFFGPT_API_KEY:   api key for an llm provider
  DIFFGPT_BASE_URL:  base url for an openai-compatible api (e.g. https://api.openai.com/v1)
  DIFFGPT_MODEL:     model to use for generation (e.g. gpt-4o, anthropic/claude-3-haiku
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if o.apiKey == "" {
			return fmt.Errorf("API Key not provided. Set DIFFGPT_API_KEY or use --api-key flag")
		}
		client := llm.NewClient(o.apiKey, o.baseUrl)

		var diffContent string
		var err error
		var repoRoot string

		stat, _ := os.Stdin.Stat()
		isPiped := (stat.Mode() & os.ModeCharDevice) == 0

		// Attempt to determine repo root regardless of input mode
		currentRepoRoot, repoErr := git.GetRepoRoot("")
		if repoErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not determine repo root: %v\n", repoErr)
			repoRoot = ""
		} else {
			repoRoot = currentRepoRoot
		}

		if isPiped {
			diffBytes, readErr := io.ReadAll(os.Stdin)
			if readErr != nil {
				return fmt.Errorf("failed to read diff from stdin: %w", readErr)
			}
			diffContent = string(diffBytes)
		} else {
			diffContent, err = git.GetStagedDiff(repoRoot)
			if err != nil {
				return fmt.Errorf("failed to get staged diff: %w", err)
			}
		}

		if strings.TrimSpace(diffContent) == "" {
			fmt.Fprintln(os.Stderr, "No changes to commit")
			return nil
		}

		examples := []config.Example{}
		cfg, loadErr := config.LoadConfig()
		if loadErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", loadErr)
		} else {
			// always load global examples if they exist
			if globalEx, ok := cfg.Examples["global"]; ok {
				examples = append(examples, globalEx...)
			}
			// load repo-specific examples
			if repoRoot != "" {
				absRepoRoot, err := filepath.Abs(repoRoot)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to get absolute path for repo root: %v\n", err)
				} else {
					if repoEx, ok := cfg.Examples[absRepoRoot]; ok {
						examples = append(examples, repoEx...)
					}
				}
			}
		}

		commitMsg, err := llm.GenerateCommitMessage(
			context.Background(), client, o.model, diffContent, o.detailed, examples,
		)
		if err != nil {
			return fmt.Errorf("failed to generate commit message: %w", err)
		}

		if err := git.Commit(commitMsg, repoRoot); err != nil {
			// Check for specific exit codes that indicate user actions rather than errors
			// Git returns 1 when commit is aborted in editor
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				fmt.Fprintln(os.Stderr, "Commit was aborted or canceled by user")
				return nil
			}
			return fmt.Errorf("failed to run git commit: %w", err)
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// diffgpt flags
	rootCmd.Flags().StringVarP(&o.apiKey, "api-key", "k", "", "api key for llm provider")
	rootCmd.Flags().StringVarP(&o.baseUrl, "base-url", "u", "https://api.openai.com/v1", "base url for llm provider")
	rootCmd.Flags().StringVarP(&o.model, "model", "m", "gpt-4o-mini", "llm to use for generation")
	rootCmd.Flags().BoolVarP(&o.detailed, "detailed", "d", false, "whether to generate a detailed commit message")

	// bind env vars to flags
	viper.BindPFlag("api_key", rootCmd.Flags().Lookup("api-key"))
	viper.BindPFlag("base_url", rootCmd.Flags().Lookup("base-url"))
	viper.BindPFlag("model", rootCmd.Flags().Lookup("model"))
}

func initConfig() {
	viper.SetEnvPrefix("DIFFGPT")
	viper.AutomaticEnv()

	o.apiKey = viper.GetString("api_key")
	o.baseUrl = viper.GetString("base_url")
	o.model = viper.GetString("model")
}