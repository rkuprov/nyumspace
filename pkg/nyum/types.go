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
	UserID string
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

type Room struct {
	nyumpb.Room
}
type Appliance struct {
	nyumpb.Appliance
}

type HouseAppliance struct {
	nyumpb.ApplianceMetadata
	Appliance nyumpb.Appliance `json:"appliance"`
}
type Code struct {
	nyumpb.Code
}

type HomeCreationRequest struct {
	UserID string
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
	UserID string
	HomeID string
	nyumpb.HomeUpdateRequest
}
type HomeUpdateResponse struct {
	nyumpb.HomeUpdateResponse
}

type HomeDeleteRequest struct {
	UserID string
	nyumpb.HomeDeleteRequest
}
type HomeDeleteResponse struct {
	nyumpb.HomeDeleteResponse
}

type UserHomesRequest struct {
	UserId string
}

type ImageCreateResponse struct {
	HomeID  string `json:"home_id"`
	ImageID string `json:"image_id"`
}
