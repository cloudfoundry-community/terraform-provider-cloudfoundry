package appdeployers

import (
	"fmt"
)

func venerableAppName(appName string) string {
	return fmt.Sprintf("%s-venerable", appName)
}
