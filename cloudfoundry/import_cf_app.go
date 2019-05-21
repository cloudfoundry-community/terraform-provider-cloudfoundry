package cloudfoundry

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"

	"github.com/hashicorp/terraform/helper/schema"
)

const dlImportPath = "%s/v2/apps/%s/download"

func resourceAppImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	session := meta.(*managers.Session)

	d.Set("path", fmt.Sprintf(dlImportPath, session.ApiEndpoint, d.Id()))
	d.Set("timeout", DefaultAppTimeout)
	return ImportStatePassthrough(d, meta)
}
