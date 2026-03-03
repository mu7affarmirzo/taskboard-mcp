package valueobject

import "fmt"

type Action string

const (
	ActionCreateTask  Action = "create_task"
	ActionMoveCard    Action = "move_card"
	ActionUpdateCard  Action = "update_card"
	ActionGetCard     Action = "get_card"
	ActionListCards   Action = "list_cards"
	ActionListLists   Action = "list_lists"
	ActionListLabels  Action = "list_labels"
	ActionCreateList  Action = "create_list"
	ActionArchiveCard Action = "archive_card"
	ActionDeleteCard  Action = "delete_card"
	ActionAddComment  Action = "add_comment"
	ActionAssignCard  Action = "assign_card"
	ActionSearchCards Action = "search_cards"
	ActionAddLabel    Action = "add_label"
	ActionSetDueDate  Action = "set_due_date"
)

var validActions = map[Action]bool{
	ActionCreateTask:  true,
	ActionMoveCard:    true,
	ActionUpdateCard:  true,
	ActionGetCard:     true,
	ActionListCards:   true,
	ActionListLists:   true,
	ActionListLabels:  true,
	ActionCreateList:  true,
	ActionArchiveCard: true,
	ActionDeleteCard:  true,
	ActionAddComment:  true,
	ActionAssignCard:  true,
	ActionSearchCards: true,
	ActionAddLabel:    true,
	ActionSetDueDate:  true,
}

func NewAction(s string) (Action, error) {
	a := Action(s)
	if !validActions[a] {
		return "", fmt.Errorf("unknown action: %s", s)
	}
	return a, nil
}

func (a Action) String() string {
	return string(a)
}

func (a Action) NeedsCard() bool {
	switch a {
	case ActionMoveCard, ActionUpdateCard, ActionGetCard,
		ActionArchiveCard, ActionDeleteCard, ActionAddComment,
		ActionAssignCard, ActionAddLabel, ActionSetDueDate:
		return true
	}
	return false
}
