package nyum

import (
	nyumpb2 "github.com/rkuprov/nyumspace/scratch/pkg/gen/nyumpb"
)

// User types

type UserRegistrationRequest struct {
	nyumpb2.UserRegistrationRequest
}
type UserRegistrationResponse struct {
	nyumpb2.UserRegistrationResponse
}

type UserRequest struct {
	nyumpb2.UserRequest
}
type UserResponse struct {
	nyumpb2.UserResponse
}

type UserUpdateRequest struct {
	UserID string
	nyumpb2.UserUpdateRequest
}
type UserUpdateResponse struct {
	nyumpb2.UserUpdateResponse
}

type UserDeleteRequest struct {
	nyumpb2.UserDeleteRequest
}
type UserDeleteResponse struct {
	nyumpb2.UserDeleteResponse
}

type UserLoginRequest struct {
	nyumpb2.UserLoginRequest
}
type UserLoginResponse struct {
	nyumpb2.UserLoginResponse
}

type UserLogoutRequest struct {
	nyumpb2.UserLogoutRequest
}
type UserLogoutResponse struct {
	nyumpb2.UserLogoutResponse
}

type Room struct {
	nyumpb2.Room
}
type Appliance struct {
	nyumpb2.Appliance
}

type HouseAppliance struct {
	nyumpb2.ApplianceMetadata
	Appliance nyumpb2.Appliance `json:"appliance"`
}
type Code struct {
	nyumpb2.Code
}

type HomeCreationRequest struct {
	UserID string
	nyumpb2.HomeCreationRequest
}
type HomeCreationResponse struct {
	nyumpb2.HomeCreationResponse
}

type HomeRequest struct {
	nyumpb2.HomeRequest
}
type HomeResponse struct {
	nyumpb2.HomeResponse
}

type HomeUpdateRequest struct {
	UserID string
	HomeID string
	nyumpb2.HomeUpdateRequest
}
type HomeUpdateResponse struct {
	nyumpb2.HomeUpdateResponse
}

type HomeDeleteRequest struct {
	UserID string
	nyumpb2.HomeDeleteRequest
}
type HomeDeleteResponse struct {
	nyumpb2.HomeDeleteResponse
}

type UserHomesRequest struct {
	UserId string
}

type ImageCreateResponse struct {
	HomeID  string `json:"home_id"`
	ImageID string `json:"image_id"`
}
