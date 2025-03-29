package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	// "github.com/spf13/viper"
)

var (
	cfgBaseUrl string
	cfgApiKey  string
	cfgModel   string
)

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
		fmt.Printf("api key:   '%v'\n", cfgApiKey)
		fmt.Printf("base url:  '%v'\n", cfgBaseUrl)
		fmt.Printf("model:     '%v'\n", cfgModel)

		// TODO:
		// 1. Initialize LLM client (using cfgAPIKey, cfgBaseURL)
		// 2. Detect input source (stdin vs. git diff --staged)
		// 3. Get diff content
		// 4. Handle empty diff
		// 5. Call LLM to generate message (using cfgModel)
		// 6. Create temp file with message
		// 7. Launch git commit -t <tempfile>
		// 8. Handle errors throughout

		// return fmt.Errorf("not implemented")
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// runs before main()
func init() {
	cobra.OnInitialize(initConfig)

	// global flags
	rootCmd.Flags().StringVarP(&cfgApiKey, "api-key", "k", "", "api key for llm provider")
	rootCmd.Flags().StringVarP(&cfgBaseUrl, "base-url", "u", "https://api.openai.com/v1", "base url for llm provider")
	rootCmd.Flags().StringVarP(&cfgApiKey, "model", "m", "gpt-4o-mini", "llm to use for generation")

	viper.BindPFlag("api_key", rootCmd.Flags().Lookup("api-key"))
	viper.BindPFlag("base_url", rootCmd.Flags().Lookup("base-url"))
	viper.BindPFlag("model", rootCmd.Flags().Lookup("model"))
}

// read config file & env variables if set
func initConfig() {
	viper.SetEnvPrefix("DIFFGPT")
	viper.AutomaticEnv()

	// TODO: extend this to load ICL examples from config file
	cfgApiKey = viper.GetString("api_key")
	cfgBaseUrl = viper.GetString("base_url")
	cfgModel = viper.GetString("model")
}
