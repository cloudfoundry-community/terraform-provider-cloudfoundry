package managers

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/cfnetv1"
	netWrapper "code.cloudfoundry.org/cfnetworking-cli-api/cfnetworking/wrapper"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	ccWrapper "code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/api/router"
	routerWrapper "code.cloudfoundry.org/cli/api/router/wrapper"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/api/uaa/constant"
	uaaWrapper "code.cloudfoundry.org/cli/api/uaa/wrapper"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/configv3"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/appdeployers"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/bits"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/noaa"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/raw"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/v3appdeployers"
)

// Session - wraps the available clients from CF cli
type Session struct {
	ClientV2  *ccv2.Client
	ClientV3  *ccv3.Client
	ClientUAA *uaa.Client

	// Used for direct endpoint calls
	RawClient *raw.RawClient

	// http client used for normal request
	HttpClient *http.Client

	// To call tcp routing with this router
	RouterClient *router.Client

	// Manage upload bits like app and buildpack in full stream
	BitsManager *bits.BitsManager

	// NOAAClient permit to access to apps logs
	NOAAClient *noaa.NOAAClient

	// NetClient permit to access to networking policy api
	NetClient *cfnetv1.Client

	// Deployer is used to deploy an frim different strategy
	Deployer *appdeployers.Deployer

	// Deployer is used to deploy an frim different strategy
	V3Deployer *v3appdeployers.Deployer

	// Actor is a new type of deployer using v3 API and composable actions
	Actor *v3appdeployers.Actor

	// RunBinder is used to to manage start stop of an app
	RunBinder *appdeployers.RunBinder

	// V3RunBinder is used to to manage start stop of an app in v3
	V3RunBinder *v3appdeployers.RunBinder

	defaultQuotaGuid string

	PurgeWhenDelete bool

	Config Config

	ApiEndpoint string
}

type CFTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (t CFTokens) IsSet() bool {
	return t.AccessToken != ""
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
		PurgeWhenDelete: c.PurgeWhenDelete,
		ApiEndpoint:     c.Endpoint,
		Config:          c,
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
	uaaClientId := c.UaaClientID
	uaaClientSecret := c.UaaClientSecret
	if uaaClientId == "" {
		uaaClientId = c.CFClientID
		uaaClientSecret = c.CFClientSecret
	}
	configUaa := &configv3.Config{
		ConfigFile: configv3.JSONConfig{
			ConfigVersion:        3,
			UAAOAuthClient:       uaaClientId,
			UAAOAuthClientSecret: uaaClientSecret,
			SkipSSLValidation:    c.SkipSslValidation,
		},
	}

	err = s.init(config, configUaa, c)
	if err != nil {
		return nil, fmt.Errorf("Error when creating clients: %s", err.Error())
	}
	s.BitsManager = bits.NewBitsManager(s.ClientV2, s.ClientV3, s.RawClient, s.HttpClient)

	err = s.loadDefaultQuotaGuid(c.DefaultQuotaName)
	if err != nil {
		return nil, fmt.Errorf("Error when loading default quota: %s", err.Error())
	}
	s.loadDeployer()
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
	if IsDebugMode() {
		ccWrappersV2 = append(ccWrappersV2, ccWrapper.NewRequestLogger(NewRequestLogger()))
	}

	ccWrappersV3 = append(ccWrappersV3, authWrapperV3)
	ccWrappersV3 = append(ccWrappersV3, ccWrapper.NewRetryRequest(config.RequestRetryCount()))
	if IsDebugMode() {
		ccWrappersV3 = append(ccWrappersV3, ccWrapper.NewRequestLogger(NewRequestLogger()))
	}
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

	_, _, err = ccClientV3.TargetCF(ccv3.TargetSettings{
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

	uaaAuthWrapper := uaaWrapper.NewUAAAuthentication(nil, configUaa)
	uaaClient.WrapConnection(uaaAuthWrapper)
	uaaClient.WrapConnection(uaaWrapper.NewRetryRequest(config.RequestRetryCount()))
	err = uaaClient.SetupResources(ccClientV2.AuthorizationEndpoint())
	if err != nil {
		return fmt.Errorf("Error setup resource uaa: %s", err)
	}

	// -------------------------
	// Obtain access and refresh tokens
	var accessToken string
	var refreshToken string
	var errType string

	tokFromStore := s.loadTokFromStoreIfNeed(configSess.StoreTokensPath, uaaClient.RefreshAccessToken)
	if tokFromStore.IsSet() {
		accessToken = tokFromStore.AccessToken
		refreshToken = tokFromStore.RefreshToken
	} else if configSess.SSOPasscode != "" {
		// try connecting with SSO passcode to retrieve access token and refresh token
		accessToken, refreshToken, err = uaaClient.Authenticate(map[string]string{
			"passcode": configSess.SSOPasscode,
		}, "", constant.GrantTypePassword)
		errType = "SSO passcode"
	} else if config.CFUsername() != "" {
		// try connecting with pair given on uaa to retrieve access token and refresh token
		accessToken, refreshToken, err = uaaClient.Authenticate(map[string]string{
			"username": config.CFUsername(),
			"password": config.CFPassword(),
		}, "", constant.GrantTypePassword)
		errType = "username/password"
	} else if config.UAAOAuthClient() != "cf" {
		accessToken, refreshToken, err = uaaClient.Authenticate(map[string]string{
			"client_id":     config.UAAOAuthClient(),
			"client_secret": config.UAAOAuthClientSecret(),
		}, "", constant.GrantTypeClientCredentials)
		errType = "client_id/client_secret"
	}
	if err != nil {
		return fmt.Errorf("Error when authenticate on cf using %s: %s", errType, err)
	}
	if accessToken == "" {
		return fmt.Errorf("A pair of username/password, a pair of client_id/client_secret, or a SSO passcode must be set.")
	}

	config.SetAccessToken(fmt.Sprintf("bearer %s", accessToken))
	config.SetRefreshToken(refreshToken)

	// Write access and refresh tokens to file if needed
	err = s.saveTokToStoreIfNeed(configSess.StoreTokensPath, accessToken, refreshToken)
	if err != nil {
		return fmt.Errorf("Error when trying to save tokens to %s: %s", configSess.StoreTokensPath, err.Error())
	}
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

		var accessTokenSess string
		var refreshTokenSess string
		if configUaa.UAAOAuthClient() == "cf" {
			accessTokenSess = accessToken
			refreshTokenSess = refreshToken
		} else {
			accessTokenSess, refreshTokenSess, err = uaaClientSess.Authenticate(map[string]string{
				"client_id":     configUaa.UAAOAuthClient(),
				"client_secret": configUaa.UAAOAuthClientSecret(),
			}, "", constant.GrantTypeClientCredentials)
		}

		if err != nil {
			return fmt.Errorf("Error when authenticate on uaa [%s]: %s", configUaa.UAAOAuthClient(), err)
		}
		if accessTokenSess == "" {
			return fmt.Errorf("A pair of pair of uaa_client_id/uaa_client_secret must be set.")
		}
		configUaa.SetAccessToken(fmt.Sprintf("bearer %s", accessTokenSess))
		configUaa.SetRefreshToken(refreshTokenSess)
		s.ClientUAA = uaaClientSess
		uaaAuthWrapperSess.SetClient(uaaClientSess)
	}
	// -------------------------

	// -------------------------
	// Create cfnetworking client with uaa client authentication to call network policies
	netUaaAuthWrapper := netWrapper.NewUAAAuthentication(nil, config)
	netWrappers := []cfnetv1.ConnectionWrapper{
		netUaaAuthWrapper,
		netWrapper.NewRetryRequest(config.RequestRetryCount()),
	}
	netUaaAuthWrapper.SetClient(uaaClient)
	if IsDebugMode() {
		netWrappers = append(netWrappers, netWrapper.NewRequestLogger(NewRequestLogger()))
	}
	s.NetClient = cfnetv1.NewClient(cfnetv1.Config{
		SkipSSLValidation: config.SkipSSLValidation(),
		DialTimeout:       config.DialTimeout(),
		AppName:           config.BinaryName(),
		AppVersion:        config.BinaryVersion(),
		URL:               s.ClientV3.NetworkPolicyV1(),
		Wrappers:          netWrappers,
	})
	// -------------------------

	// -------------------------
	// Create raw http client with uaa client authentication to make raw request
	authWrapperRaw := ccWrapper.NewUAAAuthentication(nil, config)
	authWrapperRaw.SetClient(uaaClient)
	rawWrappers := []ccv3.ConnectionWrapper{
		authWrapperRaw,
		NewRetryRequest(config.RequestRetryCount()),
	}
	if IsDebugMode() {
		rawWrappers = append(rawWrappers, ccWrapper.NewRequestLogger(NewRequestLogger()))
	}
	s.RawClient = raw.NewRawClient(raw.RawClientConfig{
		ApiEndpoint:       config.Target(),
		SkipSSLValidation: config.SkipSSLValidation(),
		DialTimeout:       config.DialTimeout(),
	}, rawWrappers...)

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
	s.NOAAClient = noaa.NewNOAAClient(s.ClientV3.Logging(), config.SkipSSLValidation(), config, configSess.AppLogsMax)
	// -------------------------

	return nil
}

func (s *Session) loadDeployer() {
	s.RunBinder = appdeployers.NewRunBinder(s.ClientV2, s.NOAAClient)
	stdStrategy := appdeployers.NewStandard(s.BitsManager, s.ClientV2, s.RunBinder)
	bgStrategy := appdeployers.NewBlueGreenV2(s.BitsManager, s.ClientV2, s.ClientV3, s.RawClient, s.RunBinder, stdStrategy)
	s.Deployer = appdeployers.NewDeployer(stdStrategy, bgStrategy)

	s.V3RunBinder = v3appdeployers.NewRunBinder(s.ClientV3, s.NOAAClient)
	s.Actor = v3appdeployers.NewActor(s.BitsManager, s.ClientV3, s.RawClient, s.V3RunBinder)

	// Initialize deployment strategies in v3
	v3std := v3appdeployers.NewStandard(s.BitsManager, s.ClientV3, s.V3RunBinder, s.Actor)
	v3bg := v3appdeployers.NewBlueGreen(s.BitsManager, s.ClientV3, s.RawClient, s.V3RunBinder, v3std)

	// Initialize deployer for rolling
	rolling := v3appdeployers.NewRolling(s.Actor)

	s.V3Deployer = v3appdeployers.NewDeployer(v3std, v3bg, rolling)

}

func (s *Session) loadDefaultQuotaGuid(quotaName string) error {
	// quotas, _, err := s.ClientV2.GetQuotas(ccv2cons.OrgQuota, ccv2.FilterByName(quotaName))
	// if err != nil {
	// 	return err
	// }
	quotas, _, err := s.ClientV3.GetOrganizationQuotas(ccv3.Query{
		Key:    ccv3.NameFilter,
		Values: []string{quotaName},
	})
	if err != nil {
		return err
	}
	if len(quotas) == 0 {
		return fmt.Errorf("can't found default quota '%s'", quotaName)
	}
	s.defaultQuotaGuid = quotas[0].GUID
	return nil
}

func (s *Session) loadTokFromStoreIfNeed(storePath string, refresher func(refreshToken string) (uaa.RefreshedTokens, error)) CFTokens {
	if storePath == "" {
		return CFTokens{}
	}
	b, err := os.ReadFile(storePath)
	if err != nil {
		return CFTokens{}
	}
	var tokens CFTokens
	err = json.Unmarshal(b, &tokens)
	if err != nil {
		return CFTokens{}
	}
	refreshed, err := refresher(tokens.RefreshToken)
	if err != nil {
		return CFTokens{}
	}
	return CFTokens{
		AccessToken:  refreshed.AccessToken,
		RefreshToken: refreshed.RefreshToken,
	}
}

func (s *Session) saveTokToStoreIfNeed(storePath, accessToken, refreshToken string) error {
	if storePath == "" {
		return nil
	}
	b, _ := json.MarshalIndent(CFTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, "", "  ")
	return os.WriteFile(storePath, b, 0644)
}

// IsDefaultGroup -
func (s *Session) DefaultQuotaGuid() string {
	return s.defaultQuotaGuid
}

func IsDebugMode() bool {
	tfDebug := strings.ToLower(os.Getenv("TF_LOG"))
	return tfDebug == "info" || tfDebug == "trace" || tfDebug == "debug"
}
