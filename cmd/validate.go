package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/usadamasa/kubectl-localmesh/internal/config"
	"github.com/usadamasa/kubectl-localmesh/internal/validate"
)

type validateOptions struct {
	configFile string
	strict     bool
}

var validateOpts = &validateOptions{}

var validateCmd = &cobra.Command{
	Use:   "validate [config-file]",
	Short: "Validate a configuration file",
	Long: `Validate the configuration file syntax and structure.

By default, runs Go-level validation (same as 'up' command).
With --strict, additionally validates against JSON Schema (detects typos, unknown fields).

Examples:
  kubectl-localmesh validate -f services.yaml
  kubectl-localmesh validate services.yaml
  kubectl-localmesh validate -f services.yaml --strict`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().StringVarP(&validateOpts.configFile, "config", "f", "", "config yaml path")
	validateCmd.Flags().BoolVar(&validateOpts.strict, "strict", false, "additionally validate against JSON Schema (detects typos, unknown fields)")
}

func runValidate(cmd *cobra.Command, args []string) error {
	// フラグが指定されていない場合、位置引数を使用
	if validateOpts.configFile == "" && len(args) > 0 {
		validateOpts.configFile = args[0]
	}

	if validateOpts.configFile == "" {
		return fmt.Errorf("config file required: use -f or provide as argument")
	}

	// Go-level validation (config.Load)
	_, err := config.Load(validateOpts.configFile)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// JSON Schema validation (optional)
	if validateOpts.strict {
		result, err := validate.ValidateSchemaFile(validateOpts.configFile)
		if err != nil {
			return fmt.Errorf("schema validation failed: %w", err)
		}
		if !result.OK() {
			cmd.PrintErrln("Schema validation errors:")
			for _, e := range result.Errors {
				cmd.PrintErrln("  - " + e)
			}
			return fmt.Errorf("schema validation failed with %d error(s)", len(result.Errors))
		}
	}

	cmd.Println("Configuration is valid.")
	return nil
}
