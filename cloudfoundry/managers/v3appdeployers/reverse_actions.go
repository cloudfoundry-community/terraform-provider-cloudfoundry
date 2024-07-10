package v3appdeployers

import (
	"log"
	"time"
)

// ReverseActionBlank : do nothing
func (a Actor) ReverseActionBlank(ctx Context) error {
	return nil
}

// ReverseActionDeleteApp : default reverse action
func (a Actor) ReverseActionDeleteApp(ctx Context) error {
	appResp := ctx["app_response"].(AppDeployResponse)
	if appResp.App.GUID == "" {
		return nil
	}
	_, _, err := a.client.DeleteApplication(appResp.App.GUID)
	return err
}

// ReverseActionDeployRevision : redeploy latest revision
func (a Actor) ReverseActionDeployRevision(ctx Context) error {
	ctxRevision := ctx["revisions"]
	appResp := ctx["app_response"].(AppDeployResponse)

	if appResp.App.GUID == "" {
		return nil
	}

	if ctxRevision != nil {
		revisionGUID := ctxRevision.(string)
		deploymentGUID, _, err := a.client.CreateApplicationDeployment(appResp.App.GUID, revisionGUID)
		if err != nil {
			return err
		}

		// Since we don't have access to user-specified timeout, it is hard-coded to 60s
		err = a.PollStartRolling(appResp.App, deploymentGUID, 1*time.Minute)
		if err != nil {
			return err
		}
	} else {
		log.Printf("No revisions found means delete app?")
		_, _, err := a.client.DeleteApplication(appResp.App.GUID)
		return err
	}

	return nil
}
