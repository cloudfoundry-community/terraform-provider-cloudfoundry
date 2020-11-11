package cloudfoundry

import (
	"github.com/hashicorp/go-getter/helper/url"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestAppMigrateStateV0toV3(t *testing.T) {
	folderBits, _ = ioutil.TempDir("", "provider-cf-migrate-app")
	defer os.RemoveAll(folderBits)
	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"v0_3_bits_url": {
			StateVersion: 0,
			Attributes: map[string]string{
				"url": "https://github.com/cloudfoundry-community/tomee-buildpack/releases/download/v4.5.2/tomee-buildpack-v4.5.2.zip",
			},
			Expected: map[string]string{
				"path": "https://github.com/cloudfoundry-community/tomee-buildpack/releases/download/v4.5.2/tomee-buildpack-v4.5.2.zip",
			},
			Meta: testSession(),
		},
		"v2_3_bits_url": {
			StateVersion: 2,
			Attributes: map[string]string{
				"url": "https://github.com/cloudfoundry-community/tomee-buildpack/releases/download/v4.5.2/tomee-buildpack-v4.5.2.zip",
			},
			Expected: map[string]string{
				"path": "https://github.com/cloudfoundry-community/tomee-buildpack/releases/download/v4.5.2/tomee-buildpack-v4.5.2.zip",
			},
			Meta: testSession(),
		},
		"v2_3_bits_path": {
			StateVersion: 2,
			Attributes: map[string]string{
				"url": filepath.Join(testDir(), "dummy-app"),
			},
			Expected: map[string]string{
				"path": filepath.Join(folderBits, "dummy-app.zip"),
			},
			Meta: testSession(),
		},
		"v2_3_bits_git": {
			StateVersion: 2,
			Attributes: map[string]string{
				"git.#":     "1",
				"git.0.url": "https://github.com/cloudfoundry-community/tomee-buildpack.git",
				"git.0.tag": "v4.5.2",
			},
			Expected: map[string]string{
				"path": filepath.Join(folderBits, "github.com", "cloudfoundry-community", "tomee-buildpack.zip"),
			},
			Meta: testSession(),
		},
		"v2_3_bits_github": {
			StateVersion: 2,
			Attributes: map[string]string{
				"github_release.#":          "1",
				"github_release.0.owner":    "cloudfoundry-community",
				"github_release.0.repo":     "tomee-buildpack",
				"github_release.0.version":  "v4.5.2",
				"github_release.0.filename": "tomee-buildpack-v4.5.2.zip",
			},
			Expected: map[string]string{
				"path": "https://github.com/cloudfoundry-community/tomee-buildpack/releases/download/v4.5.2/tomee-buildpack-v4.5.2.zip",
			},
			Meta: testSession(),
		},
		"v2_3_bits_github_tarball": {
			StateVersion: 2,
			Attributes: map[string]string{
				"github_release.#":          "1",
				"github_release.0.owner":    "cloudfoundry-community",
				"github_release.0.repo":     "tomee-buildpack",
				"github_release.0.filename": "tarball",
			},
			Expected: map[string]string{
				"path": filepath.Join(folderBits, "github.com", "cloudfoundry-community", "tomee-buildpack", "archive.zip"),
			},
			Meta: testSession(),
		},
		"v2_3_route": {
			StateVersion: 2,
			Attributes: map[string]string{
				"ports.#":                          "1",
				"ports.8080":                       "8080",
				"route.#":                          "1",
				"route.0.default_route_mapping_id": "myroute",
			},
			Expected: map[string]string{
				"routes.#":                "1",
				"routes.2221701942.route": "myroute",
				"routes.2221701942.port":  "8080",
			},
			Meta: testSession(),
		},
		"v2_3_service_binding": {
			StateVersion: 2,
			Attributes: map[string]string{
				"service_binding.#":                  "1",
				"service_binding.0.service_instance": "myinstance",
				"service_binding.0.params.%":         "1",
				"service_binding.0.params.foo":       "bar",
			},
			Expected: map[string]string{
				"service_binding.#":                           "1",
				"service_binding.3162635407.service_instance": "myinstance",
				"service_binding.3162635407.params.%":         "1",
				"service_binding.3162635407.params.foo":       "bar",
			},
			Meta: testSession(),
		},
		"v2_3_routes": {
			StateVersion: 2,
			Attributes: map[string]string{
				"ports.#":              "2",
				"ports.8080":           "8080",
				"ports.9090":           "9090",
				"routes.#":             "1",
				"routes.8989809.route": "myroute",
			},
			Expected: map[string]string{
				"routes.#":                "2",
				"routes.2221701942.route": "myroute",
				"routes.2221701942.port":  "8080",
				"routes.634072338.route":  "myroute",
				"routes.634072338.port":   "9090",
			},
			Meta: testSession(),
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "an_app",
			Attributes: tc.Attributes,
		}
		is, err := resourceApp().MigrateState(
			tc.StateVersion, is, tc.Meta)

		if err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}

		for k, v := range tc.Expected {
			if is.Attributes[k] != v {
				t.Fatalf(
					"bad: %s\n\n expected: %#v -> %#v\n got: %#v -> %#v\n in: %#v",
					tn, k, v, k, is.Attributes[k], is.Attributes)
			}
			if k == "path" {
				u, _ := url.Parse(is.Attributes[k])
				if u.Scheme != "" {
					continue
				}
				stat, err := os.Stat(is.Attributes[k])
				if err != nil {
					t.Fatalf("Error occurred when retrieving path %s: %s", is.Attributes[k], err.Error())
				}
				if stat.Size() == 0 {
					t.Fatalf("Path %s seems to be an empty file, len: %d", is.Attributes[k], stat.Size())
				}
			}
		}
	}
}
