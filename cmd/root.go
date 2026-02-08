package cmd

import (
	"github.com/spf13/cobra"
	"github.com/usadamasa/kubectl-localmesh/internal/version"
)

var globalLogLevel string

var rootCmd = &cobra.Command{
	Use:   "kubectl-localmesh",
	Short: "Local-only pseudo service mesh built on kubectl port-forward",
	Long: `kubectl-localmesh provides an ingress/gateway-like experience
for local development without installing anything into your cluster.

Built on kubectl port-forward, it runs a local Envoy proxy for host-based routing.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&globalLogLevel,
		"log-level",
		"info",
		"log level: debug|info|warn",
	)
}

func SetVersion(info version.Info) {
	rootCmd.Version = info.Version
	rootCmd.SetVersionTemplate("{{.Name}} version " + info.DisplayString() + "\n")
}

func Execute() error {
	return rootCmd.Execute()
}
