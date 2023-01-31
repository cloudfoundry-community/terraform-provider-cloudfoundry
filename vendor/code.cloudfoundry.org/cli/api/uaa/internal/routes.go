package internal

import (
	"net/http"
)

const (
	GetSSHPasscodeRequest    = "GetSSHPasscode"
	GetGroupsRequest         = "GetGroups"
	PostOAuthTokenRequest    = "PostOAuthToken"
	PostUserRequest          = "PostUser"
	DeleteUserRequest        = "DeleteUser"
	GetUserRequest           = "GetUser"
	GetUsersRequest          = "GetUsers"
	PutUserRequest           = "PutUserRequest"
	PutUserPasswordRequest   = "PutUserPassword"
	PostGroupMemberRequest   = "PostGroupMember"
	DeleteGroupMemberRequest = "DeleteGroupMember"
)

// APIRoutes is a list of routes used by the router to construct request URLs.
var APIRoutes = []Route{
	{Path: "/Groups", Method: http.MethodGet, Name: GetGroupsRequest, Resource: UAAResource},
	{Path: "/Groups/:group_guid/members", Method: http.MethodPost, Name: PostGroupMemberRequest, Resource: UAAResource},
	{Path: "/Groups/:group_guid/members/:user_guid", Method: http.MethodDelete, Name: DeleteGroupMemberRequest, Resource: UAAResource},
	{Path: "/Users", Method: http.MethodPost, Name: PostUserRequest, Resource: UAAResource},
	{Path: "/Users", Method: http.MethodGet, Name: GetUsersRequest, Resource: UAAResource},
	{Path: "/Users/:user_guid", Method: http.MethodDelete, Name: DeleteUserRequest, Resource: UAAResource},
	{Path: "/Users/:user_guid", Method: http.MethodGet, Name: GetUserRequest, Resource: UAAResource},
	{Path: "/Users/:user_guid", Method: http.MethodPut, Name: PutUserRequest, Resource: UAAResource},
	{Path: "/Users/:user_guid/password", Method: http.MethodPut, Name: PutUserPasswordRequest, Resource: UAAResource},
	{Path: "/oauth/authorize", Method: http.MethodGet, Name: GetSSHPasscodeRequest, Resource: UAAResource},
	{Path: "/oauth/token", Method: http.MethodPost, Name: PostOAuthTokenRequest, Resource: AuthorizationResource},
}
