package managers

// Config -
type Config struct {
	Endpoint          string
	User              string
	Password          string
	SSOPasscode       string
	CFClientID        string
	CFClientSecret    string
	UaaClientID       string
	UaaClientSecret   string
	SkipSslValidation bool
	AppLogsMax        int
	PurgeWhenDelete   bool
	DefaultQuotaName  string
}
