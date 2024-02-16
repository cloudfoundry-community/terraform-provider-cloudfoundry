package cloudfoundry

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func resourceServiceInstanceSharing() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceInstanceSharingCreate,
		ReadContext:   resourceServiceInstanceSharingRead,
		DeleteContext: resourceServiceInstanceSharingDelete,

		Schema: map[string]*schema.Schema{
			"service_instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the service instance to share",
			},
			"space_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the space to share the service instance with, the space can be in the same or different org",
			},
		},
	}
}

func resourceServiceInstanceSharingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)

	serviceID, spaceID, err := parseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	spaceWithOrganizationList, _, err := session.ClientV3.GetServiceInstanceSharedSpaces(serviceID)

	if err != nil {
		return diag.FromErr(err)
	}

	found := false
	for _, spaceWithOrganization := range spaceWithOrganizationList {
		if spaceWithOrganization.SpaceGUID == spaceID {
			found = true
			break
		}
	}

	if !found {
		d.SetId("")
		return nil
	}

	d.Set("service_instance_id", serviceID)
	d.Set("space_id", spaceID)
	return nil
}

func resourceServiceInstanceSharingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	serviceID := d.Get("service_instance_id").(string)
	spaceID := d.Get("space_id").(string)

	spacesGUIDList, _, err := session.ClientV3.ShareServiceInstanceToSpaces(serviceID, []string{spaceID})

	if err != nil {
		return diag.FromErr(err)
	}

	if len(spacesGUIDList.GUIDs) < 1 {
		return diag.Errorf("failed to share service instance %s to space %s", serviceID, spaceID)
	}

	d.SetId(computeID(serviceID, spaceID))
	return nil
}

func resourceServiceInstanceSharingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	session := meta.(*managers.Session)
	serviceID := d.Get("service_instance_id").(string)
	spaceID := d.Get("space_id").(string)
	_, err := session.ClientV3.UnshareServiceInstanceFromSpace(serviceID, spaceID)

	return diag.FromErr(err)
}
