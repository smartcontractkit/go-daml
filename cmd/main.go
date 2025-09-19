package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	dar    string
	output string
	debug  bool
	pkg    string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "godaml",
		Short: "Go DAML codegen tool",
		Long:  "A command-line interface tool for interacting with DAML ledgers",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("DAR file: %s\n", dar)
			fmt.Printf("Output: %s\n", output)
			fmt.Printf("Debug mode: %t\n", debug)
		},
	}

	rootCmd.Flags().StringVar(&dar, "dar", "", "path to the DAR file")
	rootCmd.Flags().StringVar(&output, "output", "", "output dir")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "enable debug mode")
	rootCmd.Flags().StringVar(&pkg, "go_package", "", "package name")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
