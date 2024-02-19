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

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
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
		shell := os.Getenv("SHELL")
		if shell == "" {
			fmt.Println("Error: The SHELL environment variable is not set. Using /bin/sh as a fallback.")
			shell = "/bin/sh"
		}

		shellCmd := exec.Command(shell, "-c", args[0])

		var out, stderr bytes.Buffer
		shellCmd.Stdout = &out
		shellCmd.Stderr = &stderr
		err := shellCmd.Run()

		if err != nil {
			color.New(color.FgRed).Printf("Error executing command: %v\n", err)
			color.New(color.FgYellow).Printf("Error details: %s\n", stderr.String())
			color.New(color.FgMagenta).Println("Possible solution: Please check the command syntax and ensure all required permissions are granted.")
			return
		}

		fullOutput := generateFullOutput(language, shellCmd.String(), out.String())

		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			fmt.Println("Error: OPENAI_API_KEY is not set")
			return
		}

		color.New(color.FgCyan).Printf("Command executed: %v\n", shellCmd.String())
		color.New(color.FgGreen).Printf("Result:\n%v\n\n", out.String())
		color.New(color.FgYellow).Printf("Waiting for aicommand response... \n")

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

		color.New(color.FgGreen).Printf("✔  aicommand response! \n\n")
		fmt.Println(resp.Choices[0].Message.Content)
	},
}

func generateFullOutput(language, command, output string) string {
	var fullOutput string
	if language == "en" {
		fullOutput = fmt.Sprintf(
			"Command executed: %v\nOutput:\n%v\nWhat does this output indicate? Are there any issues or further actions required?",
			command,
			output,
		)
	} else if language == "ja" {
		fullOutput = fmt.Sprintf(
			"実行されたコマンド: %v\n出力:\n%v\nこの出力が示すものは何ですか？ 問題はありますか、またはさらなるアクションが必要ですか？",
			command,
			output,
		)
	}
	return fullOutput
}

func init() {
	executeCmd.Flags().StringVarP(&language, "language", "l", "en", "Language for the command execution (en/ja)")
	executeCmd.Flags().StringVarP(&model, "model", "m", "gpt-3.5-turbo", "The model to be used for the OpenAI GPT (default is gpt-3.5-turbo)")
	rootCmd.AddCommand(executeCmd)
}
