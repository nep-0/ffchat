# ffchat

AI-powered ffmpeg command generator. Describe what you want to do in natural language and ffchat will generate and run the ffmpeg command.

## Install

1. Download the latest release from [Releases](https://github.com/nep-0/ffchat/releases)
2. Rename and move it to a directory in your PATH, e.g.:

```bash
mv ffchat-linux-amd64 ffchat
chmod +x ffchat
sudo mv ffchat /usr/local/bin/
```

## Usage

```bash
ffchat "convert video.mp4 to webm"    # with confirmation
ffchat -y "convert video.mp4 to webm" # skip confirmation
```

## Configuration

Set environment variables or create `~/.ffchat.json`:

```json
{
  "llm": {
    "base_url": "https://api.openai.com/v1",
    "api_key": "",
    "model": "gpt-4",
    "temperature": 0.1
  },
  "ffmpeg": {
    "path": ""
  }
}
```

- `FFCHAT_LLM_BASE_URL` - LLM API endpoint
- `FFCHAT_LLM_MODEL` - Model name
- `FFCHAT_LLM_API_KEY` - API key
- `FFCHAT_LLM_TEMPERATURE` - Generation temperature (optional)
- `FFCHAT_FFMPEG_PATH` - Custom ffmpeg path (optional)
