package gateway

import (
	"context"
	"regexp"
	"strings"

	"telegram-trello-bot/internal/domain/entity"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/pkg/timeutil"
)

type RuleParserGateway struct{}

func NewRuleParserGateway() *RuleParserGateway {
	return &RuleParserGateway{}
}

var (
	dueDateRegex  = regexp.MustCompile(`(?i)\b(?:due|by)\s+(\w+(?:\s+\d{1,2})?(?:,?\s*\d{4})?)`)
	priorityRegex = regexp.MustCompile(`(?i)\b(urgent|high\s*priority|low\s*priority)\b`)
	labelRegex    = regexp.MustCompile(`#(\w+)`)
)

func (g *RuleParserGateway) Parse(_ context.Context, rawMessage string) (*entity.Task, error) {
	var opts []entity.TaskOption

	// Extract checklist items from multi-line messages (lines starting with "- " or "* ")
	lines := strings.Split(rawMessage, "\n")
	var checklist []string
	var nonChecklistLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			checklist = append(checklist, strings.TrimSpace(trimmed[2:]))
		} else {
			nonChecklistLines = append(nonChecklistLines, line)
		}
	}
	if len(checklist) > 0 {
		opts = append(opts, entity.WithChecklist(checklist))
	}

	// Use non-checklist text for further parsing
	text := strings.Join(nonChecklistLines, " ")

	// Extract priority
	if match := priorityRegex.FindString(text); match != "" {
		match = strings.ToLower(match)
		if strings.Contains(match, "urgent") || strings.Contains(match, "high") {
			opts = append(opts, entity.WithPriority(valueobject.PriorityHigh))
		} else if strings.Contains(match, "low") {
			opts = append(opts, entity.WithPriority(valueobject.PriorityLow))
		}
	}

	// Extract labels
	labelMatches := labelRegex.FindAllStringSubmatch(text, -1)
	if len(labelMatches) > 0 {
		labels := make([]string, len(labelMatches))
		for i, m := range labelMatches {
			labels[i] = m[1]
		}
		opts = append(opts, entity.WithLabels(labels))
	}

	// Extract due date
	if match := dueDateRegex.FindStringSubmatch(text); len(match) > 1 {
		if due, err := timeutil.ParseNaturalDate(match[1]); err == nil {
			opts = append(opts, entity.WithDueDate(due))
		}
	}

	// Title: remove extracted patterns
	title := text
	title = dueDateRegex.ReplaceAllString(title, "")
	title = priorityRegex.ReplaceAllString(title, "")
	title = labelRegex.ReplaceAllString(title, "")
	title = strings.TrimRight(strings.TrimSpace(strings.Join(strings.Fields(title), " ")), " ,.")

	return entity.NewTask(title, opts...)
}

