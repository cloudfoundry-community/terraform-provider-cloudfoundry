package cfapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"code.cloudfoundry.org/cli/cf/actors"
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/applicationbits"
	"code.cloudfoundry.org/cli/cf/api/applications"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/appfiles"
	"code.cloudfoundry.org/cli/cf/commands/application"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/formatters"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal"
)

// AppStopped -
var AppStopped = "STOPPED"

// AppStarted -
var AppStarted = "STARTED"

// AppManager -
type AppManager struct {
	log *Logger

	config    coreconfig.Reader
	ccGateway net.Gateway

	apiEndpoint string

	appRepo     applications.Repository
	appBitsRepo applicationbits.Repository

	appFiles  appfiles.ApplicationFiles
	appZipper appfiles.ApplicationZipper

	pushActor actors.PushActor
	starter   application.Start
}

// CCApp -
type CCApp struct {
	ID string

	Name                    string                  `json:"name,omitempty"`
	SpaceGUID               string                  `json:"space_guid,omitempty"`
	Ports                   *[]int                  `json:"ports,omitempty"`
	Instances               *int                    `json:"instances,omitempty"`
	Memory                  *int                    `json:"memory,omitempty"`
	DiskQuota               *int                    `json:"disk_quota,omitempty"`
	StackGUID               *string                 `json:"stack_guid,omitempty"`
	State                   *string                 `json:"state,omitempty"`
	PackageState            *string                 `json:"package_state,omitempty"`
	Buildpack               *string                 `json:"buildpack,omitempty"`
	Command                 *string                 `json:"command,omitempty"`
	EnableSSH               *bool                   `json:"enable_ssh,omitempty"`
	StagingFailedReason     *string                 `json:"staging_failed_reason,omitempty"`
	StagingFailedDesc       *string                 `json:"staging_failed_description,omitempty"`
	HealthCheckHTTPEndpoint *string                 `json:"health_check_http_endpoint,omitempty"`
	HealthCheckType         *string                 `json:"health_check_type,omitempty"`
	HealthCheckTimeout      *int                    `json:"health_check_timeout,omitempty"`
	Environment             *map[string]interface{} `json:"environment_json,omitempty"`
}

// CCAppResource -
type CCAppResource struct {
	Metadata resources.Metadata `json:"metadata"`
	Entity   CCApp              `json:"entity"`
}

const appStatePingSleep = time.Second * 5

// newAppManager -
func newAppManager(config coreconfig.Reader, ccGateway net.Gateway,
	domainRepository api.DomainRepository, routeRepository api.RouteRepository, logger *Logger) (am *AppManager, err error) {

	am = &AppManager{
		log: logger,

		config:    config,
		ccGateway: ccGateway,

		apiEndpoint: config.APIEndpoint(),

		appRepo:     applications.NewCloudControllerRepository(config, ccGateway),
		appBitsRepo: applicationbits.NewCloudControllerApplicationBitsRepository(config, ccGateway),

		appFiles:  appfiles.ApplicationFiles{},
		appZipper: appfiles.ApplicationZipper{},

		starter: application.Start{},
	}
	am.pushActor = actors.NewPushActor(am.appBitsRepo, am.appZipper, am.appFiles, nil)
	return
}

// FindApp -
func (am *AppManager) FindApp(appName string) (app CCApp, err error) {

	if err = am.ccGateway.ListPaginatedResources(am.apiEndpoint,
		fmt.Sprintf("/v2/apps?q=name:%s", appName),
		CCAppResource{}, func(resource interface{}) bool {

			appResource := resource.(CCAppResource)
			app = appResource.Entity
			app.ID = appResource.Metadata.GUID

			return false
		}); err != nil {

		return
	}

	if len(app.ID) == 0 {
		err = errors.NewModelNotFoundError("Application", appName)
	}
	return
}

// ReadApp -
func (am *AppManager) ReadApp(appID string) (app CCApp, err error) {

	resource := CCAppResource{}
	err = am.ccGateway.GetResource(
		fmt.Sprintf("%s/v2/apps/%s", am.apiEndpoint, appID), &resource)

	app = resource.Entity
	app.ID = resource.Metadata.GUID
	return
}

// CreateApp -
func (am *AppManager) CreateApp(a CCApp) (app CCApp, err error) {

	body, err := json.Marshal(a)
	if err != nil {
		return
	}

	resource := CCAppResource{}
	if err = am.ccGateway.CreateResource(am.apiEndpoint,
		"/v2/apps", bytes.NewReader(body), &resource); err != nil {
		return
	}
	app = resource.Entity
	app.ID = resource.Metadata.GUID
	return
}

// UpdateApp -
func (am *AppManager) UpdateApp(a CCApp) (app CCApp, err error) {

	body, err := json.Marshal(a)
	if err != nil {
		return
	}

	request, err := am.ccGateway.NewRequest("PUT",
		fmt.Sprintf("%s/v2/apps/%s", am.apiEndpoint, a.ID),
		am.config.AccessToken(), bytes.NewReader(body))
	if err != nil {
		return
	}

	resource := CCAppResource{}
	_, err = am.ccGateway.PerformRequestForJSONResponse(request, &resource)

	app = resource.Entity
	app.ID = resource.Metadata.GUID
	return
}

// DeleteApp -
func (am *AppManager) DeleteApp(appID string, deleteServiceBindings bool) (err error) {

	if deleteServiceBindings {

		var mappings []map[string]interface{}

		if mappings, err = am.ReadServiceBindingsByApp(appID); err != nil {
			return
		}
		for _, m := range mappings {
			if bindingID, ok := m["binding_id"]; ok {
				if err = am.DeleteServiceBinding(bindingID.(string)); err != nil {
					return
				}
			}
		}
	}

	err = am.ccGateway.DeleteResource(am.apiEndpoint, fmt.Sprintf("/v2/apps/%s", appID))
	return
}

// UploadApp -
func (am *AppManager) UploadApp(app CCApp, path string) (err error) {

	err = am.pushActor.ProcessPath(path, func(appDir string) error {
		localFiles, err := am.appFiles.AppFilesInDir(appDir)
		if err != nil {
			return fmt.Errorf("error processing app files in '%s': %s", path, err.Error())
		}

		if len(localFiles) == 0 {
			return fmt.Errorf("no app files found in '%s'", path)
		}

		am.log.UI.Say("Uploading %s...", terminal.EntityNameColor(app.Name))

		uploadDir, err := ioutil.TempDir("", "apps")
		if err != nil {
			return err
		}
		defer os.RemoveAll(uploadDir)

		remoteFiles, hasFileToUpload, err := am.pushActor.GatherFiles(localFiles, appDir, uploadDir, true)
		if err != nil {
			return err
		}

		if httpError, isHTTPError := err.(errors.HTTPError); isHTTPError && httpError.StatusCode() == 504 {

			am.log.UI.Warn("Resource matching API timed out; pushing all app files.")
			remoteFiles, hasFileToUpload, err = am.pushActor.GatherFiles(localFiles, appDir, uploadDir, false)
			if err != nil {
				return err
			}
		}

		zipFile, err := ioutil.TempFile("", "uploads")
		if err != nil {
			return err
		}
		defer func() {
			zipFile.Close()
			os.Remove(zipFile.Name())
		}()

		if hasFileToUpload {
			err = am.appZipper.Zip(uploadDir, zipFile)
			if err != nil {
				if emptyDirErr, ok := err.(*errors.EmptyDirError); ok {
					return emptyDirErr
				}
				return fmt.Errorf("Error zipping application: %s", err.Error())
			}

			var zipFileSize int64
			zipFileSize, err = am.appZipper.GetZipSize(zipFile)
			if err != nil {
				return err
			}

			zipFileCount := am.appFiles.CountFiles(uploadDir)
			if zipFileCount > 0 {
				am.log.UI.Say("Uploading app files from: %s", appDir)
				am.log.UI.Say("Uploading %s, %d files", formatters.ByteSize(zipFileSize), zipFileCount)
			}
		}

		if err = am.pushActor.UploadApp(app.ID, zipFile, remoteFiles); err != nil {
			return fmt.Errorf("error uploading application.\n%s", err.Error())
		}

		am.log.UI.Ok()
		return nil
	})

	return
}

// StartApp -
func (am *AppManager) StartApp(appID string, timeout time.Duration) (err error) {

	var app CCApp

	if app, err = am.ReadApp(appID); err != nil {
		return
	}
	if app.State != nil && *app.State == AppStopped {
		app.State = &AppStarted
		if app, err = am.UpdateApp(app); err != nil {
			return
		}

		err = am.waitForAppToStart(app, timeout)
	}
	return
}

// waitForAppToStart -
func (am *AppManager) waitForAppToStart(app CCApp, timeout time.Duration) (err error) {

	am.log.UI.Say("Waiting for app %s to start.", terminal.EntityNameColor(app.Name))

	c := make(chan error, 1)
	go func() {

		var err error

		for {
			if app, err = am.ReadApp(app.ID); err != nil {
				c <- err
				return
			}
			if app.State != nil {
				if *app.State == "STOPPED" {
					c <- fmt.Errorf("app %s failed to start", app.Name)
					return
				}
				if *app.State == AppStarted {

					response := make(map[string]interface{})
					if err = am.ccGateway.GetResource(fmt.Sprintf("%s/v2/apps/%s/stats", am.apiEndpoint, app.ID), &response); err != nil {
						c <- err
						return
					}
					if i, ok := response["0"]; ok {
						state := i.(map[string]interface{})["state"].(string)
						switch state {
						case "CRASHED":
							c <- fmt.Errorf("app %s crashed", app.Name)
							return
						case "RUNNING":
							c <- nil
							return
						}
					}
				}
			}
			time.Sleep(appStatePingSleep)
		}
	}()

	select {
	case err = <-c:
		if err != nil {
			return
		}
		am.log.UI.Say("App %s is running.", terminal.EntityNameColor(app.Name))
	case <-time.After(timeout):
		err = fmt.Errorf("app %s failed to start after %d seconds", app.Name, timeout/time.Second)
	}
	return
}

// RestageApp -
func (am *AppManager) RestageApp(appID string, timeout time.Duration) (err error) {

	request, err := am.ccGateway.NewRequest("POST",
		fmt.Sprintf("%s/v2/apps/%s/restage", am.apiEndpoint, appID),
		am.config.AccessToken(), bytes.NewReader([]byte{}))
	if err != nil {
		return
	}

	resource := CCAppResource{}
	_, err = am.ccGateway.PerformRequestForJSONResponse(request, &resource)

	app := resource.Entity
	app.ID = resource.Metadata.GUID

	err = am.waitForAppToStage(app, timeout)
	return
}

// waitForAppToStage -
func (am *AppManager) waitForAppToStage(app CCApp, timeout time.Duration) (err error) {

	am.log.UI.Say("Waiting for app %s to finish staging.", terminal.EntityNameColor(app.Name))

	c := make(chan error)
	go func() {

		var err error

		for {
			if app, err = am.ReadApp(app.ID); err != nil {
				c <- err
				return
			}
			if app.PackageState != nil {
				if *app.PackageState == "FAILED" {
					c <- fmt.Errorf("app %s failed to stage: %s => %s", app.Name, *app.StagingFailedReason, *app.StagingFailedDesc)
					return
				}
				if *app.PackageState == "STAGED" {
					c <- nil
					return
				}
			}
			time.Sleep(appStatePingSleep)
		}
	}()

	select {
	case err = <-c:
		if err != nil {
			return
		}
		am.log.UI.Say("App %s has finish staging.", terminal.EntityNameColor(app.Name))
	case <-time.After(timeout):
		err = fmt.Errorf("app %s failed to stage after %d seconds", app.Name, timeout/time.Second)
	}
	return
}

// StopApp -
func (am *AppManager) StopApp(appID string, timeout time.Duration) (err error) {

	var app CCApp

	if app, err = am.ReadApp(appID); err != nil {
		return
	}
	if app.State != nil && *app.State == AppStarted {
		app.State = &AppStopped
		if app, err = am.UpdateApp(app); err != nil {
			return
		}

		c := make(chan error)
		go func() {

			var err error

			for {
				time.Sleep(appStatePingSleep)
				if app, err = am.ReadApp(app.ID); err != nil {
					c <- err
					break
				}
				if app.State != nil && *app.State == "STOPPED" {
					c <- nil
					break
				}
			}
		}()

		select {
		case err = <-c:
			if err != nil {
				return
			}
		case <-time.After(timeout):
			err = fmt.Errorf("app %s failed to stop after %d seconds", app.Name, timeout/time.Second)
		}
	}
	return
}

// CreateServiceBinding -
func (am *AppManager) CreateServiceBinding(appID, serviceInstanceID string,
	params *map[string]interface{}) (bindingID string, credentials map[string]interface{}, err error) {

	request := map[string]interface{}{
		"app_guid":              appID,
		"service_instance_guid": serviceInstanceID,
	}
	if params != nil {
		request["parameters"] = *params
	}
	body, err := json.Marshal(request)
	if err != nil {
		return
	}

	response := make(map[string]interface{})
	if err = am.ccGateway.CreateResource(am.apiEndpoint,
		"/v2/service_bindings", bytes.NewReader(body), &response); err != nil {
		return
	}

	bindingID = response["metadata"].(map[string]interface{})["guid"].(string)
	if v, ok := response["entity"].(map[string]interface{})["credentials"]; ok {
		credentials = v.(map[string]interface{})
	}
	return
}

// ReadServiceBindingsByApp -
func (am *AppManager) ReadServiceBindingsByApp(appID string) (mappings []map[string]interface{}, err error) {
	return am.readServiceBindings(appID, "app_guid")
}

// ReadServiceBindingsByServiceInstance -
func (am *AppManager) ReadServiceBindingsByServiceInstance(serviceInstanceID string) ([]map[string]interface{}, error) {
	return am.readServiceBindings(serviceInstanceID, "service_instance_guid")
}

// readServiceBindings -
func (am *AppManager) readServiceBindings(id, key string) (mappings []map[string]interface{}, err error) {

	resource := make(map[string]interface{})

	if err = am.ccGateway.ListPaginatedResources(am.apiEndpoint,
		fmt.Sprintf("/v2/service_bindings?q=%s:%s", key, id),
		resource, func(resource interface{}) bool {

			routeResource := resource.(map[string]interface{})
			mapping := make(map[string]interface{})

			mapping["binding_id"] = routeResource["metadata"].(map[string]interface{})["guid"].(string)

			switch key {
			case "service_instance_guid":
				mapping["app"] = routeResource["entity"].(map[string]interface{})["app_guid"].(string)
			case "app_guid":
				mapping["service_instance"] = routeResource["entity"].(map[string]interface{})["service_instance_guid"].(string)
			default:
				mapping["app"] = routeResource["entity"].(map[string]interface{})["app_guid"].(string)
				mapping["service_instance"] = routeResource["entity"].(map[string]interface{})["service_instance_guid"].(string)
			}

			if v, ok := routeResource["entity"].(map[string]interface{})["credentials"]; ok {
				mapping["credentials"] = v.(map[string]interface{})
			}

			mappings = append(mappings, mapping)
			return true

		}); err != nil {

		return
	}
	return
}

// DeleteServiceBinding -
func (am *AppManager) DeleteServiceBinding(bindingID string) (err error) {
	err = am.ccGateway.DeleteResource(am.apiEndpoint, fmt.Sprintf("/v2/service_bindings/%s", bindingID))
	return
}
