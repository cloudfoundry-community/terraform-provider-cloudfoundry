package cloudfoundry

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	importStateKey = "is_import_state"
)

// getListOfStructs
func getListOfStructs(v interface{}) []map[string]interface{} {
	if vvSet, ok := v.(*schema.Set); ok {
		v = vvSet.List()
	}
	vvv := []map[string]interface{}{}
	for _, vv := range v.([]interface{}) {
		if vv == nil {
			continue
		}
		vvv = append(vvv, vv.(map[string]interface{}))
	}
	return vvv
}

// getResourceChange -
func getResourceChange(key string, d *schema.ResourceData) (bool, string, string) {
	old, new := d.GetChange(key)
	return old != new, old.(string), new.(string)
}

// getListChanges -
func getListChanges(old interface{}, new interface{}) (remove []string, add []string) {

	var a bool

	for _, o := range old.(*schema.Set).List() {
		remove = append(remove, o.(string))
	}
	for _, n := range new.(*schema.Set).List() {
		nn := n.(string)
		a = true
		for i, r := range remove {
			if nn == r {
				remove = append(remove[:i], remove[i+1:]...)
				a = false
				break
			}
		}
		if a {
			add = append(add, nn)
		}
	}
	return remove, add
}

// getListChanges -
func getMapChanges(old interface{}, new interface{}) (remove []string, add []string) {
	oldM := old.(map[string]interface{})
	newM := new.(map[string]interface{})
	oldL := make([]string, 0)
	for k := range oldM {
		if _, ok := newM[k]; !ok {
			oldL = append(oldL, k)
		}
		toDelete := true
		for kNew := range newM {
			if kNew == k {
				toDelete = false
				break
			}
		}
		if toDelete {
			remove = append(remove, k)
		}
	}
	for k := range newM {
		toAdd := true
		for _, kOld := range oldL {
			if kOld == k {
				toAdd = false
				break
			}
		}
		if toAdd {
			add = append(add, k)
		}
	}

	return remove, add
}

// getListChanges -
func getListMapChanges(old interface{}, new interface{}, match func(source, item map[string]interface{}) bool) (remove []map[string]interface{}, add []map[string]interface{}) {
	if vvSet, ok := old.(*schema.Set); ok {
		old = vvSet.List()
	}
	if vvSet, ok := new.(*schema.Set); ok {
		new = vvSet.List()
	}
	oldL := old.([]interface{})
	newL := new.([]interface{})

	for _, source := range oldL {
		toDelete := true
		for _, item := range newL {
			if match(source.(map[string]interface{}), item.(map[string]interface{})) {
				toDelete = false
				break
			}
		}
		if toDelete {
			remove = append(remove, source.(map[string]interface{}))
		}
	}
	for _, source := range newL {
		toAdd := true
		for _, item := range oldL {
			if match(source.(map[string]interface{}), item.(map[string]interface{})) {
				toAdd = false
				break
			}
		}
		if toAdd {
			add = append(add, source.(map[string]interface{}))
		}
	}

	return remove, add
}

// ImportRead -
func ImportRead(read schema.ReadFunc) schema.StateFunc {
	return func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
		MarkImportState(d)
		err := read(d, meta)
		if err != nil {
			return []*schema.ResourceData{}, err
		}
		return []*schema.ResourceData{d}, nil
	}
}

// MarkImportState -
func MarkImportState(d *schema.ResourceData) {
	connInfo := d.ConnInfo()
	if connInfo == nil {
		connInfo = make(map[string]string)
	}
	connInfo[importStateKey] = ""
	d.SetConnInfo(connInfo)
}

// IsImportState -
func IsImportState(d *schema.ResourceData) bool {
	connInfo := d.ConnInfo()
	if connInfo == nil {
		return false
	}
	_, ok := connInfo[importStateKey]
	return ok
}

func computeID(first, second string) string {
	return fmt.Sprintf("%s/%s", first, second)
}

func parseID(id string) (first string, second string, err error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("unable to parse ID '%s', expected format is '<guid>/<guid>'", id)
	} else {
		first = parts[0]
		second = parts[1]
	}
	return first, second, err
}

// return the intersection of 2 slices ([1, 1, 3, 4, 5, 6] & [2, 3, 6] >> [3, 6])
// sources and items must be array of whatever and element type can be whatever and can be different
// match function must return true if item and source given match
func intersectSlices(sources interface{}, items interface{}, match func(source, item interface{}) bool) []interface{} {
	sourceValue := reflect.ValueOf(sources)
	itemsValue := reflect.ValueOf(items)
	final := make([]interface{}, 0)
	for i := 0; i < sourceValue.Len(); i++ {
		inside := false
		src := sourceValue.Index(i).Interface()
		for p := 0; p < itemsValue.Len(); p++ {
			item := itemsValue.Index(p).Interface()
			if match(src, item) {
				inside = true
				break
			}
		}
		if inside {
			final = append(final, src)
		}
	}
	return final
}

// transforms list of struct to list of string id
func objectsToIds(objects interface{}, convert func(object interface{}) string) []interface{} {
	objectsValue := reflect.ValueOf(objects)
	ids := make([]interface{}, objectsValue.Len())
	for i := 0; i < objectsValue.Len(); i++ {
		object := objectsValue.Index(i).Interface()
		ids[i] = convert(object)
	}
	return ids
}

// Try to find in a list of whatever an element
func isInSlice(objects interface{}, match func(object interface{}) bool) bool {
	objectsValue := reflect.ValueOf(objects)
	for i := 0; i < objectsValue.Len(); i++ {
		object := objectsValue.Index(i).Interface()
		if match(object) {
			return true
		}
	}
	return false
}

// Try to find in a list of whatever an element
func getInSlice(objects interface{}, match func(object interface{}) bool) ([]interface{}, bool) {
	finalOjects := make([]interface{}, 0)
	objectsValue := reflect.ValueOf(objects)
	found := false
	for i := 0; i < objectsValue.Len(); i++ {
		object := objectsValue.Index(i).Interface()
		if match(object) {
			found = true
			finalOjects = append(finalOjects, object)
		}
	}
	return finalOjects, found
}
