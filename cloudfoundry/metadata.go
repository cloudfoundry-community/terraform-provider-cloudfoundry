package cloudfoundry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"github.com/blang/semver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
)

type metadataType string

type MetadataRequest struct {
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	Labels      map[string]*string `json:"labels,omitempty"`
	Annotations map[string]*string `json:"annotations,omitempty"`
}

const (
	labelsKey      = "labels"
	annotationsKey = "annotations"

	orgMetadata           metadataType = "organizations"
	spaceMetadata         metadataType = "spaces"
	buildpackMetadata     metadataType = "buildpacks"
	appMetadata           metadataType = "apps"
	stackMetadata         metadataType = "stacks"
	segmentMetadata       metadataType = "isolation_segments"
	serviceBrokerMetadata metadataType = "service_brokers"
)

func labelsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem:     schema.TypeString,
	}
}

func annotationsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem:     schema.TypeString,
	}
}

func metadataCreate(t metadataType, d *schema.ResourceData, meta interface{}) error {
	if !isMetadataAPICompat(t, meta) {
		return nil
	}
	return metadataUpdate(t, d, meta)
}

func isMetadataAPICompat(t metadataType, meta interface{}) bool {
	apiVersion := meta.(*managers.Session).ClientV3.CloudControllerAPIVersion()
	v, err := semver.Parse(apiVersion)
	if err != nil {
		// in case version is incorrect
		// we set true anyway, it will only do the calls to api but not fail if endpoint is not found in crud
		return true
	}

	expectedRange := semver.MustParseRange(">=3.63.0")
	if t == serviceBrokerMetadata {
		expectedRange = semver.MustParseRange(">=3.71.0")
	}
	return expectedRange(v)
}

func resourceToMetadata(d *schema.ResourceData) Metadata {
	return Metadata{
		Labels:      resourceToPayload(d, labelsKey),
		Annotations: resourceToPayload(d, annotationsKey),
	}
}

// resourceToPayload - create metadata update payload from resource state
//
// note: we *should* construct payload in a way where only new/changed value
//       are present, but re-giving existing values clarifies the code
//
// 1. construct payload as requested by "new" value
// 2. find delete keys and create { "key" : nil } in payload
//    ie: keys existing in "old" but not in "new"
func resourceToPayload(d *schema.ResourceData, key string) map[string]*string {
	res := map[string]*string{}
	old, new := d.GetChange(key)
	oldV := mapInterfaceToMapString(old.(map[string]interface{}))
	newV := mapInterfaceToMapString(new.(map[string]interface{}))

	// 1.
	for key, val := range newV {
		v := val
		res[key] = &v
	}

	// 2.
	for key := range oldV {
		if _, ok := newV[key]; !ok {
			res[key] = nil
		}
	}

	return res
}

func metadataRead(t metadataType, d *schema.ResourceData, meta interface{}, forceRead bool) error {
	if !isMetadataAPICompat(t, meta) {
		return nil
	}
	_, hasLabels := d.GetOk(labelsKey)
	_, hasAnnotations := d.GetOk(annotationsKey)
	if !hasAnnotations && !hasLabels && !forceRead && !IsImportState(d) {
		return nil
	}

	metadata := resourceToMetadata(d)
	oldMetadata, err := metadataRetrieve(t, d, meta)
	if err != nil {
		return err
	}

	labels := make(map[string]interface{})
	if IsImportState(d) {
		for k, v := range oldMetadata.Labels {
			labels[k] = v
		}
	} else {
		for k := range metadata.Labels {
			if _, ok := oldMetadata.Labels[k]; !ok {
				continue
			}
			labels[k] = oldMetadata.Labels[k]
		}
	}
	d.Set(labelsKey, labels)

	annotations := make(map[string]interface{})
	if IsImportState(d) {
		for k, v := range oldMetadata.Annotations {
			annotations[k] = v
		}
	} else {
		for k := range metadata.Annotations {
			if _, ok := oldMetadata.Annotations[k]; !ok {
				continue
			}
			annotations[k] = oldMetadata.Annotations[k]
		}
	}
	d.Set(annotationsKey, annotations)

	return nil
}

func metadataUpdate(t metadataType, d *schema.ResourceData, meta interface{}) error {
	if !isMetadataAPICompat(t, meta) {
		return nil
	}

	metadata := resourceToMetadata(d)
	if len(metadata.Labels) == 0 && len(metadata.Annotations) == 0 {
		return nil
	}

	b, err := json.Marshal(MetadataRequest{Metadata: metadata})
	if err != nil {
		return err
	}

	client := meta.(*managers.Session).RawClient
	endpoint := pathMetadata(t, d)
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

func metadataRetrieve(t metadataType, d *schema.ResourceData, meta interface{}) (Metadata, error) {
	client := meta.(*managers.Session).RawClient
	req, err := client.NewRequest("GET", pathMetadata(t, d), nil)
	if err != nil {
		return Metadata{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return Metadata{}, err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Metadata{}, err
	}

	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			return Metadata{}, nil
		}
		return Metadata{}, ccerror.RawHTTPStatusError{
			StatusCode:  resp.StatusCode,
			RawResponse: b,
		}
	}

	var metadataReq MetadataRequest
	err = json.Unmarshal(b, &metadataReq)
	if err != nil {
		return Metadata{}, err
	}
	return metadataReq.Metadata, nil
}

func pathMetadata(t metadataType, d *schema.ResourceData) string {
	return fmt.Sprintf("/v3/%s/%s", t, d.Id())
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
