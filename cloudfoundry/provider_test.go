package cloudfoundry

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"crypto/tls"
	"fmt"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

var tstSession *managers.Session

var testOrgID string
var testOrgName string
var testSpaceID string
var testSpaceName string

var helperTest *HelpersTest

func init() {

	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"cloudfoundry": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {

	if !testAccEnvironmentSet() {
		t.Fatal("Acceptance environment has not been set.")
	}
}

func testAccEnvironmentSet() bool {

	endpoint := os.Getenv("CF_API_URL")
	user := os.Getenv("CF_USER")
	password := os.Getenv("CF_PASSWORD")
	uaaClientID := os.Getenv("CF_UAA_CLIENT_ID")
	uaaClientSecret := os.Getenv("CF_UAA_CLIENT_SECRET")

	if endpoint == "" ||
		user == "" ||
		password == "" ||
		uaaClientID == "" ||
		uaaClientSecret == "" {

		fmt.Println("CF_API_URL, CF_USER, CF_PASSWORD, CF_UAA_CLIENT_ID, CF_UAA_CLIENT_SECRET, " +
			" must be set for acceptance tests to work.")
		return false
	}
	return true
}

func testSession() *managers.Session {

	if !testAccEnvironmentSet() {
		panic(fmt.Errorf("ERROR! test CF_* environment variables have not been set"))
	}

	if tstSession == nil {
		c := managers.Config{
			Endpoint:        os.Getenv("CF_API_URL"),
			User:            os.Getenv("CF_USER"),
			Password:        os.Getenv("CF_PASSWORD"),
			CFClientID:      os.Getenv("CF_CLIENT_ID"),
			CFClientSecret:  os.Getenv("CF_CLIENT_SECRET"),
			UaaClientID:     os.Getenv("CF_UAA_CLIENT_ID"),
			UaaClientSecret: os.Getenv("CF_UAA_CLIENT_SECRET"),
		}

		c.SkipSslValidation, _ = strconv.ParseBool(os.Getenv("CF_SKIP_SSL_VALIDATION"))

		if c.User == "" && c.CFClientID == "" {
			panic(fmt.Errorf("Couple of user/password or uaa_client_id/uaa_client_secret must be set"))
		}
		if c.User != "" && c.CFClientID == "" {
			c.CFClientID = "cf"
			c.CFClientSecret = ""
		}
		var (
			err     error
			session *managers.Session
		)

		if session, err = managers.NewSession(c); err != nil {
			fmt.Printf("ERROR! Error creating a new session: %s\n", err.Error())
			panic(err.Error())
		}
		tstSession = session
		helperTest = &HelpersTest{tstSession}
	}
	return tstSession
}

func apiURL() string {
	return os.Getenv("CF_API_URL")
}

func defaultSysDomain() (domain string) {
	apiURL := apiURL()
	return apiURL[strings.Index(apiURL, ".")+1:]
}

func defaultAppDomain() (domain string) {
	if domain = os.Getenv("TEST_APP_DOMAIN"); len(domain) == 0 {
		domain = defaultSysDomain()
	}
	return domain
}

func defaultBaseDir() string {
	_, file, _, _ := runtime.Caller(1)
	return filepath.Dir(filepath.Dir(file))
}

func defaultTestOrg(t *testing.T) (string, string) {

	testOrgName := os.Getenv("TEST_ORG_NAME")
	if len(testOrgName) > 0 {

		var (
			err  error
			orgs []ccv2.Organization
		)
		session := testSession()
		client := session.ClientV2
		if orgs, _, err = client.GetOrganizations(ccv2.FilterByName(testOrgName)); err != nil {
			t.Fatal(err.Error())
		}
		testOrgID = orgs[0].GUID

	} else {
		t.Fatal("Environment variable TEST_ORG_NAME must be set for acceptance tests to work.")
	}

	return testOrgID, testOrgName
}

func defaultTestSpace(t *testing.T) (string, string) {

	testSpaceName := os.Getenv("TEST_SPACE_NAME")
	if len(testSpaceName) > 0 {

		var (
			err    error
			spaces []ccv2.Space
		)
		orgID, _ := defaultTestOrg(t)
		session := testSession()
		client := session.ClientV2
		if spaces, _, err = client.GetSpaces(
			ccv2.FilterByName(testSpaceName),
			ccv2.FilterEqual(constant.OrganizationGUIDFilter, orgID),
		); err != nil {
			t.Fatal(err.Error())
		}
		testSpaceID = spaces[0].GUID

	} else {
		t.Fatal("Environment variable TEST_SPACE_NAME must be set for acceptance tests to work.")
	}

	return testSpaceID, testSpaceName
}

func deleteServiceBroker(name string) {
	session := testSession()
	client := session.ClientV2
	sbs, _, err := client.GetServiceBrokers(ccv2.FilterByName(name))
	if err != nil {
		panic(err)
	}
	for _, sb := range sbs {
		helperTest.ForceDeleteServiceBroker(sb.GUID)
	}

}

func getTestSecurityGroup() string {
	defaultAsg := os.Getenv("TEST_DEFAULT_ASG")
	if len(defaultAsg) == 0 {
		defaultAsg = "public_networks"
	}
	return defaultAsg
}

func getTestDefaultIsolationSegment(t *testing.T) (string, string) {
	if os.Getenv("TEST_DEFAULT_SEGMENT") != "" {
		session := testSession()
		client := session.ClientV3
		segments, _, err := client.GetIsolationSegments(ccv3.Query{
			Key:    ccv3.NameFilter,
			Values: []string{os.Getenv("TEST_DEFAULT_SEGMENT")},
		})
		if err != nil {
			t.Fatal(err.Error())
		}

		return segments[0].GUID, segments[0].Name
	}
	t.Fatal("Environment variable TEST_ORG_NAME must be set for acceptance tests to work.")
	return "", ""
}

func getTestBrokerCredentials(t *testing.T) (
	serviceBrokerURL string,
	serviceBrokerUser string,
	serviceBrokerPassword string,
	serviceBrokerPlanPath string) {

	serviceBrokerURL = os.Getenv("TEST_SERVICE_BROKER_URL")
	serviceBrokerUser = os.Getenv("TEST_SERVICE_BROKER_USER")
	serviceBrokerPassword = os.Getenv("TEST_SERVICE_BROKER_PASSWORD")
	serviceBrokerPlanPath = os.Getenv("TEST_SERVICE_PLAN_PATH")

	if len(serviceBrokerURL) == 0 ||
		len(serviceBrokerUser) == 0 ||
		len(serviceBrokerPassword) == 0 ||
		len(serviceBrokerPlanPath) == 0 {

		t.Fatal("TEST_SERVICE_BROKER_URL, TEST_SERVICE_BROKER_USER, TEST_SERVICE_BROKER_PASSWORD " +
			"and TEST_SERVICE_PLAN_PATH must be set for acceptance tests to work.")
	}
	return
}

func getTestServiceBrokers(t *testing.T) (
	serviceName1 string,
	serviceName2 string,
	servicePlan string) {

	serviceName1 = os.Getenv("TEST_SERVICE_1")
	serviceName2 = os.Getenv("TEST_SERVICE_2")
	servicePlan = os.Getenv("TEST_SERVICE_PLAN")

	if len(serviceName1) == 0 ||
		len(serviceName2) == 0 ||
		len(servicePlan) == 0 {

		t.Fatal("TEST_SERVICE_1, TEST_SERVICE_2 and TEST_SERVICE_PLAN must be set for acceptance tests to work.")
	}
	return serviceName1, serviceName2, servicePlan
}

// func assertContains(str string, list []string) bool {
// 	for _, s := range list {
// 		if str == s {
// 			return true
// 		}
// 	}
// 	return false
// }

func assertSame(actual interface{}, expected interface{}) error {
	if actual != expected {
		return fmt.Errorf("expected '%s' found '%s' ", expected, actual)
	}
	return nil
}

func assertEquals(attributes map[string]string,
	key string, expected interface{}) error {
	v, ok := attributes[key]

	expectedValue := reflect.ValueOf(expected)

	if ok {

		var s string
		if expectedValue.Kind() == reflect.Ptr {

			if expectedValue.IsNil() {
				return fmt.Errorf("expected resource '%s' to not be present but it was '%s'", key, v)
			}

			expectedValueContent := reflect.Indirect(reflect.ValueOf(expected))
			switch expectedValueContent.Kind() {
			case reflect.String:
				s = expectedValueContent.String()
			case reflect.Int:
				s = fmt.Sprintf("%d", expectedValueContent.Int())
			case reflect.Bool:
				s = fmt.Sprintf("%t", expectedValueContent.Bool())
			default:
				return fmt.Errorf("unable to determine underlying content of expected value: %s", expectedValueContent.Kind())
			}
		} else {
			switch expected.(type) {
			case string:
				s = fmt.Sprintf("%s", expected)
			case int:
				s = fmt.Sprintf("%d", expected)
			case bool:
				s = fmt.Sprintf("%t", expected)
			default:
				s = fmt.Sprintf("%v", expected)
			}
		}
		if v != s {
			return fmt.Errorf("expected resource '%s' to be '%s' but it was '%s'", key, expected, v)
		}
	} else if expectedValue.Kind() == reflect.Ptr && !expectedValue.IsNil() {
		return fmt.Errorf("expected resource '%s' to be '%s' but it was not present", key, reflect.Indirect(reflect.ValueOf(expected)))
	}
	return nil
}

func assertListEquals(attributes map[string]string,
	key string, actualLen int,
	match func(map[string]string, int) bool) (err error) {

	var n int

	num := attributes[key+".#"]
	if len(num) > 0 {
		n, err = strconv.Atoi(num)
		if err != nil {
			return
		}
	} else {
		n = 0
	}

	if actualLen > 0 && n == 0 {
		return fmt.Errorf(
			"expected resource '%s' to be empty but it has '%d' elements", key, actualLen)
	}
	if actualLen != n {
		return fmt.Errorf(
			"expected resource '%s' to have '%d' elements but it has '%d' elements",
			key, n, actualLen)
	}
	if n > 0 {
		found := 0

		var (
			values map[string]string
			ok     bool
		)

		keyValues := make(map[string]map[string]string)
		for k, v := range attributes {
			keyParts := strings.Split(k, ".")
			if key == keyParts[0] && keyParts[1] != "#" {
				i := keyParts[1]
				if values, ok = keyValues[i]; !ok {
					values = make(map[string]string)
					keyValues[i] = values
				}
				if len(keyParts) == 2 {
					values["value"] = v
				} else {
					values[strings.Join(keyParts[2:], ".")] = v
				}
			}
		}

		for _, values := range keyValues {
			for j := 0; j < actualLen; j++ {
				if match(values, j) {
					found++
					break
				}
			}
		}
		if n != found {
			return fmt.Errorf(
				"expected list resource '%s' to match '%d' elements but matched only '%d' elements",
				key, n, found)
		}
	}
	return nil
}

func assertSetEquals(
	attributes map[string]string,
	key string,
	expected []interface{}) (err error) {

	var n int

	num := attributes[key+".#"]
	if len(num) > 0 {
		n, err = strconv.Atoi(num)
		if err != nil {
			return err
		}
	} else {
		n = 0
	}

	if len(expected) > 0 && n == 0 {
		return fmt.Errorf(
			"expected resource '%s' to be '%v' but it was empty", key, expected)
	}
	if len(expected) != n {
		return fmt.Errorf(
			"expected resource '%s' to have '%d' elements but it has '%d' elements",
			key, len(expected), n)
	}
	if n > 0 {
		found := 0
		for _, e := range expected {
			if _, ok := attributes[key+"."+strconv.Itoa(hashcode.String(e.(string)))]; ok {
				found++
			}
		}
		if n != found {
			return fmt.Errorf(
				"expected set resource '%s' to have elements '%v' but matched only '%d' elements",
				key, expected, found)
		}
	}
	return err
}

func assertMapEquals(key string, attributes map[string]string, actual map[string]interface{}) error {
	expected := make(map[string]interface{})
	for k, v := range attributes {
		keyParts := strings.Split(k, ".")
		if keyParts[0] == key && keyParts[1] != "%" {

			l := len(keyParts)
			m := expected
			for _, kk := range keyParts[1 : l-1] {
				if _, ok := m[kk]; !ok {
					m[kk] = make(map[string]interface{})
				}
				m = m[kk].(map[string]interface{})
			}
			m[keyParts[l-1]] = v
		}
	}

	normExpected := normalizeMap(expected, make(map[string]interface{}), "", "_")
	normActual := normalizeMap(actual, make(map[string]interface{}), "", "_")

	if !reflect.DeepEqual(normExpected, normActual) {
		return fmt.Errorf("map with key '%s' expected to be:\n%# v\nbut was:%# v", key, expected, actual)
	}
	return nil
}

func assertHTTPResponse(url string, expectedStatusCode int, expectedResponses *[]string) (err error) {

	var resp *http.Response

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	if resp, err = client.Get(url); err != nil {
		return err
	}
	if expectedStatusCode != resp.StatusCode {
		return fmt.Errorf(
			"expected response status code from url '%s' to be '%d', but actual was: %s",
			url, expectedStatusCode, resp.Status)
	}
	if expectedResponses != nil {
		in := resp.Body
		out := bytes.NewBuffer(nil)
		if _, err = io.Copy(out, in); err != nil {
			return err
		}
		content := out.String()

		found := false
		for _, e := range *expectedResponses {
			if e == content {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf(
				"expected response from url '%s' to be one of '%v', but actual was '%s'",
				url, *expectedResponses, content)
		}
	}
	return nil
}

func TestMain(m *testing.M) {
	if os.Getenv("TF_ACC_CREATE") == "" {
		os.Exit(m.Run())
	}

	// defer and os.Exit are not friends :(
	clean := make([]func(), 0)
	org, _, err := testSession().ClientV2.CreateOrganization("tf-acc-org", "")
	if err != nil {
		panic(err)
	}
	os.Setenv("TEST_ORG_NAME", org.Name)
	clean = append(clean, func() {
		j, _, err := testSession().ClientV2.DeleteOrganization(org.GUID)
		if err != nil {
			panic(err)
		}
		_, err = testSession().ClientV2.PollJob(j)
		if err != nil {
			panic(err)
		}
	})

	space, _, err := testSession().ClientV2.CreateSpace("tf-acc-space", org.GUID)
	if err != nil {
		panic(err)
	}
	os.Setenv("TEST_SPACE_NAME", space.Name)
	clean = append(clean, func() {
		j, _, err := testSession().ClientV2.DeleteSpace(space.GUID)
		if err != nil {
			panic(err)
		}
		_, err = testSession().ClientV2.PollJob(j)
		if err != nil {
			panic(err)
		}
	})

	segment, _, err := testSession().ClientV3.CreateIsolationSegment(ccv3.IsolationSegment{
		Name: "segment-one",
	})
	if err != nil {
		panic(err)
	}
	os.Setenv("TEST_DEFAULT_SEGMENT", segment.Name)
	clean = append(clean, func() {
		_, err := testSession().ClientV3.DeleteIsolationSegment(segment.GUID)
		if err != nil {
			panic(err)
		}
	})

	exitCode := m.Run()
	fmt.Println("Cleaning previous created resources")
	for i := len(clean) - 1; i >= 0; i-- {
		clean[i]()
	}
	os.Exit(exitCode)
}
