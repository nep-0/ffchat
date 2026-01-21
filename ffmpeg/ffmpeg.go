package ffmpeg

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type FFmpeg struct {
	Path string
}

func New() (*FFmpeg, error) {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, fmt.Errorf("ffmpeg is not installed or not in PATH: %w", err)
	}

	return &FFmpeg{Path: path}, nil
}

func (f *FFmpeg) Version() (string, error) {
	cmd := exec.Command(f.Path, "-version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get ffmpeg version: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return lines[0], nil
	}

	return "unknown", nil
}

func (f *FFmpeg) IsFFmpegCommand(content string) bool {
	content = strings.TrimSpace(content)

	patterns := []string{
		`^ffmpeg\s`,
		`^"/.*ffmpeg.*"\s`,
		"`ffmpeg[^`]*`",
	}

	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, content)
		if err == nil && matched {
			return true
		}
	}

	if strings.Contains(content, "ffmpeg") && strings.Contains(content, "-i ") {
		return true
	}

	return false
}

func (f *FFmpeg) ParseCommand(content string) (string, error) {
	content = strings.TrimSpace(content)

	var command string

	if strings.HasPrefix(content, "```") {
		lines := strings.SplitSeq(content, "\n")
		for line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "```") {
				continue
			}
			if strings.HasPrefix(line, "ffmpeg") {
				command = line
				break
			}
		}
	} else if strings.HasPrefix(content, "`") && strings.HasSuffix(content, "`") {
		command = strings.Trim(content, "`")
	} else if strings.HasPrefix(content, "\"") && strings.HasSuffix(content, "\"") {
		command = strings.Trim(content, "\"")
	} else {
		lines := strings.SplitSeq(content, "\n")
		for line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "ffmpeg") {
				command = line
				break
			}
		}
	}

	if command == "" {
		return "", fmt.Errorf("no ffmpeg command found in response")
	}

	command = strings.TrimSpace(command)

	return command, nil
}

func (f *FFmpeg) Run(command string, confirm bool) error {
	if confirm {
		fmt.Printf("\nGenerated command:\n%s\n\n", command)
		fmt.Print("Execute this command? (y/N): ")

		var response string
		fmt.Scanln(&response)

		if response != "y" && response != "Y" {
			fmt.Println("Command cancelled.")
			return nil
		}
	}

	fmt.Println("Executing command...")

	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %w\nstderr: %s", err, stderr.String())
	}

	fmt.Println("Command executed successfully.")
	return nil
}

func (f *FFmpeg) GetPath() string {
	return f.Path
}
