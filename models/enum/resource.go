package enum

// ResourceType is the type of resource.
type ResourceType string

const (
	// ResourceUser is the resource for user.
	ResourceUser ResourceType = "USER"

	// ResourceRole is the resource for role.
	ResourceRole ResourceType = "ROLE"

	// ResourcePermission is the resource for permission.
	ResourcePermission ResourceType = "PERMISSION"

	// ResourceProduct is the resource for product.
	ResourceProduct ResourceType = "PRODUCT"

	// ResourceOrder is the resource for order.
	ResourceOrder ResourceType = "ORDER"
)
