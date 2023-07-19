/*
Package cmd : execute.go is executed by "go run main.go execute"
Copyright © 2023 NAME HERE syu.m.5151@gmail.com
*/

package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

var language string
var model string

var executeCmd = &cobra.Command{
	Use:   "execute",
	Short: "Execute a shell command and send the result to OpenAI GPT",
	Args:  cobra.MinimumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		if language != "en" && language != "ja" {
			fmt.Println("Invalid language. Please select either 'en' for English or 'ja' for Japanese.")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Execute the command
		shellCmd := exec.Command(args[0], args[1:]...)
		var out bytes.Buffer
		shellCmd.Stdout = &out
		err := shellCmd.Run()
		if err != nil {
			fmt.Println("Error executing command:", err)
			return
		}

		// Concatenate the executed command and its result
		fullOutput := ""
		if language == "en" {
			fullOutput = fmt.Sprintf("Command executed: %v\nResult:\n%v\nCan you explain this result?", shellCmd.String(), out.String())
		} else if language == "ja" {
			fullOutput = fmt.Sprintf("実行したコマンド: %v\n結果:\n%v\nこの結果について説明していただけますか？", shellCmd.String(), out.String())
		}

		// Fetch the API key
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			fmt.Println("Error: OPENAI_API_KEY is not set")
			return
		}

		// Show the executed command and its result
		color.New(color.FgCyan).Printf("Command executed: %v\n", shellCmd.String())
		color.New(color.FgGreen).Printf("Result:\n%v\n\n", out.String())

		// Create a spinner
		s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
		s.Start()

		// Create a client for OpenAI
		client := openai.NewClient(apiKey)

		// Create a request for ChatGPT
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

		// Show the response
		fmt.Println(resp.Choices[0].Message.Content)
	},
}

func init() {
	executeCmd.Flags().StringVarP(&language, "language", "l", "en", "Language for the command execution (en/ja)")
	executeCmd.Flags().StringVarP(&model, "model", "m", "gpt-3.5-turbo", "The model to be used for the OpenAI GPT (default is gpt-3.5-turbo)")
	rootCmd.AddCommand(executeCmd)
}

