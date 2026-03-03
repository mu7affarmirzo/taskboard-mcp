# MCP Tools Roadmap — Improving Development Efficiency

A catalog of MCP servers and tool enhancements to streamline development workflows with Claude Code. Organized by impact level.

---

## High Impact

### 1. GitHub MCP Server

Manage repositories, pull requests, and issues directly from Claude Code without switching to the browser.

**Available as:** Official MCP server (`@modelcontextprotocol/server-github`)

**Tools provided:**
- `create_pull_request` — Create PRs with title, body, base/head branches
- `list_pull_requests` — List open/closed/merged PRs
- `get_pull_request` — Get PR details, diff, review status
- `create_issue` — File bugs or feature requests
- `list_issues` — Browse open issues with label/assignee filters
- `add_comment` — Comment on PRs or issues
- `merge_pull_request` — Merge when checks pass
- `list_branches` — See all branches
- `get_file_contents` — Read files from any branch without checkout

**Example prompts:**
- "Create a PR from this branch to main"
- "What issues are assigned to me?"
- "Add a comment on PR #42: LGTM, ready to merge"
- "List all open PRs with the 'bug' label"

**Setup:**
```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_your_token"
      }
    }
  }
}
```

---

### 2. Trello Enhancements (Extend Existing Server)

Additional tools for the Trello MCP server already built in this project.

**New tools to add:**

| Tool | Description | Example prompt |
|------|-------------|----------------|
| `create_list` | Create a new list on the board | "Create a Sprint 5 list on Trello" |
| `archive_card` | Archive a completed card | "Archive the card we just finished" |
| `add_comment` | Add a comment/note to a card | "Add a comment to card X: deployed to staging" |
| `assign_card` | Assign/unassign members to a card | "Assign me to this card" |
| `search_cards` | Search cards by keyword across all lists | "Find all cards mentioning 'payment'" |
| `delete_card` | Permanently delete a card | "Delete the test card we created" |
| `add_label` | Add a label to a card | "Add the Bug label to card X" |
| `set_due_date` | Set or clear a card's due date | "Set due date to next Friday" |

**Priority:** `create_list`, `archive_card`, `add_comment`, and `search_cards` are the most useful additions.

---

### 3. Database MCP Server

Query your project databases directly from Claude Code. Useful for debugging, checking data integrity, and understanding schemas without switching to a database client.

**Available as:** Official MCP server (`@modelcontextprotocol/server-postgres`) or community SQLite servers

**Tools provided:**
- `query` — Run SELECT queries against the database
- `describe_table` — Get table schema (columns, types, constraints)
- `list_tables` — List all tables in the database

**Example prompts:**
- "Show me the last 10 users that signed up"
- "What's the schema of the orders table?"
- "Count how many active subscriptions we have"
- "Find the user with email john@example.com"

**Setup (PostgreSQL):**
```json
{
  "mcpServers": {
    "postgres": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-postgres"],
      "env": {
        "POSTGRES_CONNECTION_STRING": "postgresql://user:pass@localhost:5432/mydb"
      }
    }
  }
}
```

**Safety:** Configure as read-only to prevent accidental writes.

---

## Medium Impact

### 4. Docker MCP Server

Manage containers, images, and services from Claude Code. Eliminates context-switching to terminal for container operations.

**Tools to implement:**
- `list_containers` — Show running/stopped containers with status
- `container_logs` — Fetch recent logs from a container
- `restart_container` — Restart a specific service
- `container_stats` — CPU/memory usage per container
- `compose_status` — Show docker compose service status

**Example prompts:**
- "What containers are running?"
- "Show me the last 50 lines of logs from the auth service"
- "Restart the API gateway"
- "Which container is using the most memory?"

**Implementation:** Can be built as a custom MCP server (similar to the Trello one) or use community servers like `@modelcontextprotocol/server-docker`.

---

### 5. Notion MCP Server

Already connected in this workspace. Use it for documentation, specs, and knowledge management.

**Available tools:**
- `notion-search` — Search across the workspace
- `notion-fetch` — Get page/database contents
- `notion-create-pages` — Create new pages and database entries
- `notion-update-page` — Update existing pages

**Example prompts:**
- "Create a design doc for the payment integration"
- "Search Notion for the API specification"
- "Add a new entry to the sprint planning database"
- "Update the deployment checklist"

**Already configured** — no additional setup needed.

---

### 6. Slack / Telegram Notification Server

Send notifications to yourself or a team channel when tasks complete, deployments finish, or errors occur.

**Tools to implement:**
- `send_message` — Send a message to a channel or DM
- `notify_deployment` — Post deployment status to a channel
- `alert_error` — Send an error alert with details

**Example prompts:**
- "Notify #deployments: auth service v2.1 deployed to staging"
- "Send me a DM: all tests passed"
- "Alert the team: database migration completed"

**Implementation:** Custom MCP server using Slack Bot API or Telegram Bot API (the bot infrastructure already exists in this project).

---

## Quick Wins

### 7. Filesystem MCP Server

Extended file operations beyond the current working directory. Useful for cross-project work.

**Available as:** Official MCP server (`@modelcontextprotocol/server-filesystem`)

**Tools provided:**
- `read_file` / `write_file` — Read/write files in allowed directories
- `list_directory` — Browse directory contents
- `search_files` — Search for files by name pattern
- `move_file` — Move or rename files

**Setup:**
```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/Users/you/devProjects"]
    }
  }
}
```

**Use case:** Access shared configs, scripts, or documentation across the `bot` and `travel` projects simultaneously.

---

### 8. Time Tracking Server (Custom)

Track time spent on tasks with Trello card integration.

**Tools to implement:**
- `start_timer` — Start tracking time on a card
- `stop_timer` — Stop the current timer and log duration
- `time_report` — Show time spent today/this week by card
- `log_time` — Manually log time to a card

**Example prompts:**
- "Start tracking time on card 69a66e50"
- "How long have I worked today?"
- "Stop the timer and log it"
- "Show me this week's time report"

**Implementation:** Custom MCP server with local SQLite storage for time entries, linked to Trello card IDs.

---

## Recommended Setup Order

1. **Trello enhancements** — Low effort, extends what you already have
2. **GitHub MCP** — Ready-made, just configure credentials
3. **Database MCP** — Ready-made, very useful for debugging
4. **Docker MCP** — Medium effort, useful if you work with containers daily
5. **Notification server** — Build when you need automated alerts
6. **Time tracking** — Build when you want productivity metrics

---

## Adding an MCP Server

All MCP servers are configured in `.mcp.json` at the project root:

```json
{
  "mcpServers": {
    "server-name": {
      "command": "path/to/binary",
      "args": ["optional", "args"],
      "env": {
        "API_KEY": "value"
      }
    }
  }
}
```

After adding a server, restart Claude Code for it to discover the new tools.
