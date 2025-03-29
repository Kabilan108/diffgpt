package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kabilan108/diffgpt/internal/git"
	"github.com/kabilan108/diffgpt/internal/llm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	// "github.com/spf13/viper"
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

		stat, _ := os.Stdin.Stat()
		isPiped := (stat.Mode() & os.ModeCharDevice) == 0

		if isPiped {
			diffBytes, readErr := io.ReadAll(os.Stdin)
			if readErr != nil {
				return fmt.Errorf("failed to read diff from stdin: %w", readErr)
			}
			diffContent = string(diffBytes)
		} else {
			diffContent, err = git.GetStagedDiff()
			if err != nil {
				return fmt.Errorf("failed to get staged diff: %w", err)
			}
		}

		if strings.TrimSpace(diffContent) == "" {
			fmt.Fprintln(os.Stderr, "No changes to commit")
			return nil
		}

		commitMsg, err := llm.GenerateCommitMessage(
			context.Background(), client, o.model, diffContent, o.detailed,
		)
		if err != nil {
			return fmt.Errorf("failed to generate commit message: %w", err)
		}

		if err := git.Commit(commitMsg); err != nil {
			if strings.Contains(err.Error(), "exit status") {
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

	// TODO: extend this to load ICL examples from config file
	o.apiKey = viper.GetString("api_key")
	o.baseUrl = viper.GetString("base_url")
	o.model = viper.GetString("model")
}
