package jsonry

import (
	"encoding/json"
	"fmt"
	"reflect"

	"code.cloudfoundry.org/jsonry/internal/path"
	"code.cloudfoundry.org/jsonry/internal/tree"
)

// Marshal converts the specified Go struct into JSON. The input must be a struct or a pointer to a struct.
// Where a field is optional, the suffix ",omitempty" can be specified. This will mean that the field will
// be omitted from the JSON output if it is a nil pointer or has zero value for the type.
// When a field is a slice or an array, a single list hint "[]" may be specified in the JSONry path so that the array
// is created at the correct position in the JSON output.
//
// If a type implements the json.Marshaler interface, then the MarshalJSON() method will be called.
//
// If a type implements the jsonry.Omissible interface, then the OmitJSONry() method will be used to
// to determine whether or not to marshal the field, overriding any `,omitempty` tags.
//
// The field type can be string, bool, int*, uint*, float*, map, slice, array or struct. JSONry is recursive.
func Marshal(in interface{}) ([]byte, error) {
	iv := reflect.Indirect(reflect.ValueOf(in))

	if iv.Kind() != reflect.Struct {
		return nil, fmt.Errorf(`the input must be a struct, not "%s"`, iv.Kind())
	}

	m, err := marshalStruct(iv)
	if err != nil {
		return nil, err
	}

	return json.Marshal(m)
}

func marshalStruct(in reflect.Value) (map[string]interface{}, error) {
	out := make(tree.Tree)
	t := in.Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		if public(f) {
			path := path.ComputePath(f)
			val := in.Field(i)

			if shouldMarshal(path, val) {
				r, err := marshal(val)
				if err != nil {
					return nil, wrapErrorWithFieldContext(err, f.Name, f.Type)
				}

				out.Attach(path, r)
			}
		}
	}

	return out, nil
}

func marshal(in reflect.Value) (r interface{}, err error) {
	input := reflect.Indirect(in)
	kind := input.Kind()

	switch {
	case kind == reflect.Invalid:
		r = nil
	case input.Type().Implements(reflect.TypeOf((*json.Marshaler)(nil)).Elem()):
		r, err = marshalJSONMarshaler(input)
	case kind == reflect.Interface:
		r, err = marshal(input.Elem())
	case basicType(kind):
		r = in.Interface()
	case kind == reflect.Struct:
		r, err = marshalStruct(input)
	case kind == reflect.Slice || kind == reflect.Array:
		r, err = marshalList(input)
	case kind == reflect.Map:
		r, err = marshalMap(input)
	default:
		err = newUnsupportedTypeError(input.Type())
	}

	return
}

func marshalList(in reflect.Value) (out []interface{}, err error) {
	if in.Type().Kind() == reflect.Slice && in.IsNil() {
		return out, nil
	}

	out = make([]interface{}, in.Len())
	for i := 0; i < in.Len(); i++ {
		r, err := marshal(in.Index(i))
		if err != nil {
			return nil, wrapErrorWithIndexContext(err, i, in.Type())
		}
		out[i] = r
	}

	return out, nil
}

func marshalMap(in reflect.Value) (out map[string]interface{}, err error) {
	if in.IsNil() {
		return out, nil
	}

	out = make(map[string]interface{})
	iter := in.MapRange()
	for iter.Next() {
		k := iter.Key()
		if k.Kind() != reflect.String {
			return nil, newUnsupportedKeyTypeError(in.Type())
		}

		r, err := marshal(iter.Value())
		if err != nil {
			return nil, wrapErrorWithKeyContext(err, k.String(), k.Type())
		}
		out[k.String()] = r
	}

	return out, nil
}

func marshalJSONMarshaler(in reflect.Value) (interface{}, error) {
	const method = "MarshalJSON"
	t := in.MethodByName(method).Call(nil)

	if err := checkForError(t[1]); err != nil {
		return nil, newForeignError(fmt.Sprintf("error from %s() call", method), err)
	}

	var r interface{}
	err := json.Unmarshal(t[0].Bytes(), &r)
	if err != nil {
		return nil, newForeignError(fmt.Sprintf(`error parsing %s() output "%s"`, method, t[0].Bytes()), err)
	}

	return r, nil
}

func shouldMarshal(p path.Path, v reflect.Value) bool {
	switch {
	case p.OmitAlways:
		return false
	case v.Type().Implements(reflect.TypeOf((*Omissible)(nil)).Elem()):
		return !v.MethodByName("OmitJSONry").Call(nil)[0].Bool()
	case p.OmitEmpty && isEmpty(v):
		return false
	default:
		return true
	}
}

func isEmpty(v reflect.Value) bool {
	k := v.Kind()
	switch {
	case k == reflect.Interface, k == reflect.Ptr:
		return v.IsZero() || v.IsNil()
	case k == reflect.String, k == reflect.Map, k == reflect.Slice, k == reflect.Array:
		return v.Len() == 0
	case basicType(k):
		return v.IsZero()
	default:
		return false
	}
}
