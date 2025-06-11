package nyum

import "github.com/rkuprov/nyumspace/pkg/gen/nyumpb"

// User types

type UserRegistrationRequest struct {
	nyumpb.UserRegistrationRequest
}
type UserRegistrationResponse struct {
	nyumpb.UserRegistrationResponse
}

type UserRequest struct {
	nyumpb.UserRequest
}
type UserResponse struct {
	nyumpb.UserResponse
}

type UserUpdateRequest struct {
	nyumpb.UserUpdateRequest
}
type UserUpdateResponse struct {
	nyumpb.UserUpdateResponse
}

type UserDeleteRequest struct {
	nyumpb.UserDeleteRequest
}
type UserDeleteResponse struct {
	nyumpb.UserDeleteResponse
}

type UserLoginRequest struct {
	nyumpb.UserLoginRequest
}
type UserLoginResponse struct {
	nyumpb.UserLoginResponse
}

type UserLogoutRequest struct {
	nyumpb.UserLogoutRequest
}
type UserLogoutResponse struct {
	nyumpb.UserLogoutResponse
}

// Home types

type Room struct {
	nyumpb.Room
}
type Appliance struct {
	nyumpb.Appliance
}
type Code struct {
	nyumpb.Code
}

type HomeCreationRequest struct {
	nyumpb.HomeCreationRequest
}
type HomeCreationResponse struct {
	nyumpb.HomeCreationResponse
}

type HomeRequest struct {
	nyumpb.HomeRequest
}
type HomeResponse struct {
	nyumpb.HomeResponse
}

type HomeUpdateRequest struct {
	nyumpb.HomeUpdateRequest
}
type HomeUpdateResponse struct {
	nyumpb.HomeUpdateResponse
}

type HomeDeleteRequest struct {
	nyumpb.HomeDeleteRequest
}
type HomeDeleteResponse struct {
	nyumpb.HomeDeleteResponse
}
