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
	Run:  runAICommand,
}

func runAICommand(cmd *cobra.Command, args []string) {
	input, executedCommand, err := getInput(args)
	if err != nil {
		color.Red("Error: %v", err)
		return
	}

	fullOutput := generateFullOutput(language, executedCommand, input, customPrompt)

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		color.Red("Error: OPENAI_API_KEY is not set")
		return
	}

	color.Cyan("Data received for analysis.")
	color.Green("Result:\n%v\n", input)
	color.Yellow("Waiting for AI response...")

	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Start()

	response, err := getAIResponse(apiKey, fullOutput)
	s.Stop()

	if err != nil {
		color.Red("Error: %v", err)
		return
	}

	color.Green("✔ AI response received!\n")
	fmt.Println(response)
}

func getInput(args []string) (string, string, error) {
	if len(args) == 0 {
		inputBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", "", fmt.Errorf("error reading from stdin: %v", err)
		}
		return string(inputBytes), "Input from pipeline", nil
	}

	shell := getShell()
	shellCmd := exec.Command(shell, "-c", strings.Join(args, " "))
	var out bytes.Buffer
	shellCmd.Stdout = &out
	err := shellCmd.Run()
	if err != nil {
		return "", "", fmt.Errorf("error executing command: %v", err)
	}
	return out.String(), strings.Join(args, " "), nil
}

func getShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
		fmt.Println("Using /bin/sh as a fallback since the SHELL environment variable is not set.")
	}
	return shell
}

func generateFullOutput(language, executedCommand, output, customPrompt string) string {
	template := getTemplate(language)
	context := getContext(customPrompt)
	return fmt.Sprintf(template, executedCommand, output, context)
}

func getTemplate(language string) string {
	templates := map[string]string{
		"en": `Command: %s
Output:
%s

AI Analysis Task:
1. Summarize the output concisely.
2. Identify any errors or warnings.
3. Suggest next steps or optimizations if applicable.
4. Explain any unusual or important aspects of the output.

Context: %s`,
		"ja": `コマンド: %s
出力:
%s

AI分析タスク:
1. 出力を簡潔に要約してください。
2. エラーや警告を特定してください。
3. 該当する場合、次のステップや最適化を提案してください。
4. 出力の異常な点や重要な側面を説明してください。

コンテキスト: %s`,
	}

	if template, ok := templates[language]; ok {
		return template
	}
	return templates["en"]
}

func getContext(customPrompt string) string {
	if customPrompt == "" {
		return "No additional context provided."
	}
	return customPrompt
}

func getAIResponse(apiKey, fullOutput string) (string, error) {
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
	if err != nil {
		return "", fmt.Errorf("ChatCompletion error: %v", err)
	}

	return resp.Choices[0].Message.Content, nil
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	language = getDefaultLanguage()
	rootCmd.Flags().StringVarP(&language, "language", "l", language, "Language for the command execution (en/ja)")
	rootCmd.Flags().StringVarP(&model, "model", "m", "gpt-3.5-turbo", "The model to be used for the OpenAI GPT (default is gpt-3.5-turbo)")
	rootCmd.Flags().StringVarP(&customPrompt, "prompt", "p", "", "Custom prompt to be sent to the AI")
}

func getDefaultLanguage() string {
	langEnv := os.Getenv("LANG")
	if len(langEnv) >= 2 && langEnv[:2] == "ja" {
		return "ja"
	}
	return "en"
}
