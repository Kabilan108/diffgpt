package cmd

import (
	"context"
	"fmt"
	"os"

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
		// placeholder for core logic
		if len(args) == 0 {
			fmt.Printf("api key:   '%v'\n", o.apiKey)
			fmt.Printf("base url:  '%v'\n", o.baseUrl)
			fmt.Printf("model:     '%v'\n", o.model)
			return nil
		}

		// TODO:
		// 1. Initialize LLM client (using cfgAPIKey, cfgBaseURL)
		client := llm.NewClient(o.apiKey, o.baseUrl)

		// 2. Detect input source (stdin vs. git diff --staged)
		// 3. Get diff content
		diffContent := `--- a/example.txt
+++ b/example.txt
@@ -1,3 +1,4 @@
 Line 1
 Line 2
 Line 3
+Added Line 4
` // Dummy diff for now
		fmt.Println("Using dummy diff content for testing:")
		fmt.Println(diffContent)
		fmt.Println("---")

		// 4. Handle empty diff
		// 5. Call LLM to generate message (using cfgModel)
		commitMsg, err := llm.GenerateCommitMessage(
			context.Background(), client, o.model, diffContent, o.detailed,
		)
		if err != nil {
			return fmt.Errorf("failed to generate commit message: %w", err)
		}

		fmt.Println("Generated Commit Message:")
		fmt.Println(commitMsg)
		fmt.Println("---")
		// 6. Launch git commit -eF -
		// 7. Handle errors throughout

		// return fmt.Errorf("not implemented")
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
	rootCmd.Flags().StringVarP(&o.apiKey, "model", "m", "gpt-4o-mini", "llm to use for generation")
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
