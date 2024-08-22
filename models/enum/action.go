package enum

type ActionType string

const (
	ActionCreate ActionType = "CREATE"
	ActionRead   ActionType = "READ"
	ActionUpdate ActionType = "UPDATE"
	ActionDelete ActionType = "DELETE"
	ActionList   ActionType = "LIST"
)
