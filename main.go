package main

import (
	"bufio"
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

	var command string
	command, err = llmClient.GenerateFFmpegCommand(prompt)
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

	command = parsedCommand

	if !noConfirm {
		for {
			fmt.Print("Execute this command? (y/N/e): ")

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response == "y" || response == "yes" {
				break
			} else if response == "e" || response == "edit" {
				fmt.Print("How do you want to modify the command? ")
				editPrompt, _ := reader.ReadString('\n')
				editPrompt = strings.TrimSpace(editPrompt)

				command, err = llmClient.ModifyCommand(command, editPrompt)
				if err != nil {
					fmt.Printf("Error modifying command: %v\n", err)
					continue
				}

				fmt.Printf("\nUpdated command:\n%s\n\n", command)
				continue
			} else {
				fmt.Println("Command cancelled.")
				return
			}
		}

		parsedCommand = command
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

Confirmation:
  y - execute
  e - edit (ask LLM to modify)
  n - cancel

Configuration:
  Set environment variables or create ~/.ffchat.json:
    FFCHAT_LLM_BASE_URL    - LLM API endpoint
    FFCHAT_LLM_MODEL       - Model name  
    FFCHAT_LLM_API_KEY     - API key (optional)
    FFCHAT_LLM_TEMPERATURE - Generation temperature
    FFCHAT_FFMPEG_PATH     - Custom ffmpeg path
`)
}
