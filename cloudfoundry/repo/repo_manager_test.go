package repo_test

import "testing"

// checkError -
func checkError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("test failed with error: %s", err.Error())
		t.FailNow()
	}
}
