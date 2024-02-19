/*
Package cmd : execute.go is executed by "go run main.go execute"
Copyright © 2023 NAME HERE syu.m.5151@gmail.com
*/
package cmd

import (
  "strings"
	"bytes"
	"context"
	"fmt"
	"io"
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
	Short: "Execute a shell command and send the result to OpenAI GPT, or analyze piped input",
	Long: `Use this command to execute a specified shell command and send its output to OpenAI GPT for analysis, 
or to analyze text piped from another command. For example:

echo "example text" | go run main.go execute
go run main.go execute "ls -la"`,
	Args: cobra.MaximumNArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		if language != "en" && language != "ja" {
			fmt.Println("Invalid language. Please select either 'en' for English or 'ja' for Japanese.")
			os.Exit(1)
		}
	},
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
			shellCmd := exec.Command(shell, "-c", args[0])
			executedCommand = args[0]
			var out bytes.Buffer
			shellCmd.Stdout = &out
			err = shellCmd.Run()
			if err != nil {
				color.New(color.FgRed).Printf("Error executing command: %v\n", err)
				return
			}
			input = out.String()
		}

		fullOutput := generateFullOutput(language, executedCommand, input)

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

func generateFullOutput(language, executedCommand, output string) string {
	var fullOutput string
	if language == "en" {
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

func init() {
  	// 環境変数LANGをチェックしてデフォルト言語を設定
	langEnv := os.Getenv("LANG")
	if strings.HasPrefix(langEnv, "ja") {
		language = "ja"
	} else {
		// LANGが日本語以外の場合、または設定されていない場合は英語をデフォルトにする
		language = "en"
	}
	executeCmd.Flags().StringVarP(&language, "language", "l", language , "Language for the command execution (en/ja)")
	executeCmd.Flags().StringVarP(&model, "model", "m", "gpt-3.5-turbo", "The model to be used for the OpenAI GPT (default is gpt-3.5-turbo)")
	rootCmd.AddCommand(executeCmd)
}
