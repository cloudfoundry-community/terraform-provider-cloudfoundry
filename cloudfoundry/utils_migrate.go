package cloudfoundry

import (
	"strings"
)

func cleanByKeyAttribute(keyToClean string, m map[string]string) map[string]string {
	copyM := m
	for k := range copyM {
		kSplit := strings.Split(k, ".")
		if kSplit[0] == keyToClean {
			delete(m, k)
		}
	}
	return m
}
