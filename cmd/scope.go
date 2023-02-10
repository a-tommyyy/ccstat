/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/atomiyama/ccstat/pkg/ccstat"
	"github.com/spf13/cobra"
)

// scopeCmd represents the scope command
var scopeCmd = &cobra.Command{
	Use:        "scope [flags]",
	Short:      "Aggregate sum of insertion/deletion line count by commit scope",
	ArgAliases: []string{"REPO_PATH"},
	Run: func(cmd *cobra.Command, args []string) {
		res, err := ccstat.AggByScope()
		if err != nil {
			os.Exit(1)
		}
		fmt.Println(res)
	},
}

func init() {
	rootCmd.AddCommand(scopeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scopeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scopeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
