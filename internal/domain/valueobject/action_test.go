package valueobject

import "testing"

func TestNewAction_Valid(t *testing.T) {
	actions := []string{
		"create_task", "move_card", "update_card", "get_card",
		"list_cards", "list_lists", "list_labels", "create_list",
		"archive_card", "delete_card", "add_comment", "assign_card",
		"search_cards", "add_label", "set_due_date",
	}
	for _, s := range actions {
		a, err := NewAction(s)
		if err != nil {
			t.Errorf("NewAction(%q) returned error: %v", s, err)
		}
		if a.String() != s {
			t.Errorf("NewAction(%q).String() = %q, want %q", s, a.String(), s)
		}
	}
}

func TestNewAction_Invalid(t *testing.T) {
	_, err := NewAction("invalid_action")
	if err == nil {
		t.Error("NewAction(\"invalid_action\") expected error, got nil")
	}
}

func TestAction_NeedsCard(t *testing.T) {
	needsCard := []Action{
		ActionMoveCard, ActionUpdateCard, ActionGetCard,
		ActionArchiveCard, ActionDeleteCard, ActionAddComment,
		ActionAssignCard, ActionAddLabel, ActionSetDueDate,
	}
	for _, a := range needsCard {
		if !a.NeedsCard() {
			t.Errorf("%q.NeedsCard() = false, want true", a)
		}
	}

	noCard := []Action{
		ActionCreateTask, ActionListCards, ActionListLists,
		ActionListLabels, ActionCreateList, ActionSearchCards,
	}
	for _, a := range noCard {
		if a.NeedsCard() {
			t.Errorf("%q.NeedsCard() = true, want false", a)
		}
	}
}
