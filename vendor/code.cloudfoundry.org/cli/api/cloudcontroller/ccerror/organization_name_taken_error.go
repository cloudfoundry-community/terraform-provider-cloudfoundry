package ccerror

// OrganizationNameTakenError is returned when an organization with the
// requested name already exists in the Cloud Controller.
type OrganizationNameTakenError struct {
	Message string
}

func (e OrganizationNameTakenError) Error() string {
	return e.Message
}
