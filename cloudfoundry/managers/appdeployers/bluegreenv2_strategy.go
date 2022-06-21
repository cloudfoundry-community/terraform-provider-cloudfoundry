package appdeployers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	constantV3 "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/bits"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/raw"
)

type BlueGreenV2 struct {
	bitsManager *bits.BitsManager
	client      *ccv2.Client
	clientV3    *ccv3.Client
	rawClient   *raw.RawClient
	runBinder   *RunBinder
	standard    *Standard
}

type metadataRequest struct {
	Metadata metadata `json:"metadata"`
}

type metadataType string

type metadata struct {
	Labels      map[string]*string `json:"labels,omitempty"`
	Annotations map[string]*string `json:"annotations,omitempty"`
}

const (
	appMetadata          metadataType  = "apps"
	stopAppTimeout       time.Duration = 15 // CF SHOULD send a SIGKILL if an app is not stopped after 10 seconds
	delayBetweenRequests time.Duration = 1
)

func NewBlueGreenV2(bitsManager *bits.BitsManager, client *ccv2.Client, clientV3 *ccv3.Client, rawClient *raw.RawClient, runBinder *RunBinder, standard *Standard) *BlueGreenV2 {
	return &BlueGreenV2{
		bitsManager: bitsManager,
		client:      client,
		clientV3:    clientV3,
		rawClient:   rawClient,
		runBinder:   runBinder,
		standard:    standard,
	}
}

func (s BlueGreenV2) Deploy(appDeploy AppDeploy) (AppDeployResponse, error) {
	if appDeploy.App.State == constant.ApplicationStopped || appDeploy.App.GUID == "" {
		return s.standard.Deploy(appDeploy)
	}
	appDeploy.Mappings = clearMappingId(appDeploy.Mappings)
	appDeploy.ServiceBindings = clearBindingId(appDeploy.ServiceBindings)
	actions := Actions{
		{
			Forward: func(ctx Context) (Context, error) {
				_, _, err := s.client.UpdateApplication(ccv2.Application{
					GUID: appDeploy.App.GUID,
					Name: venerableAppName(appDeploy.App.Name),
				})
				return ctx, err
			},
		},
		{
			Forward: func(ctx Context) (Context, error) {
				app := appDeploy.App
				app.GUID = ""
				appResp, err := s.standard.Deploy(AppDeploy{
					App:             app,
					ServiceBindings: appDeploy.ServiceBindings,
					Mappings:        appDeploy.Mappings,
					Path:            appDeploy.Path,
					StageTimeout:    appDeploy.StageTimeout,
					BindTimeout:     appDeploy.BindTimeout,
					StartTimeout:    appDeploy.StartTimeout,
				})
				ctx["app_response"] = appResp
				return ctx, err
			},
			ReversePrevious: func(ctx Context) error {
				// if in error app must be already deleted by standard deployer
				// we only need to rename old app to its actual name
				_, _, err := s.client.UpdateApplication(ccv2.Application{
					GUID: appDeploy.App.GUID,
					Name: appDeploy.App.Name,
				})
				return err
			},
		},
		{
			Forward: func(ctx Context) (Context, error) {
				// copy metadata from original app since they do
				// not carry over in the ccv2.Application data structure
				appResp := ctx["app_response"].(AppDeployResponse)

				metadata, err := metadataRetrieve(appDeploy.App.GUID, appMetadata, s.rawClient)
				if err == nil {
					_ = metadataUpdate(appResp.App.GUID, appMetadata, s.rawClient, metadata)
				}
				return ctx, nil
			},
		},
		{
			Forward: func(ctx Context) (Context, error) {
				// Ask CF to stop application (desired state)
				_, _, err := s.clientV3.UpdateApplicationStop(appDeploy.App.GUID)
				return ctx, err
			},
		},
		{
			Forward: func(ctx Context) (Context, error) {
				// Ensure application is stopped before continuing
				channelIsStopped := make(chan bool, 1)
				channelError := make(chan error, 1)
				var err error

				go func() {
					isStopped := false
					for !isStopped {
						isStopped, err = isAppStopped(s.clientV3, appDeploy.App.GUID)
						time.Sleep(delayBetweenRequests * time.Second)
					}
					channelIsStopped <- isStopped
					channelError <- err
				}()

				select {
				case <-channelIsStopped:
				case <-time.After(stopAppTimeout * time.Second):
					// App is not in expected state (stopped) after waiting for the timeout
					log.Print("Timeout of reached while waiting for application to stop.")
				case <-channelError:
					return ctx, err
				}

				return ctx, err
			},
		},
		{
			Forward: func(ctx Context) (Context, error) {
				_, err := s.client.DeleteApplication(appDeploy.App.GUID)
				return ctx, err
			},
		},
	}
	ctx, err := actions.Execute()
	if err != nil {
		return AppDeployResponse{}, err
	}
	return ctx["app_response"].(AppDeployResponse), nil
}

func (s BlueGreenV2) Restage(appDeploy AppDeploy) (AppDeployResponse, error) {
	if appDeploy.App.State == constant.ApplicationStopped {
		return s.standard.Restage(appDeploy)
	}
	appDeploy.Mappings = clearMappingId(appDeploy.Mappings)
	appDeploy.ServiceBindings = clearBindingId(appDeploy.ServiceBindings)
	defaultReverse := func(ctx Context) error {
		appResp := ctx["app_response"].(AppDeployResponse)
		if appResp.App.GUID != "" {
			_, err := s.client.DeleteApplication(appResp.App.GUID)
			if err != nil {
				return err
			}
		}
		_, _, err := s.client.UpdateApplication(ccv2.Application{
			GUID: appDeploy.App.GUID,
			Name: appDeploy.App.Name,
		})
		return err
	}
	actions := Actions{
		{
			Forward: func(ctx Context) (Context, error) {
				_, _, err := s.client.UpdateApplication(ccv2.Application{
					GUID: appDeploy.App.GUID,
					Name: venerableAppName(appDeploy.App.Name),
				})
				return ctx, err
			},
		},
		{
			Forward: func(ctx Context) (Context, error) {
				app := appDeploy.App
				app.GUID = ""
				app.State = constant.ApplicationStopped
				appResp, err := s.standard.Deploy(AppDeploy{
					App:             app,
					ServiceBindings: appDeploy.ServiceBindings,
					Mappings:        appDeploy.Mappings,
					Path:            "",
					StageTimeout:    appDeploy.StageTimeout,
					BindTimeout:     appDeploy.BindTimeout,
					StartTimeout:    appDeploy.StartTimeout,
				})
				ctx["app_response"] = appResp
				return ctx, err
			},
			ReversePrevious: defaultReverse,
		},
		{
			Forward: func(ctx Context) (Context, error) {
				if appDeploy.IsDockerImage() {
					return ctx, nil
				}
				appResp := ctx["app_response"].(AppDeployResponse)
				err := s.bitsManager.CopyApp(appDeploy.App.GUID, appResp.App.GUID)
				return ctx, err
			},
			ReversePrevious: defaultReverse,
		},
		{
			Forward: func(ctx Context) (Context, error) {
				appResp := ctx["app_response"].(AppDeployResponse)
				app, err := s.runBinder.Start(AppDeploy{
					App:          appResp.App,
					StageTimeout: appDeploy.StageTimeout,
					BindTimeout:  appDeploy.BindTimeout,
					StartTimeout: appDeploy.StartTimeout,
				})
				if err != nil {
					return ctx, err
				}
				ctx["app_response"] = AppDeployResponse{
					App:             app,
					RouteMapping:    rejoinMappingPort(app.Ports[0], appResp.RouteMapping),
					ServiceBindings: appResp.ServiceBindings,
				}
				return ctx, err
			},
			ReversePrevious: defaultReverse,
		},
		{
			Forward: func(ctx Context) (Context, error) {
				// copy metadata from original app
				appResp := ctx["app_response"].(AppDeployResponse)

				metadata, err := metadataRetrieve(appDeploy.App.GUID, "apps", s.rawClient)
				if err == nil {
					_ = metadataUpdate(appResp.App.GUID, "apps", s.rawClient, metadata)
				}
				return ctx, nil
			},
		},
		{
			Forward: func(ctx Context) (Context, error) {
				_, err := s.client.DeleteApplication(appDeploy.App.GUID)
				return ctx, err
			},
		},
	}
	ctx, err := actions.Execute()
	if err != nil {
		return AppDeployResponse{}, err
	}
	return ctx["app_response"].(AppDeployResponse), nil
}

func (BlueGreenV2) IsCreateNewApp() bool {
	return true
}

func (BlueGreenV2) Names() []string {
	return []string{"blue-green", "blue-green-v2"}
}

// These methods should be integrated in the CCV clients at some point
func metadataUpdate(appGuid string, t metadataType, client *raw.RawClient, metadata metadata) error {
	if len(metadata.Labels) == 0 && len(metadata.Annotations) == 0 {
		return nil
	}

	b, err := json.Marshal(metadataRequest{Metadata: metadata})
	if err != nil {
		return err
	}

	endpoint := pathMetadata(t, appGuid)
	req, err := client.NewRequest("PATCH", endpoint, b)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	if resp.StatusCode != 200 && resp.StatusCode != 404 && resp.StatusCode != 202 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return ccerror.RawHTTPStatusError{
			StatusCode:  resp.StatusCode,
			RawResponse: b,
		}
	}
	return nil
}

func metadataRetrieve(appGuid string, t metadataType, client *raw.RawClient) (metadata, error) {
	path := pathMetadata(t, appGuid)
	req, err := client.NewRequest("GET", path, nil)
	if err != nil {
		return metadata{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return metadata{}, err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return metadata{}, err
	}

	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			return metadata{}, nil
		}
		return metadata{}, ccerror.RawHTTPStatusError{
			StatusCode:  resp.StatusCode,
			RawResponse: b,
		}
	}

	var metadataReq metadataRequest
	err = json.Unmarshal(b, &metadataReq)
	if err != nil {
		return metadata{}, err
	}
	return metadataReq.Metadata, nil
}

func pathMetadata(t metadataType, appGuid string) string {
	return fmt.Sprintf("/v3/%s/%s", t, appGuid)
}

// Check the app subprocesses state to ensure it is really down.
// Note: when you ask CF to stop an app, it tries to stop it gracefully, but after a timeout, it sends a SIGKILL
func isAppStopped(clientV3 *ccv3.Client, appGUID string) (bool, error) {
	isStopped := false
	processes, _, err := clientV3.GetApplicationProcesses(appGUID)

	if err == nil {
		for _, process := range processes {
			processInstances, _, err := clientV3.GetProcessInstances(process.GUID)
			if err == nil {
				for _, processInstance := range processInstances {
					if processInstance.State == constantV3.ProcessInstanceDown {
						isStopped = true
					} else if processInstance.State == constantV3.ProcessInstanceCrashed {
						isStopped = true
						log.Print("Process with GUID" + process.GUID + " has crashed. Considered stopped and proceeding with BlueGreen...")
					}
				}
			} else {
				break
			}
		}
	}

	return isStopped, err
}
