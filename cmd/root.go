/*
Package cmd is the root of all commands.
Copyright © 2023 syu.m.5151@gmail.com
*/
package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

var (
	language     string
	model        string
	customPrompt string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "aicommand [command]",
	Short: "Shell command result analyzer using OpenAI GPT",
	Long: `This tool allows you to send the result of a specified shell command to OpenAI GPT and get an explanation of that result.
For example, use it as follows:

$ aicommand "ls -la" --language=ja

The above command sends the result of the "ls -la" command to the GPT-3.5-turbo model and retrieves an explanation in Japanese.

You can also pipe input to the command:

$ echo "example text" | aicommand`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var input string
		var executedCommand string
		var err error

		if len(args) == 0 {
			// Reading from stdin
			inputBytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				color.New(color.FgRed).Printf("Error reading from stdin: %v\n", err)
				return
			}
			input = string(inputBytes)
			executedCommand = "Input from pipeline"
		} else {
			// Executing the provided command
			shell := os.Getenv("SHELL")
			if shell == "" {
				shell = "/bin/sh"
				fmt.Println("Using /bin/sh as a fallback since the SHELL environment variable is not set.")
			}
			shellCmd := exec.Command(shell, "-c", strings.Join(args, " "))
			executedCommand = strings.Join(args, " ")
			var out bytes.Buffer
			shellCmd.Stdout = &out
			err = shellCmd.Run()
			if err != nil {
				color.New(color.FgRed).Printf("Error executing command: %v\n", err)
				return
			}
			input = out.String()
		}

		fullOutput := generateFullOutput(language, executedCommand, input, customPrompt)

		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			fmt.Println("Error: OPENAI_API_KEY is not set")
			return
		}

		color.New(color.FgCyan).Printf("Data received for analysis.\n")
		color.New(color.FgGreen).Printf("Result:\n%v\n\n", input)
		color.New(color.FgYellow).Printf("Waiting for AI response... \n")

		s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
		s.Start()

		client := openai.NewClient(apiKey)

		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: model,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: fullOutput,
					},
				},
			},
		)

		s.Stop()

		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			return
		}

		color.New(color.FgGreen).Printf("✔ AI response received! \n\n")
		fmt.Println(resp.Choices[0].Message.Content)
	},
}

func generateFullOutput(language, executedCommand, output, customPrompt string) string {
	var fullOutput string
	if customPrompt != "" {
		fullOutput = fmt.Sprintf(
			"%s\nExecuted command or input source: %v\nOutput:\n%v",
			customPrompt,
			executedCommand,
			output,
		)
	} else if language == "en" {
		fullOutput = fmt.Sprintf(
			"Executed command or input source: %v\nOutput:\n%v\nWhat does this output indicate? Are there any issues or further actions required?",
			executedCommand,
			output,
		)
	} else if language == "ja" {
		fullOutput = fmt.Sprintf(
			"実行されたコマンドまたは入力ソース: %v\n出力:\n%v\nこの出力が示すものは何ですか？ 問題はありますか、またはさらなるアクションが必要ですか？",
			executedCommand,
			output,
		)
	}
	return fullOutput
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
	// Set default language based on LANG environment variable
	langEnv := os.Getenv("LANG")
	if len(langEnv) >= 2 && langEnv[:2] == "ja" {
		language = "ja"
	} else {
		language = "en"
	}

	// Add flags to rootCmd
	rootCmd.Flags().StringVarP(&language, "language", "l", language, "Language for the command execution (en/ja)")
	rootCmd.Flags().StringVarP(&model, "model", "m", "gpt-3.5-turbo", "The model to be used for the OpenAI GPT (default is gpt-3.5-turbo)")
	rootCmd.Flags().StringVarP(&customPrompt, "prompt", "p", "", "Custom prompt to be sent to the AI")
}
