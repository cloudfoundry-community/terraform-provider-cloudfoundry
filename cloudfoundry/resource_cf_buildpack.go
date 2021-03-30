package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

func resourceBuildpack() *schema.Resource {
	return &schema.Resource{

		Create: resourceBuildpackCreate,
		Read:   resourceBuildpackRead,
		Update: resourceBuildpackUpdate,
		Delete: resourceBuildpackDelete,

		Importer: &schema.ResourceImporter{
			State: ImportRead(resourceBuildpackRead),
		},
		SchemaVersion: 3,
		MigrateState:  resourceBuildpackMigrateState,
		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"position": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"locked": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"path": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Path to a buildpack zip in the form of unix path or http url",
			},
			"source_code_hash": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"filename": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			labelsKey:      labelsSchema(),
			annotationsKey: annotationsSchema(),
		},
	}
}

func resourceBuildpackCreate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	name := d.Get("name").(string)
	position := d.Get("position").(int)
	locked := d.Get("locked").(bool)
	enabled := d.Get("enabled").(bool)
	path := d.Get("path").(string)

	bp, _, err := session.ClientV2.CreateBuildpack(ccv2.Buildpack{
		Name:     name,
		Enabled:  BoolToNullBool(enabled),
		Locked:   BoolToNullBool(locked),
		Position: IntToNullInt(position),
	})
	if err != nil {
		return err
	}
	err = session.BitsManager.UploadBuildpack(bp.GUID, path)
	if err != nil {
		return err
	}
	bp, _, err = session.ClientV2.GetBuildpack(bp.GUID)
	if err != nil {
		return err
	}
	d.SetId(bp.GUID)
	d.Set("position", bp.Position.Value)
	d.Set("enabled", bp.Enabled.Value)
	d.Set("locked", bp.Locked.Value)
	d.Set("filename", bp.Filename)

	err = metadataCreate(buildpackMetadata, d, meta)
	if err != nil {
		return err
	}
	return nil
}

func resourceBuildpackRead(d *schema.ResourceData, meta interface{}) error {

	session := meta.(*managers.Session)

	bp, _, err := session.ClientV2.GetBuildpack(d.Id())
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", bp.Name)
	d.Set("position", bp.Position.Value)
	d.Set("enabled", bp.Enabled.Value)
	d.Set("locked", bp.Locked.Value)
	d.Set("filename", bp.Filename)

	err = metadataRead(buildpackMetadata, d, meta, false)
	if err != nil {
		return err
	}
	return err
}

func resourceBuildpackUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	session := meta.(*managers.Session)
	if session == nil {
		return fmt.Errorf("client is nil")
	}

	if d.HasChange("name") || d.HasChange("position") || d.HasChange("locked") || d.HasChange("enabled") {
		name := d.Get("name").(string)
		position := d.Get("position").(int)
		locked := d.Get("locked").(bool)
		enabled := d.Get("enabled").(bool)
		_, _, err := session.ClientV2.UpdateBuildpack(ccv2.Buildpack{
			GUID:     d.Id(),
			Name:     name,
			Enabled:  BoolToNullBool(enabled),
			Locked:   BoolToNullBool(locked),
			Position: IntToNullInt(position),
		})
		if err != nil {
			return err
		}
	}

	if d.HasChange("path") || d.HasChange("source_code_hash") || d.HasChange("filename") {
		return session.BitsManager.UploadBuildpack(d.Id(), d.Get("path").(string))
	}
	err = metadataUpdate(buildpackMetadata, d, meta)
	if err != nil {
		return err
	}
	return nil
}

func resourceBuildpackDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	_, err := session.ClientV2.DeleteBuildpack(d.Id())
	return err
}
