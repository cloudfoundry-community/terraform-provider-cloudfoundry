package resources

// Revision : application revision
type Revision struct {
	GUID        string  `json:"guid"`
	Version     int     `json:"version"`
	Deployable  bool    `json:"deployable"`
	Description string  `json:"description"`
	Droplet     Droplet `json:"droplet"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}
