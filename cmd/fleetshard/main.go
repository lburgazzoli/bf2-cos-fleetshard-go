package main

import (
	"flag"
	"github.com/spf13/cobra"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/cmd/fleetshard/run"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/logger"
	"k8s.io/klog/v2"
	"os"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "fleetshard",
		Short: "fleetshard",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	rootCmd.AddCommand(run.NewRunCmd())

	fs := flag.NewFlagSet("", flag.PanicOnError)

	klog.InitFlags(fs)
	logger.Options.BindFlags(fs)

	rootCmd.PersistentFlags().AddGoFlagSet(fs)

	if err := rootCmd.Execute(); err != nil {
		klog.ErrorS(err, "problem running command")
		os.Exit(1)
	}
}
