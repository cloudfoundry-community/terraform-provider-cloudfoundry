package managers

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	ccv2cons "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	ccWrapper "code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/api/router"
	routerWrapper "code.cloudfoundry.org/cli/api/router/wrapper"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/constant"
	uaaWrapper "code.cloudfoundry.org/cli/api/uaa/wrapper"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/configv3"
	"crypto/tls"
	"fmt"
	"github.com/hashicorp/go-uuid"
	"net"
	"net/http"
	"time"
)

// Session - wraps the available clients from CF cli
type Session struct {
	ClientV2  *ccv2.Client
	ClientV3  *ccv3.Client
	ClientUAA *uaa.Client

	// Used for direct endpoint calls
	RawClient *RawClient

	// http client used for normal request
	HttpClient *http.Client

	// To call tcp routing with this router
	RouterClient *router.Client

	// Manage upload bits like app and buildpack in full stream
	BitsManager *BitsManager

	// NOAAClient permit to access to apps logs
	NOAAClient *NOAAClient

	uaaDefaultCfGroups map[string]uaa.Group

	defaultQuotaGuid string
}

// NewSession -
func NewSession(c Config) (s *Session, err error) {
	if c.User == "" && c.CFClientID == "" {
		return nil, fmt.Errorf("Couple of user/password or uaa_client_id/uaa_client_secret must be set")
	}
	if c.User != "" && c.CFClientID == "" {
		c.CFClientID = "cf"
		c.CFClientSecret = ""
	}
	if c.Password == "" && c.CFClientID != "cf" && c.CFClientSecret != "" {
		c.User = ""
	}
	s = &Session{
		uaaDefaultCfGroups: make(map[string]uaa.Group),
	}
	config := &configv3.Config{
		ConfigFile: configv3.JSONConfig{
			ConfigVersion:        3,
			Target:               c.Endpoint,
			UAAOAuthClient:       c.CFClientID,
			UAAOAuthClientSecret: c.CFClientSecret,
			SkipSSLValidation:    c.SkipSslValidation,
		},
		ENV: configv3.EnvOverride{
			CFUsername: c.User,
			CFPassword: c.Password,
			BinaryName: "terraform-provider",
		},
	}

	configUaa := &configv3.Config{
		ConfigFile: configv3.JSONConfig{
			ConfigVersion:        3,
			UAAOAuthClient:       c.UaaClientID,
			UAAOAuthClientSecret: c.UaaClientSecret,
			SkipSSLValidation:    c.SkipSslValidation,
		},
	}

	err = s.init(config, configUaa, c)
	if err != nil {
		return nil, fmt.Errorf("Error when creating clients: %s", err.Error())
	}
	s.BitsManager = NewBitsManager(s)

	err = s.loadUaaDefaultCfGroups()
	if err != nil {
		return nil, fmt.Errorf("Error when loading uaa groups: %s", err.Error())
	}

	err = s.loadDefaultQuotaGuid()
	if err != nil {
		return nil, fmt.Errorf("Error when loading default quota: %s", err.Error())
	}
	return s, nil
}

func (s *Session) init(config *configv3.Config, configUaa *configv3.Config, configSess Config) error {
	// -------------------------
	// Create v3 and v2 clients
	ccWrappersV2 := []ccv2.ConnectionWrapper{}
	ccWrappersV3 := []ccv3.ConnectionWrapper{}
	authWrapperV2 := ccWrapper.NewUAAAuthentication(nil, config)
	authWrapperV3 := ccWrapper.NewUAAAuthentication(nil, config)

	ccWrappersV2 = append(ccWrappersV2, authWrapperV2)
	ccWrappersV2 = append(ccWrappersV2, ccWrapper.NewRetryRequest(config.RequestRetryCount()))

	ccWrappersV3 = append(ccWrappersV3, authWrapperV3)
	ccWrappersV3 = append(ccWrappersV3, ccWrapper.NewRetryRequest(config.RequestRetryCount()))

	ccClientV2 := ccv2.NewClient(ccv2.Config{
		AppName:            config.BinaryName(),
		AppVersion:         config.BinaryVersion(),
		JobPollingTimeout:  config.OverallPollingTimeout(),
		JobPollingInterval: config.PollingInterval(),
		Wrappers:           ccWrappersV2,
	})

	ccClientV3 := ccv3.NewClient(ccv3.Config{
		AppName:            config.BinaryName(),
		AppVersion:         config.BinaryVersion(),
		JobPollingTimeout:  config.OverallPollingTimeout(),
		JobPollingInterval: config.PollingInterval(),
		Wrappers:           ccWrappersV3,
	})

	_, err := ccClientV2.TargetCF(ccv2.TargetSettings{
		URL:               config.Target(),
		SkipSSLValidation: config.SkipSSLValidation(),
		DialTimeout:       config.DialTimeout(),
	})
	if err != nil {
		return fmt.Errorf("Error creating ccv2 client: %s", err)
	}
	if ccClientV2.AuthorizationEndpoint() == "" {
		return translatableerror.AuthorizationEndpointNotFoundError{}
	}

	_, err = ccClientV3.TargetCF(ccv3.TargetSettings{
		URL:               config.Target(),
		SkipSSLValidation: config.SkipSSLValidation(),
		DialTimeout:       config.DialTimeout(),
	})
	if err != nil {
		return fmt.Errorf("Error creating ccv3 client: %s", err)
	}
	// -------------------------

	// -------------------------
	// create an uaa client with cf_username/cf_password or client_id/client secret
	// to use it in v2 and v3 api for authenticate requests
	uaaClient := uaa.NewClient(config)

	uaaAuthWrapper := uaaWrapper.NewUAAAuthentication(nil, config)
	uaaClient.WrapConnection(uaaAuthWrapper)
	uaaClient.WrapConnection(uaaWrapper.NewRetryRequest(config.RequestRetryCount()))
	err = uaaClient.SetupResources(ccClientV2.AuthorizationEndpoint())
	if err != nil {
		return fmt.Errorf("Error setup resource uaa: %s", err)
	}

	// -------------------------
	// try connecting with pair given on uaa to retrieve access token and refresh token
	var accessToken string
	var refreshToken string
	if config.CFUsername() != "" {
		accessToken, refreshToken, err = uaaClient.Authenticate(map[string]string{
			"username": config.CFUsername(),
			"password": config.CFPassword(),
		}, "", constant.GrantTypePassword)
	} else if config.UAAOAuthClient() != "cf" {
		accessToken, refreshToken, err = uaaClient.Authenticate(map[string]string{
			"client_id":     config.UAAOAuthClient(),
			"client_secret": config.UAAOAuthClientSecret(),
		}, "", constant.GrantTypeClientCredentials)
	}
	if err != nil {
		return fmt.Errorf("Error when authenticate on cf: %s", err)
	}
	if accessToken == "" {
		return fmt.Errorf("A pair of username/password or a pair of client_id/client_secret muste be set.")
	}

	config.SetAccessToken(fmt.Sprintf("bearer %s", accessToken))
	config.SetRefreshToken(refreshToken)

	// -------------------------
	// assign uaa client to request wrappers
	uaaAuthWrapper.SetClient(uaaClient)
	authWrapperV2.SetClient(uaaClient)
	authWrapperV3.SetClient(uaaClient)
	// -------------------------

	// store client in the sessions
	s.ClientV2 = ccClientV2
	s.ClientV3 = ccClientV3
	// -------------------------

	// -------------------------
	// Create uaa client with given admin client_id only if user give it
	if configUaa.UAAOAuthClient() != "" {
		uaaClientSess := uaa.NewClient(configUaa)

		uaaAuthWrapperSess := uaaWrapper.NewUAAAuthentication(nil, configUaa)
		uaaClientSess.WrapConnection(uaaAuthWrapperSess)
		uaaClientSess.WrapConnection(uaaWrapper.NewRetryRequest(config.RequestRetryCount()))
		err = uaaClientSess.SetupResources(ccClientV2.AuthorizationEndpoint())
		if err != nil {
			return fmt.Errorf("Error setup resource uaa: %s", err)
		}

		accessTokenSess, refreshTokenSess, err := uaaClientSess.Authenticate(map[string]string{
			"client_id":     configUaa.UAAOAuthClient(),
			"client_secret": configUaa.UAAOAuthClientSecret(),
		}, "", constant.GrantTypeClientCredentials)
		if err != nil {
			return fmt.Errorf("Error when authenticate on uaa: %s", err)
		}
		if accessTokenSess == "" {
			return fmt.Errorf("A pair of pair of uaa_client_id/uaa_client_secret muste be set.")
		}
		configUaa.SetAccessToken(fmt.Sprintf("bearer %s", accessTokenSess))
		configUaa.SetRefreshToken(refreshTokenSess)
		s.ClientUAA = uaaClientSess
		uaaAuthWrapperSess.SetClient(uaaClientSess)
	}
	// -------------------------

	// -------------------------
	// Create raw http client with uaa client authentication to make raw request
	authWrapperRaw := ccWrapper.NewUAAAuthentication(nil, config)
	authWrapperRaw.SetClient(uaaClient)
	s.RawClient = NewRawClient(RawClientConfig{
		ApiEndpoint:       config.Target(),
		SkipSSLValidation: config.SkipSSLValidation(),
		DialTimeout:       config.DialTimeout(),
	}, authWrapperRaw, NewRetryRequest(config.RequestRetryCount()))

	s.HttpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.SkipSSLValidation(),
			},
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				KeepAlive: 30 * time.Second,
				Timeout:   config.DialTimeout(),
			}).DialContext,
		},
	}
	// -------------------------

	// -------------------------
	// Create router client for tcp routing
	routerConfig := router.Config{
		AppName:    config.BinaryName(),
		AppVersion: config.BinaryVersion(),
		ConnectionConfig: router.ConnectionConfig{
			DialTimeout:       config.DialTimeout(),
			SkipSSLValidation: config.SkipSSLValidation(),
		},
		RoutingEndpoint: ccClientV2.RoutingEndpoint(),
	}

	routerWrappers := []router.ConnectionWrapper{}

	rAuthWrapper := routerWrapper.NewUAAAuthentication(uaaClient, config)
	errorWrapper := routerWrapper.NewErrorWrapper()
	retryWrapper := newRetryRequestRouter(config.RequestRetryCount())

	routerWrappers = append(routerWrappers, rAuthWrapper, retryWrapper, errorWrapper)
	routerConfig.Wrappers = routerWrappers

	s.RouterClient = router.NewClient(routerConfig)
	// -------------------------

	// -------------------------
	// Create NOAA client for accessing logs from an app
	s.NOAAClient = NewNOAAClient(s.ClientV3.Logging(), config.SkipSSLValidation(), config, configSess.AppLogsMax)
	// -------------------------

	return nil
}

func (s *Session) loadUaaDefaultCfGroups() error {

	if s.ClientUAA == nil {
		return nil
	}
	client := s.ClientUAA

	// Retrieve default scope/groups for a new user by creating
	// a dummy user and extracting the default scope of that user
	username, err := uuid.GenerateUUID()
	if err != nil {
		return err
	}
	userResource := uaa.User{
		Username: username,
		Password: "password",
		Origin:   "uaa",
		Emails:   []uaa.Email{{Value: "email@domain.com"}},
	}
	user, err := client.CreateUserFromObject(userResource)
	if err != nil {
		return err
	}

	err = client.DeleteUser(user.ID)
	if err != nil {
		return err
	}
	for _, g := range user.Groups {
		s.uaaDefaultCfGroups[g.Name()] = g
	}

	return nil
}

func (s *Session) loadDefaultQuotaGuid() error {
	quotas, _, err := s.ClientV2.GetQuotas(ccv2cons.OrgQuota, ccv2.FilterByName("default"))
	if err != nil {
		return err
	}
	if len(quotas) == 0 {
		return fmt.Errorf("Can't found default quota")
	}
	s.defaultQuotaGuid = quotas[0].GUID
	return nil
}

// IsDefaultGroup -
func (s *Session) IsUaaDefaultCfGroup(group string) bool {
	_, ok := s.uaaDefaultCfGroups[group]
	return ok
}

// IsDefaultGroup -
func (s *Session) DefaultQuotaGuid() string {
	return s.defaultQuotaGuid
}
