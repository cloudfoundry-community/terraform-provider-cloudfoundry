package cloudfoundry

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/cfapi"
)

func resourceSpaceImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	session := meta.(*cfapi.Session)
	if session == nil {
		return []*schema.ResourceData{}, fmt.Errorf("client is nil")
	}
	sm := session.SpaceManager()
	asgIds, err := sm.ListStagingASGs(d.Id())
	if err != nil {
		return []*schema.ResourceData{}, err
	}
	d.Set("staging_asgs", asgIds)
	return ImportStatePassthrough(d, meta)
}
