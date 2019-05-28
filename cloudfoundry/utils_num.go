package cloudfoundry

import (
	"fmt"
	"strconv"
	"strings"
)

// resourceIntegerSet -
func resourceIntegerSet(v interface{}) int {
	return v.(int)
}

func portRangeParse(portRange string) (start int, end int, err error) {
	portRangeSplit := strings.Split(portRange, "-")
	if len(portRangeSplit) > 2 {
		return 0, 0, fmt.Errorf("Invalid range")
	}
	start, err = strconv.Atoi(portRangeSplit[0])
	if err != nil {
		return 0, 0, err
	}
	if len(portRangeSplit) == 1 {
		return start, start, nil
	}
	end, err = strconv.Atoi(portRangeSplit[1])
	if err != nil {
		return 0, 0, err
	}
	return start, end, nil
}
