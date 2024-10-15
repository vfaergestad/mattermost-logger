# Mattermost Message Monitor

A Go-based application that connects to Mattermost's WebSocket API to monitor and log messages from specified channels. It captures `"posted"` events, extracts relevant information including channel names, and writes the data to a JSON file.

## Features

- **Real-time Monitoring:** Connects to Mattermost's WebSocket API to listen for events in real-time.
- **Event Filtering:** Processes only `"posted"` events from specified channels.
- **Structured Logging:** Structured logs with configurable log levels (`DEBUG`, `INFO`, `WARN`, `ERROR`).
- **JSON Output:** Writes captured messages to a file, including details like channel names.
- **File Rotation**: Rotate the output file on given intervals.

## Requirements

- **Go:** Version `1.16` or higher.
- **Mattermost Server:** Access to a Mattermost server with appropriate permissions.
- **Access Token:** A valid Mattermost access token with access to the monitored channels.

## Installation

### Docker

A docker image of the application is available at `ghcr.io/vfaergestad/mattermost-logger`.

Run with docker compose:

```yaml
services:
  mattermost-logger:
    image: ghcr.io/vfaergestad/mattermost-logger:latest
    container_name: mattermost-logger
    volumes:
      - ./config.json:/app/config.json
      - ./out:/app/out
```

### Build the binary

1. **Clone the Repository:**

```git clone https://github.com/vfaergestad/mattermost-logger.git```

2. **Navigate to the Project Directory:**

```cd mattermost-logger```

3. **Install Dependencies:**

```go mod tidy```

4. **Build the Application:**

```go build -o mattermost_logger cmd/main.go```

## Configuration

Create a `config.json` file in the project directory with the following structure:

```json
{
"mattermost_domain": "your-mattermost-domain.com",
"mattermost_port": 443,
"use_tls": true,
"auth_token": "your_access_token",
"channel_ids": [
  "channel_id_1",
  "channel_id_2"
],
"output_dir": "out",
"output_file_prefix": "messages",
"rotation_interval": "5m",
"insecure_skip_tls_verify": false,
"log_level": "INFO"
}
```

### Configuration Fields

- **`mattermost_domain`**: The domain of your Mattermost server.
- **`mattermost_port`**: (Optional) The port of your Mattermost server. Defaults to `443` for tls, else `80`.
- **`use_tls`**: Set to `true` if using TLS (`wss`), otherwise `false` (`ws`).
- **`auth_token`**: Your Mattermost access token.
- **`channel_ids`**: An array of channel IDs to monitor.
- `output_dir`: (Optional) The directory to save the output of the bot. Defaults to `out`.
- `output_file_prefix`: (Optional) The prefix of the files created by the bot. Defaults to `messages`.
- `rotation_interval`: (Optional) The interval that the bot will rotate its output files. Supported: `1m`, `5m`, `24h`, `168h` (1 week), `720h` (1 month). Default: `24h`
- **`insecure_skip_tls_verify`**: (Optional) Set to `true` to skip TLS certificate verification (not recommended for production).
- **`log_level`**: (Optional) Set the logging level (`DEBUG`, `INFO`, `WARN`, `ERROR`). Defaults to `INFO`.

## Logging

The application uses the `zap` library for structured logging. Logs are categorized into different levels:

- **`DEBUG`**: Detailed information, useful during development and troubleshooting.
- **`INFO`**: General operational messages.
- **`WARN`**: Indicators of potential issues.
- **`ERROR`**: Critical problems that may prevent some operations from functioning correctly.

### Log Configuration

Adjust the `log_level` in `config.json` to control the verbosity of the logs.

```"log_level": "INFO"```

Available levels:

- `DEBUG`
- `INFO`
- `WARN`
- `ERROR`

## Output

Captured messages are written to the specified `output_file` in JSON format. Each entry includes:

- **`id`**: Message ID.
- **`created_at`**: Timestamp when the message was created on Mattermost.
- **`user_id`**: ID of the user who posted the message.
- **`channel_id`**: ID of the channel where the message was posted.
- **`channel_name`**: Name of the channel.
- **`message`**: The message content.
- **`username`**: Username of the sender.
- **`processed_at`**: Timestamp when the message was processed by the application.

### Sample `messages.json` Entry

```json
{
"id": "mpno5dgjp3fhxjy7smsx4cjttc",
"created_at": "2024-09-14T12:32:04Z",
"user_id": "ww98uzzdtfy4mmca7wfy96gp1w",
"channel_id": "pjpk8d8qopg9cowtggf1bbqkknz3h",
"channel_name": "social",
"message": "heyhey",
"username": "nordmann",
"processed_at": "2024-09-14T12:32:04Z"
}
```
