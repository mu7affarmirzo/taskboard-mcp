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

func (p *TelegramPresenter) FormatHelp() string {
	return `*Telegram -> Trello Bot*

Just send me any message and I'll create a Trello card!

*Tips:*
- Include "urgent" or "high priority" to mark as important
- Add "#label" to tag your task
- Say "due Friday" or "by March 15" to set a deadline

*Commands:*
/boards - List your Trello boards
/setboard - Set default board
/setlist - Set default list
/help - Show this message`
}
