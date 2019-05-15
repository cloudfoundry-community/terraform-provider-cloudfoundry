package cloudfoundry

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/types"
)

func IntToNullInt(v int) types.NullInt {
	return types.NullInt{
		IsSet: true,
		Value: v,
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
