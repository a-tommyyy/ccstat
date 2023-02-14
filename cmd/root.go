/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

	"github.com/atomiyama/ccstat/pkg/ccstat"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ccstat",
	Short: "ccstat - git conventional commit analyzer",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		// Build revision range struct
		after, _ := cmd.Flags().GetString("after")
		before, _ := cmd.Flags().GetString("before")
		follow, _ := cmd.Flags().GetString("follow")
		rev := &ccstat.Options{After: after, Before: before, FollowPath: follow}

		// run
		groupBy, err := cmd.Flags().GetString("group-by")
		if err != nil {
			os.Exit(1)
		}
		ccs := ccstat.New(nil)
		switch groupBy {
		case "scope":
			err = ccs.AggByScope(rev)
		case "type":
			panic("Not implemented yet")
			//ccs.AggByType()
		}
		if err != nil {
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ccstat.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().StringP("after", "A", "", "Show commits more recent than a specific date")
	rootCmd.Flags().StringP("before", "B", "", "Show commits older than a specific date")
	rootCmd.Flags().StringP("group-by", "g", "scope", "Aggregate commits group by spicific segment; Must be one of 'scope' and 'type'")
	rootCmd.Flags().StringP("follow", "f", "", "Continue listing the history of a file beyond renames (works only for a single file).")
}
