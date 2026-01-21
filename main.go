package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"ffchat/config"
	"ffchat/ffmpeg"
	"ffchat/llm"
)

var (
	noConfirm bool
)

func main() {
	showHelp := flag.Bool("help", false, "show help")
	n := flag.Bool("y", false, "skip confirmation prompt")
	flag.Parse()

	if *showHelp || flag.NArg() == 0 {
		printHelp()
		return
	}

	noConfirm = *n
	prompt := strings.Join(flag.Args(), " ")

	config, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		fmt.Println("Please set up your configuration using environment variables or ~/.ffchat.json")
		fmt.Println("Required: FFCHAT_LLM_BASE_URL, FFCHAT_LLM_MODEL")
		os.Exit(1)
	}

	ff, err := ffmpeg.New()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	llmClient := llm.New(config.LLM.BaseURL, config.LLM.APIKey, config.LLM.Model, config.LLM.Temperature)

	fmt.Println("Thinking...")

	command, err := llmClient.GenerateFFmpegCommand(prompt)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	isFFmpeg := ff.IsFFmpegCommand(command)
	if !isFFmpeg {
		fmt.Printf("Warning: Response doesn't appear to be an ffmpeg command.\n")
		fmt.Printf("Response: %s\n", command)
		os.Exit(1)
	}

	parsedCommand, err := ff.ParseCommand(command)
	if err != nil {
		fmt.Printf("Error parsing command: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated command:\n%s\n\n", parsedCommand)

	if !noConfirm {
		fmt.Print("Execute this command? (y/N): ")

		var response string
		fmt.Scanln(&response)

		if response != "y" && response != "Y" {
			fmt.Println("Command cancelled.")
			return
		}
	}

	fmt.Println("Executing command...")

	if err := ff.Run(parsedCommand, false); err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf(`ffchat - AI-powered ffmpeg command generator

Usage:
  ffchat [options] "<prompt>"
  ffchat --help

Options:
  -y     Skip confirmation prompt
  --help Show this help

Examples:
  ffchat "convert video.mp4 to webm"
  ffchat -y "extract audio from video.mp3"
  ffchat "resize image.jpg to 800x600"

Configuration:
  Set environment variables or create ~/.ffchat.json:
    FFCHAT_LLM_BASE_URL    - LLM API endpoint
    FFCHAT_LLM_MODEL       - Model name  
    FFCHAT_LLM_API_KEY     - API key (optional)
    FFCHAT_LLM_TEMPERATURE - Generation temperature
    FFCHAT_FFMPEG_PATH     - Custom ffmpeg path
`)
}
