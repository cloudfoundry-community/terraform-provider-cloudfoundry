package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/api/uaa"
	"code.cloudfoundry.org/cli/types"
)

func IntToNullInt(v int) types.NullInt {
	return types.NullInt{
		IsSet: true,
		Value: v,
	}
}

func IntToNullUint64Zero(v int) types.NullUint64 {
	if v < 0 {
		return types.NullUint64{
			IsSet: false,
		}
	}

	return types.NullUint64{
		IsSet: true,
		Value: uint64(v),
	}
}

func BoolToNullBool(b bool) types.NullBool {
	return types.NullBool{
		IsSet: true,
		Value: b,
	}
}

func IntToNullByteSize(v int) types.NullByteSizeInMb {
	if v < 0 {
		return types.NullByteSizeInMb{
			IsSet: false,
		}
	}
	return types.NullByteSizeInMb{
		IsSet: true,
		Value: uint64(v),
	}
}

func StringToFilteredString(val string) types.FilteredString {
	if val == "" {
		return types.FilteredString{
			IsSet: false,
		}
	}
	return types.FilteredString{
		IsSet: true,
		Value: val,
	}
}

func IntToNullByteSizeZero(v int) types.NullByteSizeInMb {
	if v <= 0 {
		return types.NullByteSizeInMb{
			IsSet: false,
		}
	}
	return types.NullByteSizeInMb{
		IsSet: true,
		Value: uint64(v),
	}
}

func NullByteSizeToInt(v types.NullByteSizeInMb) int {
	if !v.IsSet {
		return -1
	}
	return int(v.Value)
}

func UsersToIDs(users []ccv2.User) []interface{} {
	ids := make([]interface{}, len(users))
	for i, u := range users {
		ids[i] = u.GUID
	}
	return ids
}

func IsErrNotAuthorized(err error) bool {
	if _, ok := err.(ccerror.ForbiddenError); ok {
		return true
	}
	if httpErr, ok := err.(ccerror.RawHTTPStatusError); ok && httpErr.StatusCode == 403 {
		return true
	}
	if uaaErr, ok := err.(uaa.RawHTTPStatusError); ok && uaaErr.StatusCode == 403 {
		return true
	}
	return false
}

func IsErrNotFound(err error) bool {
	if httpErr, ok := err.(ccerror.RawHTTPStatusError); ok && httpErr.StatusCode == 404 {
		return true
	}
	if _, ok := err.(ccerror.ResourceNotFoundError); ok {
		return true
	}
	if uaaErr, ok := err.(uaa.RawHTTPStatusError); ok && uaaErr.StatusCode == 404 {
		return true
	}
	return false
}

func filterAppGuid(guid string) ccv2.Filter {
	return ccv2.FilterEqual(constant.AppGUIDFilter, guid)
}

func filterRouteGuid(guid string) ccv2.Filter {
	return ccv2.FilterEqual(constant.RouteGUIDFilter, guid)
}

func filterServiceInstanceGuid(guid string) ccv2.Filter {
	return ccv2.FilterEqual(constant.ServiceInstanceGUIDFilter, guid)
}
