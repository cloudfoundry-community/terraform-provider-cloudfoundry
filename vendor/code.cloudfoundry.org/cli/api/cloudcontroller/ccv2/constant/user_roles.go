package constant

type UserRole int

const (
	OrgManager UserRole = iota
	BillingManager
	OrgAuditor
	OrgUser
	SpaceManager
	SpaceDeveloper
	SpaceAuditor
)
