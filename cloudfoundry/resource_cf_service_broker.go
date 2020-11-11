package cloudfoundry

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

const (
	catalogEndpoint = "/v2/catalog"
)

func resourceServiceBroker() *schema.Resource {

	return &schema.Resource{

		Create: resourceServiceBrokerCreate,
		Read:   resourceServiceBrokerRead,
		Update: resourceServiceBrokerUpdate,
		Delete: resourceServiceBrokerDelete,

		Importer: &schema.ResourceImporter{
			State: ImportRead(resourceServiceBrokerRead),
		},

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"url": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"password": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"space": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"service_plans": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
			"services": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
			"catalog_hash": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"catalog_change": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Special marker to know and trigger a service broker update, this should not be set to true on your resource declaration",
			},
			"fail_when_catalog_not_accessible": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Set to true if you want to see errors when getting service broker catalog",
			},
			labelsKey:      labelsSchema(),
			annotationsKey: annotationsSchema(),
		},
	}
}

func resourceServiceBrokerCreate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	// do as first to not try add broker if catalog not accessible
	err := serviceBrokerUpdateCatalogSignature(d, meta)
	if err != nil {
		return err
	}

	sb, _, err := session.ClientV2.CreateServiceBroker(
		d.Get("name").(string),
		d.Get("username").(string),
		d.Get("password").(string),
		d.Get("url").(string),
		d.Get("space").(string),
	)
	if err != nil {
		return err
	}
	if err = readServiceDetail(sb.GUID, session, d); err != nil {
		return err
	}
	d.SetId(sb.GUID)

	err = metadataCreate(serviceBrokerMetadata, d, meta)
	if err != nil {
		return err
	}

	return nil
}

func resourceServiceBrokerRead(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	// do as first to not try add broker if catalog not accessible
	err := serviceBrokerUpdateCatalogSignature(d, meta)
	if err != nil {
		return err
	}

	sb, _, err := session.ClientV2.GetServiceBroker(d.Id())
	if err != nil {
		if IsErrNotFound(err) {
			d.SetId("")
			return nil
		}
		return err
	}
	err = readServiceDetail(d.Id(), session, d)
	if err != nil {
		return err
	}

	d.Set("name", sb.Name)
	d.Set("url", sb.BrokerURL)
	d.Set("username", sb.AuthUsername)
	d.Set("space", sb.SpaceGUID)

	err = metadataRead(serviceBrokerMetadata, d, meta, false)
	if err != nil {
		return err
	}
	return nil
}

func resourceServiceBrokerUpdate(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)

	// do as first to not try add broker if catalog not accessible
	err := serviceBrokerUpdateCatalogSignature(d, meta)
	if err != nil {
		return err
	}

	_, _, err = session.ClientV2.UpdateServiceBroker(ccv2.ServiceBroker{
		GUID:         d.Id(),
		AuthUsername: d.Get("username").(string),
		AuthPassword: d.Get("password").(string),
		BrokerURL:    d.Get("url").(string),
		SpaceGUID:    d.Get("space").(string),
		Name:         d.Get("name").(string),
	})
	if err != nil {
		return err
	}

	if err = readServiceDetail(d.Id(), session, d); err != nil {
		return err
	}

	err = metadataUpdate(serviceBrokerMetadata, d, meta)
	if err != nil {
		return err
	}
	return nil
}

func resourceServiceBrokerDelete(d *schema.ResourceData, meta interface{}) error {
	session := meta.(*managers.Session)
	if !session.PurgeWhenDelete {
		_, err := session.ClientV2.DeleteServiceBroker(d.Id())
		return err
	}

	svcs, _, err := session.ClientV2.GetServices(ccv2.FilterEqual(constant.ServiceBrokerGUIDFilter, d.Id()))
	if err != nil {
		return err
	}
	for _, svc := range svcs {
		sis, _, err := session.ClientV2.GetServiceInstances(ccv2.FilterEqual(constant.ServiceGUIDFilter, svc.GUID))
		if err != nil {
			return err
		}
		for _, si := range sis {
			_, _, err := session.ClientV2.DeleteServiceInstance(si.GUID, true, true)
			if err != nil {
				return err
			}
		}
	}
	_, err = session.ClientV2.DeleteServiceBroker(d.Id())
	return err
}

func readServiceDetail(id string, session *managers.Session, d *schema.ResourceData) error {
	services, _, err := session.ClientV2.GetServices(ccv2.FilterEqual(constant.ServiceBrokerGUIDFilter, id))
	if err != nil {
		return err
	}

	servicePlansTf := make(map[string]interface{})
	servicesTf := make(map[string]interface{})
	for _, s := range services {
		servicesTf[s.Label] = s.GUID
		servicePlans, _, err := session.ClientV2.GetServicePlans(ccv2.FilterEqual(constant.ServiceGUIDFilter, s.GUID))
		if err != nil {
			return err
		}
		for _, sp := range servicePlans {
			servicePlansTf[s.Label+"/"+sp.Name] = sp.GUID
		}
	}
	d.Set("service_plans", servicePlansTf)
	d.Set("services", servicesTf)

	return err
}

func serviceBrokerUpdateCatalogSignature(d *schema.ResourceData, meta interface{}) error {
	signature, err := serviceBrokerCatalogSignature(d, meta)
	failNotAccessible := d.Get("fail_when_catalog_not_accessible").(bool)
	if d.HasChange("fail_when_catalog_not_accessible") {
		_, newFailNotAccessible := d.GetChange("fail_when_catalog_not_accessible")
		failNotAccessible = newFailNotAccessible.(bool)
	}
	if err != nil && failNotAccessible {
		return fmt.Errorf("Error when getting catalog signature: %s", err.Error())
	}
	if err != nil {
		log.Printf(
			"[WARN] skipping generating catalog sha1, error during request creation: %s",
			err.Error(),
		)
		return nil
	}
	previousSignature := d.Get("catalog_hash")
	d.Set("catalog_hash", signature)
	if d.IsNewResource() {
		return nil
	}
	d.Set("catalog_change", previousSignature != signature)
	return nil
}

func serviceBrokerCatalogSignature(d *schema.ResourceData, meta interface{}) (string, error) {
	client := meta.(*managers.Session).HttpClient
	catalogUrl := d.Get("url").(string)
	if strings.HasSuffix(catalogUrl, "/") {
		catalogUrl = strings.TrimSuffix(catalogUrl, "/")
	}
	catalogUrl += catalogEndpoint
	req, err := http.NewRequest("GET", catalogUrl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("X-Broker-API-Version", "2.11")
	req.SetBasicAuth(d.Get("username").(string), d.Get("password").(string))
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Status code: %s, Body: %s ", resp.Status, string(bodyBytes))
	}

	h := sha1.New()
	_, err = h.Write(bodyBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(h.Sum(nil)), nil
}
