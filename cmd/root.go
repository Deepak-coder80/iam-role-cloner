/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "iam-role-cloner",
	Short: "A CLI tool to clone IAM roles between AWS environments",
	Long: `IAM Role Cloner is a powerful CLI tool that helps you clone IAM roles
between different AWS environments (dev, staging, prod) with pattern replacement.

Features:
- Clone roles between AWS profiles
- Replace environment patterns in role names and policies
- Batch processing of multiple roles
- Comprehensive logging and error handling

Example usage:
  iam-role-cloner clone --source-profile dev --dest-profile prod
  iam-role-cloner list --profile dev --pattern "dev_*"`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 Welcome to IAM Role Cloner!")
		fmt.Println("===============================")
		fmt.Println()
		fmt.Println("A powerful tool to clone IAM roles between AWS environments.")
		fmt.Println()
		fmt.Println("Available commands:")
		fmt.Println("  clone    Clone IAM roles between profiles")
		fmt.Println("  list     List IAM roles in a profile")
		fmt.Println("  version  Show version information")
		fmt.Println()
		fmt.Println("Use 'iam-role-cloner [command] --help' for more information about a command.")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  iam-role-cloner clone --help")
		fmt.Println("  iam-role-cloner list --profile dev")
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

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.iam-role-cloner.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolP("dry-run", "d", false, "Show what would be done without actually doing it")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
