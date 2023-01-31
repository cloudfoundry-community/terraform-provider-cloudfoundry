package constant

// RelationshipType represents the Cloud Controller To-One resource targeted by
// a relationship.
type RelationshipType string

const (
	// RelationshipTypeApplication is a relationship with a Cloud Controller
	// application.
	RelationshipTypeApplication RelationshipType = "app"

	// RelationshipTypeSpace is a relationship with a Cloud Controller space.
	RelationshipTypeSpace RelationshipType = "space"

	// RelationshipTypeOrganization is a relationship with a Cloud Controller
	// organization.
	RelationshipTypeOrganization RelationshipType = "organization"
)
