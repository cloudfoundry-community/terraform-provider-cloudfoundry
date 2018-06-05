package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

const dlImportPath = "%s/v2/apps/%s/download"

func resourceAppImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	session := meta.(*cfapi.Session)
	if session == nil {
		return []*schema.ResourceData{}, fmt.Errorf("client is nil")
	}
	am := session.AppManager()
	mappings, err := am.ReadServiceBindingsByApp(d.Id())
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	// remove credentials from mapping, non string value can't be evaluated by terraform
	for i, mapping := range mappings {
		delete(mapping, "credentials")
		mappings[i] = mapping
	}

	err = d.Set("service_binding", mappings)
	if err != nil {
		return []*schema.ResourceData{}, err
	}
	d.Set("url", fmt.Sprintf(dlImportPath, session.Info().APIEndpoint, d.Id()))
	d.Set("route", make([]interface{}, 0))
	d.Set("timeout", DefaultAppTimeout)
	return ImportStatePassthrough(d, meta)
}
