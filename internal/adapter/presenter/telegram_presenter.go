package presenter

import (
	"fmt"
	"strings"

	"telegram-trello-bot/internal/usecase/dto"
)

type TelegramPresenter struct{}

func NewTelegramPresenter() *TelegramPresenter {
	return &TelegramPresenter{}
}

func (p *TelegramPresenter) FormatTaskCreated(output *dto.CreateTaskOutput) string {
	var sb strings.Builder
	sb.WriteString("*Task Created!*\n\n")
	sb.WriteString(fmt.Sprintf("*Title:* %s\n", output.TaskTitle))
	sb.WriteString(fmt.Sprintf("*Priority:* %s\n", output.Priority))
	if output.DueDate != nil {
		sb.WriteString(fmt.Sprintf("*Due:* %s\n", output.DueDate.Format("Jan 2, 2006")))
	}
	if len(output.Labels) > 0 {
		sb.WriteString(fmt.Sprintf("*Labels:* %s\n", strings.Join(output.Labels, ", ")))
	}
	if len(output.Members) > 0 {
		sb.WriteString(fmt.Sprintf("*Assigned to:* %s\n", strings.Join(output.Members, ", ")))
	}
	sb.WriteString(fmt.Sprintf("\n%s", output.CardURL))
	return sb.String()
}

func (p *TelegramPresenter) FormatTaskPreview(output *dto.ParseTaskOutput) string {
	var sb strings.Builder
	sb.WriteString("*Create this task?*\n\n")
	sb.WriteString(fmt.Sprintf("*Title:* %s\n", output.TaskTitle))
	sb.WriteString(fmt.Sprintf("*Priority:* %s\n", output.Priority))
	if output.DueDate != nil {
		sb.WriteString(fmt.Sprintf("*Due:* %s\n", output.DueDate.Format("Jan 2, 2006")))
	}
	if len(output.Labels) > 0 {
		sb.WriteString(fmt.Sprintf("*Labels:* %s\n", strings.Join(output.Labels, ", ")))
	}
	if len(output.Members) > 0 {
		sb.WriteString(fmt.Sprintf("*Assigned to:* %s\n", strings.Join(output.Members, ", ")))
	}
	if len(output.Checklist) > 0 {
		sb.WriteString("*Checklist:*\n")
		for _, item := range output.Checklist {
			sb.WriteString(fmt.Sprintf("  - %s\n", item))
		}
	}
	return sb.String()
}

func (p *TelegramPresenter) FormatIntentPreview(output *dto.IntentOutput) string {
	var sb strings.Builder
	sb.WriteString("*Create this task?*\n\n")
	sb.WriteString(fmt.Sprintf("*Title:* %s\n", output.Title))
	if output.Priority != "" {
		sb.WriteString(fmt.Sprintf("*Priority:* %s\n", output.Priority))
	}
	if output.Description != "" {
		sb.WriteString(fmt.Sprintf("*Description:* %s\n", output.Description))
	}
	if output.DueDate != nil {
		sb.WriteString(fmt.Sprintf("*Due:* %s\n", output.DueDate.Format("Jan 2, 2006")))
	}
	if len(output.Labels) > 0 {
		sb.WriteString(fmt.Sprintf("*Labels:* %s\n", strings.Join(output.Labels, ", ")))
	}
	if len(output.Members) > 0 {
		sb.WriteString(fmt.Sprintf("*Assigned to:* %s\n", strings.Join(output.Members, ", ")))
	}
	if len(output.Checklist) > 0 {
		sb.WriteString("*Checklist:*\n")
		for _, item := range output.Checklist {
			sb.WriteString(fmt.Sprintf("  - %s\n", item))
		}
	}
	return sb.String()
}

func (p *TelegramPresenter) FormatActionResult(output *dto.ActionOutput) string {
	switch output.Action {
	case "get_card":
		return p.FormatCardDetails(output)
	case "list_cards", "search_cards":
		return p.FormatCardList(output)
	case "list_lists":
		return p.FormatListOfLists(output)
	case "list_labels":
		return p.FormatLabelList(output)
	default:
		return p.formatGenericResult(output)
	}
}

func (p *TelegramPresenter) formatGenericResult(output *dto.ActionOutput) string {
	msg := output.Message
	if output.CardURL != "" {
		msg += "\n" + output.CardURL
	}
	return msg
}

func (p *TelegramPresenter) FormatCardDetails(output *dto.ActionOutput) string {
	if len(output.Items) == 0 {
		return output.Message
	}
	item := output.Items[0]
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*%s*\n", item.Name))
	if item.Extra != "" {
		sb.WriteString(fmt.Sprintf("\n%s\n", item.Extra))
	}
	if item.URL != "" {
		sb.WriteString(fmt.Sprintf("\n%s", item.URL))
	}
	return sb.String()
}

func (p *TelegramPresenter) FormatCardList(output *dto.ActionOutput) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*%s*\n\n", output.Message))
	if len(output.Items) == 0 {
		sb.WriteString("No cards found.")
		return sb.String()
	}
	for i, item := range output.Items {
		sb.WriteString(fmt.Sprintf("%d. %s", i+1, item.Name))
		if item.URL != "" {
			sb.WriteString(fmt.Sprintf(" - [link](%s)", item.URL))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (p *TelegramPresenter) FormatListOfLists(output *dto.ActionOutput) string {
	var sb strings.Builder
	sb.WriteString("*Board Lists:*\n\n")
	for i, item := range output.Items {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.Name))
	}
	return sb.String()
}

func (p *TelegramPresenter) FormatLabelList(output *dto.ActionOutput) string {
	var sb strings.Builder
	sb.WriteString("*Board Labels:*\n\n")
	for _, item := range output.Items {
		if item.Extra != "" {
			sb.WriteString(fmt.Sprintf("- %s (%s)\n", item.Name, item.Extra))
		} else {
			sb.WriteString(fmt.Sprintf("- %s\n", item.Name))
		}
	}
	return sb.String()
}

func (p *TelegramPresenter) FormatBoardList(output *dto.ListBoardsOutput) string {
	var sb strings.Builder
	sb.WriteString("*Your Trello Boards:*\n\n")
	for i, b := range output.Boards {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, b.Name))
	}
	sb.WriteString("\nTap a board to set it as default.")
	return sb.String()
}

func (p *TelegramPresenter) FormatBoardSelected(boardID string) string {
	return fmt.Sprintf("Board set to *%s*. Now select a list:", boardID)
}

func (p *TelegramPresenter) FormatListSelected(listID string) string {
	return fmt.Sprintf("List set to *%s*. You're all set! Send me a message to create a task.", listID)
}

func (p *TelegramPresenter) FormatError(err error) string {
	return fmt.Sprintf("Something went wrong: %s", err.Error())
}

func (p *TelegramPresenter) FormatWelcome(output *dto.RegisterUserOutput) string {
	var sb strings.Builder
	if output.IsNewUser {
		sb.WriteString("*Welcome!*\n\n")
	} else {
		sb.WriteString("*Welcome back!*\n\n")
	}
	sb.WriteString("To get started, connect your Trello account:\n\n")
	sb.WriteString(fmt.Sprintf("1. [Click here to authorize](%s)\n", output.TrelloAuthURL))
	sb.WriteString("2. Copy the token from the page\n")
	sb.WriteString("3. Send /connect <your\\_token>\n\n")
	sb.WriteString("Then use /boards to select your board and list.")
	return sb.String()
}

func (p *TelegramPresenter) FormatTrelloConnected() string {
	return "Trello account connected! Now use /boards to select your default board and list."
}

func (p *TelegramPresenter) FormatHelp() string {
	return `*Telegram -> Trello Bot*

Send me any message and I'll figure out what you want to do!

*Create tasks:*
- "create a task: implement user auth, due Friday"
- "add task: fix login bug, high priority"

*Manage cards:*
- "move the login card to Done"
- "archive the old task"
- "delete the test card"
- "add a comment to payment card: API keys updated"

*Search & browse:*
- "what cards are in Testing?"
- "search for auth"
- "what lists are on the board?"
- "show labels"

*Assign & update:*
- "assign john to the auth card"
- "set due date on login card to Friday"
- "add label Bug to payment card"

*Commands:*
/start - Register and get Trello authorization link
/connect <token> - Connect your Trello account
/boards - List your Trello boards
/help - Show this message`
}
