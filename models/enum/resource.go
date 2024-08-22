package enum

type ResourceType string

const (
	ResourceUser       ResourceType = "USER"
	ResourceRole       ResourceType = "ROLE"
	ResourcePermission ResourceType = "PERMISSION"
	ResourceProduct    ResourceType = "PRODUCT"
	ResourceOrder      ResourceType = "ORDER"
)
