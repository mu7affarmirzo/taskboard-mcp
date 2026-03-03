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

Just send me any message and I'll create a Trello card!

*Tips:*
- Include "urgent" or "high priority" to mark as important
- Add "#label" to tag your task
- Use "@username" to assign board members
- Say "due Friday" or "by March 15" to set a deadline

*Commands:*
/start - Register and get Trello authorization link
/connect <token> - Connect your Trello account
/boards - List your Trello boards
/help - Show this message`
}
