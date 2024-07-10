package v3appdeployers

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

// SetCurrentRevision : set revision for fallback
func (a Actor) SetCurrentRevision(appDeploy AppDeploy, _ FallbackFunction) Action {
    return Action{
        Forward: func(ctx Context) (Context, error) {
            appResp := ctx["app_response"].(AppDeployResponse)

            // Action code
            revisions, _, err := a.client.GetApplicationRevisions(ccv3.Query{
                Key:    ccv3.AppGUIDFilter,
                Values: []string{appResp.App.GUID},
            })
            if err != nil {
                return ctx, err
            }

            if len(revisions) != 0 {
                ctx["revisions"] = revisions[0].GUID
            }

            ctx["app_response"] = appResp
            return ctx, nil
        },
        ReversePrevious: a.ReverseActionBlank,
    }
}

