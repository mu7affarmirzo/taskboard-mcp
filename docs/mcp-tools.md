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

## create_list

Create a new list on the configured Trello board.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | List name |

**Returns:** `list_id`, `name`

**Example prompts:**
- "Create a Trello list called Sprint 5"
- "Add a new Testing list to the board"

**Example response:**
```json
{
  "list_id": "6921c5dff222e50109e635a1",
  "name": "Sprint 5"
}
```

---

## archive_card

Archive a Trello card (soft delete — card can be restored from the archive).

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `card_id` | string | Yes | Trello card ID |

**Returns:** `card_id`, `status`

**Example prompts:**
- "Archive the Trello card we just finished"
- "Archive card 69a66e50"

**Example response:**
```json
{
  "card_id": "69a66e50e444ce8a5a00aef3",
  "status": "archived"
}
```

---

## delete_card

Permanently delete a Trello card. This cannot be undone.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `card_id` | string | Yes | Trello card ID |

**Returns:** `card_id`, `status`

**Example prompts:**
- "Delete the test card we created"
- "Permanently remove card 69a66e50"

**Example response:**
```json
{
  "card_id": "69a66e50e444ce8a5a00aef3",
  "status": "deleted"
}
```

---

## add_comment

Add a comment to a Trello card. Useful for logging progress, deployment notes, or status updates.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `card_id` | string | Yes | Trello card ID |
| `text` | string | Yes | Comment text |

**Returns:** `comment_id`, `card_id`, `text`

**Example prompts:**
- "Add a comment to card 69a66e50: deployed to staging"
- "Comment on the card: PR #42 merged"
- "Log a note on the card: waiting for design review"

**Example response:**
```json
{
  "comment_id": "69b12345abcdef",
  "card_id": "69a66e50e444ce8a5a00aef3",
  "text": "deployed to staging"
}
```

---

## assign_card

Assign members to a Trello card by username or full name. Replaces the current member list.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `card_id` | string | Yes | Trello card ID |
| `member_names` | string | Yes | Comma-separated member usernames or full names |

**Returns:** `card_id`, `assigned_count`, `member_ids`

**Example prompts:**
- "Assign john to card 69a66e50"
- "Assign john, jane to the card"
- "Assign me to this Trello card"

**Example response:**
```json
{
  "card_id": "69a66e50e444ce8a5a00aef3",
  "assigned_count": 2,
  "member_ids": ["m1", "m2"]
}
```

---

## search_cards

Search for cards on the board by keyword. Returns up to 20 matching cards.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `query` | string | Yes | Search query |

**Returns:** Array of `{ card_id, title, card_url, list_id }`

**Example prompts:**
- "Search Trello for cards about payment"
- "Find all cards mentioning authentication"
- "Search the board for bug fixes"

**Example response:**
```json
[
  {
    "card_id": "c1",
    "title": "Payment integration",
    "card_url": "https://trello.com/c/abc",
    "list_id": "list-1"
  }
]
```

---

## add_label

Add a label to a Trello card by label name. Preserves existing labels.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `card_id` | string | Yes | Trello card ID |
| `label_name` | string | Yes | Label name (case-insensitive) |

**Returns:** `card_id`, `label_name`, `label_id`, `status`

**Example prompts:**
- "Add the Bug label to card 69a66e50"
- "Label this card as Feature"
- "Tag the card with Urgent"

**Example response:**
```json
{
  "card_id": "69a66e50e444ce8a5a00aef3",
  "label_name": "Bug",
  "label_id": "lb2",
  "status": "added"
}
```

---

## set_due_date

Set or clear the due date on a Trello card.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `card_id` | string | Yes | Trello card ID |
| `due_date` | string | No | Due date in `YYYY-MM-DD` format. Omit or leave empty to clear the due date |

**Returns:** `card_id`, `status`

**Example prompts:**
- "Set the due date on card 69a66e50 to March 15th"
- "Clear the due date on this card"
- "This card is due next Friday"

**Example response:**
```json
{
  "card_id": "69a66e50e444ce8a5a00aef3",
  "status": "set to 2026-03-15"
}
```

---

## Typical Workflow

A common development workflow using these tools:

1. **Start a task** — "Create a Trello card called 'Add payment integration' in bot | Tasks"
2. **Add labels** — "Add the Feature label to the card"
3. **Assign team** — "Assign john to the card"
4. **Track progress** — "Add a comment: implemented Stripe webhook handler"
5. **Change status** — "Move the card to Customer | Tasks"
6. **Check board** — "List all cards in bot | Tasks"
7. **Search** — "Search Trello for payment cards"
8. **Clean up** — "Archive the card we just finished"

## Error Handling

All tools return errors as structured MCP error results. Common errors:

| Error | Cause |
|-------|-------|
| `list "X" not found on board` | The list name doesn't match any list on the board |
| `no list specified and TRELLO_LIST_ID not configured` | `create_card` called without `list_name` and no default list set |
| `no fields to update` | `update_card` called without any optional fields |
| `label "X" not found on board` | The label name doesn't match any label on the board |
| `label already assigned to card` | The label is already on the card |
| `no matching members found` | No board members matched the given names |
| `create card failed: trello API returned status 401` | Invalid API key or token |
| `context deadline exceeded` | Trello API didn't respond within 30 seconds |
