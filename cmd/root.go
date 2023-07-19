/*
Package cmd is the root of all commands.
Copyright © 2023 syu.m.5151@gmail.com
*/package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "aicommand",
	Short: "Shell command result analyzer using OpenAI GPT",
	Long: `This tool allows you to send the result of a specified shell command to OpenAI GPT and get an explanation of that result.
For example, use it as follows:

$ aicommand execute --language=ja "ls -la"

The above command sends the result of the "ls -la" command to the GPT-3.5-turbo model and retrieves an explanation in Japanese.`,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
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

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.aicommand.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
