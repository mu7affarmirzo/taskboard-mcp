# Trello MCP Server — Tools Reference

All available tools exposed by the Trello MCP server. Each tool can be invoked by Claude Code through natural language prompts.

---

## create_card

Create a new Trello card on the configured board.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `title` | string | Yes | Card title |
| `description` | string | No | Card description (supports markdown) |
| `list_name` | string | No | List name to create the card in. Falls back to `TRELLO_LIST_ID` env var if omitted |
| `due_date` | string | No | Due date in `YYYY-MM-DD` format |

**Returns:** `card_id`, `card_url`, `title`

**Example prompts:**
- "Create a Trello card called 'Implement user auth' in the bot | Tasks list"
- "Add a card to Trello: Fix login bug, due 2026-03-15"
- "Create a Trello card for the current task with a description of what we're doing"

**Example response:**
```json
{
  "card_id": "69a66e50e444ce8a5a00aef3",
  "card_url": "https://trello.com/c/cr7ANfvJ",
  "title": "Implement user auth"
}
```

---

## move_card

Move a Trello card to a different list. Useful for updating task status (e.g., moving from "Tasks" to "Testing" or "Done").

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `card_id` | string | Yes | Trello card ID |
| `list_name` | string | Yes | Target list name (case-insensitive) |

**Returns:** `card_id`, `moved_to`, `list_id`

**Example prompts:**
- "Move Trello card 69a66e50 to the Seller| Tasks list"
- "I'm done with this task, move the card to Done"
- "Move card cr7ANfvJ to Testing"

**Example response:**
```json
{
  "card_id": "69a66e50e444ce8a5a00aef3",
  "moved_to": "Seller| Tasks",
  "list_id": "6921c414367209006ec23250"
}
```

---

## update_card

Update one or more fields on an existing Trello card.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `card_id` | string | Yes | Trello card ID |
| `title` | string | No | New card title |
| `description` | string | No | New card description |
| `due_date` | string | No | New due date in `YYYY-MM-DD` format |

At least one optional field must be provided.

**Returns:** `card_id`, `title`, `card_url`

**Example prompts:**
- "Update the Trello card title to 'Implement OAuth2'"
- "Change the due date on card 69a66e50 to March 20th"
- "Update the card description with our implementation notes"

**Example response:**
```json
{
  "card_id": "69a66e50e444ce8a5a00aef3",
  "title": "Implement OAuth2",
  "card_url": "https://trello.com/c/cr7ANfvJ"
}
```

---

## get_card

Get full details of a Trello card including labels, members, and due date.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `card_id` | string | Yes | Trello card ID |

**Returns:** `card_id`, `title`, `description`, `card_url`, `list_id`, `due`, `labels`, `members`

**Example prompts:**
- "Show me the details of Trello card 69a66e50"
- "What's on that Trello card?"
- "Get the card info for cr7ANfvJ"

**Example response:**
```json
{
  "card_id": "69a66e50e444ce8a5a00aef3",
  "title": "MCP Server Test Card",
  "description": "Testing the MCP server integration",
  "card_url": "https://trello.com/c/cr7ANfvJ",
  "list_id": "6921c4df29524a27835674c0",
  "due": "2026-03-10",
  "labels": ["Bug"],
  "members": ["John Doe"]
}
```

---

## list_cards

List all cards in a specific list on the board.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `list_name` | string | Yes | List name (case-insensitive) |

**Returns:** Array of `{ card_id, title, card_url }`

**Example prompts:**
- "Show me all cards in the bot | Tasks list"
- "What's in the To Do list on Trello?"
- "List the Seller tasks"

**Example response:**
```json
[
  {
    "card_id": "6921c4e774f7de651be6cebe",
    "title": "sellers",
    "card_url": "https://trello.com/c/XwPH2HlR"
  },
  {
    "card_id": "69a66e50e444ce8a5a00aef3",
    "title": "MCP Server Test Card",
    "card_url": "https://trello.com/c/cr7ANfvJ"
  }
]
```

---

## list_lists

List all lists on the configured Trello board.

**Parameters:** None

**Returns:** Array of `{ list_id, name }`

**Example prompts:**
- "What lists are on the Trello board?"
- "Show me the board columns"
- "List the Trello lists"

**Example response:**
```json
[
  { "list_id": "6921c414367209006ec23250", "name": "Seller| Tasks" },
  { "list_id": "6921c4df29524a27835674c0", "name": "bot | Tasks" },
  { "list_id": "6921c5d4d0b0ba589c73d685", "name": "Customer | Tasks" },
  { "list_id": "6921c5dff222e50109e635a1", "name": "Investors | Tasks" }
]
```

---

## list_labels

List all labels on the configured Trello board.

**Parameters:** None

**Returns:** Array of `{ label_id, name, color }`

**Example prompts:**
- "What labels are available on the Trello board?"
- "Show me the board labels"

**Example response:**
```json
[
  { "label_id": "lb1", "name": "Bug", "color": "red" },
  { "label_id": "lb2", "name": "Feature", "color": "green" },
  { "label_id": "lb3", "name": "Urgent", "color": "orange" }
]
```

---

## Typical Workflow

A common development workflow using these tools:

1. **Start a task** — "Create a Trello card called 'Add payment integration' in bot | Tasks"
2. **Track progress** — "Update the card description with: implemented Stripe webhook handler"
3. **Change status** — "Move the card to Customer | Tasks"
4. **Check board** — "List all cards in bot | Tasks"

## Error Handling

All tools return errors as structured MCP error results. Common errors:

| Error | Cause |
|-------|-------|
| `list "X" not found on board` | The list name doesn't match any list on the board |
| `no list specified and TRELLO_LIST_ID not configured` | `create_card` called without `list_name` and no default list set |
| `no fields to update` | `update_card` called without any optional fields |
| `create card failed: trello API returned status 401` | Invalid API key or token |
| `context deadline exceeded` | Trello API didn't respond within 30 seconds |
