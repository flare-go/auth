package enum

// ActionType is the type of action.
type ActionType string

const (
	// ActionCreate is the action to create.
	ActionCreate ActionType = "CREATE"

	// ActionRead is the action to read.
	ActionRead ActionType = "READ"

	// ActionUpdate is the action to update.
	ActionUpdate ActionType = "UPDATE"

	// ActionDelete is the action to delete.
	ActionDelete ActionType = "DELETE"

	// ActionList is the action to list.
	ActionList ActionType = "LIST"
)
