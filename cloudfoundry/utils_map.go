package cloudfoundry

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

func jsonNumToValue(in json.Number) interface{} {
	var result interface{}
	var err error
	result, err = in.Int64()
	if err == nil {
		return result
	}
	result, err = in.Float64()
	if err == nil {
		return result
	}
	return in.String()
}

func mapInterfaceToMapString(in map[string]interface{}) map[string]string {
	out := make(map[string]string)
	for k, v := range in {
		out[k] = fmt.Sprint(v)
	}
	return out
}

// normalizeMap -
func normalizeMap(in interface{}, outMap map[string]interface{}, key, delim string) map[string]interface{} {

	if in == nil {
		outMap[key] = ""
	} else {

		rt := reflect.TypeOf(in)

		if jsonNum, ok := in.(json.Number); ok {
			in = jsonNumToValue(jsonNum)
			rt = reflect.TypeOf(in)
		}

		switch rt.Kind() {

		case reflect.String:
			outMap[key] = in.(string)
		case reflect.Bool:
			outMap[key] = strconv.FormatBool(in.(bool))

		case reflect.Int:
			outMap[key] = strconv.FormatInt(int64(in.(int)), 10)
		case reflect.Int8:
			outMap[key] = strconv.FormatInt(int64(in.(int8)), 10)
		case reflect.Int16:
			outMap[key] = strconv.FormatInt(int64(in.(int16)), 10)
		case reflect.Int32:
			outMap[key] = strconv.FormatInt(int64(in.(int32)), 10)
		case reflect.Int64:
			outMap[key] = strconv.FormatInt(in.(int64), 10)

		case reflect.Uint:
			outMap[key] = strconv.FormatUint(uint64(in.(uint)), 10)
		case reflect.Uint8:
			outMap[key] = strconv.FormatUint(uint64(in.(uint8)), 10)
		case reflect.Uint16:
			outMap[key] = strconv.FormatUint(uint64(in.(uint16)), 10)
		case reflect.Uint32:
			outMap[key] = strconv.FormatUint(uint64(in.(uint32)), 10)
		case reflect.Uint64:
			outMap[key] = strconv.FormatUint(in.(uint64), 10)

		case reflect.Float32:
			outMap[key] = strconv.FormatFloat(float64(in.(float32)), 'g', -1, 32)
		case reflect.Float64:
			outMap[key] = strconv.FormatFloat(in.(float64), 'g', -1, 64)

		case reflect.Array, reflect.Slice:

			switch rt.Elem().Kind() {
			case reflect.String:
				for i, v := range in.([]string) {
					normalizeMap(v, outMap, fmt.Sprintf("%s%s%d", key, delim, i), delim)
				}
			case reflect.Bool:
				for i, v := range in.([]bool) {
					normalizeMap(v, outMap, fmt.Sprintf("%s%s%d", key, delim, i), delim)
				}
			case reflect.Int:
				for i, v := range in.([]int) {
					normalizeMap(v, outMap, fmt.Sprintf("%s%s%d", key, delim, i), delim)
				}
			case reflect.Int8:
				for i, v := range in.([]int8) {
					normalizeMap(v, outMap, fmt.Sprintf("%s%s%d", key, delim, i), delim)
				}
			case reflect.Int16:
				for i, v := range in.([]string) {
					normalizeMap(v, outMap, fmt.Sprintf("%s%s%d", key, delim, i), delim)
				}
			case reflect.Int32:
				for i, v := range in.([]int32) {
					normalizeMap(v, outMap, fmt.Sprintf("%s%s%d", key, delim, i), delim)
				}
			case reflect.Int64:
				for i, v := range in.([]int64) {
					normalizeMap(v, outMap, fmt.Sprintf("%s%s%d", key, delim, i), delim)
				}
			case reflect.Uint:
				for i, v := range in.([]uint) {
					normalizeMap(v, outMap, fmt.Sprintf("%s%s%d", key, delim, i), delim)
				}
			case reflect.Uint8:
				for i, v := range in.([]uint8) {
					normalizeMap(v, outMap, fmt.Sprintf("%s%s%d", key, delim, i), delim)
				}
			case reflect.Uint16:
				for i, v := range in.([]uint16) {
					normalizeMap(v, outMap, fmt.Sprintf("%s%s%d", key, delim, i), delim)
				}
			case reflect.Uint32:
				for i, v := range in.([]uint32) {
					normalizeMap(v, outMap, fmt.Sprintf("%s%s%d", key, delim, i), delim)
				}
			case reflect.Uint64:
				for i, v := range in.([]uint64) {
					normalizeMap(v, outMap, fmt.Sprintf("%s%s%d", key, delim, i), delim)
				}
			case reflect.Float32:
				for i, v := range in.([]float32) {
					normalizeMap(v, outMap, fmt.Sprintf("%s%s%d", key, delim, i), delim)
				}
			case reflect.Float64:
				for i, v := range in.([]float64) {
					normalizeMap(v, outMap, fmt.Sprintf("%s%s%d", key, delim, i), delim)
				}
			default:
				for i, v := range in.([]interface{}) {
					normalizeMap(v, outMap, fmt.Sprintf("%s%s%d", key, delim, i), delim)
				}
			}

		case reflect.Map:

			if len(key) != 0 {
				key = key + delim
			}

			switch rt.Elem().Kind() {
			case reflect.String:
				for k, v := range in.(map[string]string) {
					normalizeMap(v, outMap, key+k, delim)
				}
			case reflect.Bool:
				for k, v := range in.(map[string]bool) {
					normalizeMap(v, outMap, key+k, delim)
				}
			case reflect.Int:
				for k, v := range in.(map[string]int) {
					normalizeMap(v, outMap, key+k, delim)
				}
			case reflect.Int8:
				for k, v := range in.(map[string]int8) {
					normalizeMap(v, outMap, key+k, delim)
				}
			case reflect.Int16:
				for k, v := range in.(map[string]string) {
					normalizeMap(v, outMap, key+k, delim)
				}
			case reflect.Int32:
				for k, v := range in.(map[string]int32) {
					normalizeMap(v, outMap, key+k, delim)
				}
			case reflect.Int64:
				for k, v := range in.(map[string]int64) {
					normalizeMap(v, outMap, key+k, delim)
				}
			case reflect.Uint:
				for k, v := range in.(map[string]uint) {
					normalizeMap(v, outMap, key+k, delim)
				}
			case reflect.Uint8:
				for k, v := range in.(map[string]uint8) {
					normalizeMap(v, outMap, key+k, delim)
				}
			case reflect.Uint16:
				for k, v := range in.(map[string]uint16) {
					normalizeMap(v, outMap, key+k, delim)
				}
			case reflect.Uint32:
				for k, v := range in.(map[string]uint32) {
					normalizeMap(v, outMap, key+k, delim)
				}
			case reflect.Uint64:
				for k, v := range in.(map[string]uint64) {
					normalizeMap(v, outMap, key+k, delim)
				}
			case reflect.Float32:
				for k, v := range in.(map[string]float32) {
					normalizeMap(v, outMap, key+k, delim)
				}
			case reflect.Float64:
				for k, v := range in.(map[string]float64) {
					normalizeMap(v, outMap, key+k, delim)
				}
			default:
				for k, v := range in.(map[string]interface{}) {
					normalizeMap(v, outMap, key+k, delim)
				}
			}
		}
	}

	return outMap
}
