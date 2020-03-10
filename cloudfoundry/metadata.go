package cloudfoundry

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"encoding/json"
	"fmt"
	"github.com/blang/semver"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"io/ioutil"
)

type metadataType string

type MetadataRequest struct {
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

const (
	labelsKey      = "labels"
	annotationsKey = "annotations"

	orgMetadata       metadataType = "organizations"
	spaceMetadata     metadataType = "spaces"
	buildpackMetadata metadataType = "buildpacks"
	appMetadata       metadataType = "apps"
	stackMetadata     metadataType = "stacks"
	segmentMetadata   metadataType = "isolation_segments"
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
	if !isMetadataApiCompat(meta) {
		return nil
	}
	return metadataUpdate(t, d, meta)
}

func isMetadataApiCompat(meta interface{}) bool {
	apiVersion := meta.(*managers.Session).ClientV3.CloudControllerAPIVersion()
	v, err := semver.Parse(apiVersion)
	if err != nil {
		// in case version is incorrect
		// we set true anyway, it will only do the calls to api but not fail if endpoint is not found in crud
		return true
	}
	expectedRange := semver.MustParseRange(">=3.63.0")
	return expectedRange(v)
}

func metadataUpdate(t metadataType, d *schema.ResourceData, meta interface{}) error {
	if !isMetadataApiCompat(meta) {
		return nil
	}
	metadata := resourceMetadataToMetadata(d)
	if len(metadata.Labels) == 0 && len(metadata.Annotations) == 0 &&
		!d.HasChange(labelsKey) && !d.HasChange(annotationsKey) {
		return nil
	}

	oldMetadata, err := metadataRetrieve(t, d, meta)
	if err != nil {
		return err
	}

	metadata = mergeMetadata(oldMetadata, metadata)

	b, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	client := meta.(*managers.Session).RawClient
	req, err := client.NewRequest("PUT", pathMetadata(t, d), ioutil.NopCloser(bytes.NewBuffer(b)))
	if err != nil {
		return err
	}
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
	if resp.StatusCode != 200 && resp.StatusCode != 404 {
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

func metadataRead(t metadataType, d *schema.ResourceData, meta interface{}, forceRead bool) error {
	if !isMetadataApiCompat(meta) {
		return nil
	}
	_, hasLabels := d.GetOk(labelsKey)
	_, hasAnnotations := d.GetOk(annotationsKey)
	if !hasAnnotations && !hasLabels && !forceRead && !IsImportState(d) {
		return nil
	}
	metadata := resourceMetadataToMetadata(d)
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

func resourceMetadataToMetadata(d *schema.ResourceData) Metadata {
	labels := mapInterfaceToMapString(d.Get(labelsKey).(map[string]interface{}))
	annotations := mapInterfaceToMapString(d.Get(annotationsKey).(map[string]interface{}))
	return Metadata{
		Labels:      labels,
		Annotations: annotations,
	}
}

func mergeMetadata(o Metadata, n Metadata) Metadata {
	labels := o.Labels
	for k, v := range n.Labels {
		labels[k] = v
	}

	annotations := o.Annotations
	for k, v := range n.Annotations {
		annotations[k] = v
	}
	return Metadata{
		Labels:      labels,
		Annotations: annotations,
	}
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
