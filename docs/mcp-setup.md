# Trello MCP Server — Setup Guide

A stdio-based MCP (Model Context Protocol) server that gives Claude Code direct access to your Trello board. Create cards, move them between lists, and track progress — all from natural language prompts.

## Prerequisites

- Go 1.24+
- A Trello account with API access
- Claude Code CLI

## Getting Trello Credentials

### 1. API Key

Go to <https://trello.com/power-ups/admin> and copy your API key.

### 2. Token

Open the following URL in your browser, replacing `YOUR_API_KEY` with your actual key:

```
https://trello.com/1/authorize?expiration=never&scope=read,write&response_type=token&key=YOUR_API_KEY
```

Click **Allow** and copy the token.

### 3. Board ID

Open your Trello board in the browser. The board ID is the short code after `/b/` in the URL:

```
https://trello.com/b/JYD7xt8n/my-board
                     ^^^^^^^^
                     Board ID
```

### 4. List ID (Optional)

If you want a default list for new cards (so you don't have to specify `list_name` every time), get the list ID by running:

```bash
curl "https://api.trello.com/1/boards/YOUR_BOARD_ID/lists?key=YOUR_API_KEY&token=YOUR_TOKEN&fields=id,name"
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `TRELLO_API_KEY` | Yes | Trello API key |
| `TRELLO_TOKEN` | Yes | Trello user token |
| `TRELLO_BOARD_ID` | Yes | Target board ID (short code from URL) |
| `TRELLO_LIST_ID` | No | Default list ID for `create_card` when `list_name` is omitted |

## Installation

### Option A: Project-scoped (this repo only)

The `.mcp.json` file in the project root configures Claude Code to launch the server automatically.

Edit `.mcp.json` and fill in your credentials:

```json
{
  "mcpServers": {
    "trello": {
      "command": "go",
      "args": ["run", "./cmd/mcp"],
      "env": {
        "TRELLO_API_KEY": "your-api-key",
        "TRELLO_TOKEN": "your-token",
        "TRELLO_BOARD_ID": "your-board-id"
      }
    }
  }
}
```

Restart Claude Code. The server starts on first tool call.

### Option B: Global (all projects)

Build the binary once:

```bash
cd /path/to/this/repo
go build -o ~/bin/trello-mcp ./cmd/mcp
```

Add to your global Claude Code settings at `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "trello": {
      "command": "/Users/you/bin/trello-mcp",
      "env": {
        "TRELLO_API_KEY": "your-api-key",
        "TRELLO_TOKEN": "your-token",
        "TRELLO_BOARD_ID": "your-board-id"
      }
    }
  }
}
```

This makes Trello tools available in every Claude Code session, regardless of the project.

### Option C: Using .env file (development)

Store credentials in your `.env` file:

```env
TRELLO_API_KEY=your-api-key
TRELLO_TOKEN=your-token
TRELLO_BOARD_ID=your-board-id
TRELLO_LIST_ID=optional-default-list-id
```

Then reference them in `.mcp.json` or run manually:

```bash
export $(grep "^TRELLO_" .env | xargs) && go run ./cmd/mcp
```

## Verifying the Setup

Test that the server starts and registers tools:

```bash
echo '{"jsonrpc":"2.0","method":"initialize","id":1,"params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  | TRELLO_API_KEY=key TRELLO_TOKEN=token TRELLO_BOARD_ID=board \
  go run ./cmd/mcp 2>/dev/null
```

You should see a JSON response with `"serverInfo":{"name":"trello","version":"1.0.0"}`.

To list all registered tools:

```bash
printf '{"jsonrpc":"2.0","method":"initialize","id":1,"params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}\n{"jsonrpc":"2.0","method":"notifications/initialized"}\n{"jsonrpc":"2.0","method":"tools/list","id":2}\n' \
  | TRELLO_API_KEY=key TRELLO_TOKEN=token TRELLO_BOARD_ID=board \
  go run ./cmd/mcp 2>/dev/null | tail -1 | python3 -m json.tool
```

## How It Works

```
Claude Code  --(stdio/JSON-RPC)-->  cmd/mcp/main.go  --(HTTPS)-->  Trello API
```

1. Claude Code reads `.mcp.json` (or global settings) and launches the MCP server as a subprocess
2. Communication happens over stdin/stdout using JSON-RPC 2.0
3. When you ask Claude Code something like "create a Trello card for this task", it calls the appropriate MCP tool
4. The MCP server translates the request into Trello API calls and returns the result
5. Claude Code presents the result in natural language

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Server won't start | Check all 3 required env vars are set |
| Timeout errors | Network issue — the client uses a 30s timeout |
| "list not found" | List name is case-insensitive but must match exactly (including special characters) |
| Tools not showing in Claude Code | Restart Claude Code after editing `.mcp.json` |
| Permission denied | Ensure your Trello token has `read,write` scope |
