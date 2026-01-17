//go:build unit

package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	nyumpb2 "github.com/rkuprov/nyumspace/scratch/pkg/gen/nyumpb"
	"github.com/rkuprov/nyumspace/scratch/pkg/nyum"
)

// MockAdmin implements the admin.Server interface for testing
type MockAdmin struct {
	shouldReturnError bool
	errorMessage      string
}

// Ensure MockAdmin implements Server interface
var _ Server = (*MockAdmin)(nil)

// NewSuccessfulMock creates a MockAdmin that returns successful responses
func NewSuccessfulMock() *MockAdmin {
	return &MockAdmin{
		shouldReturnError: false,
	}
}

// NewErrorMock creates a MockAdmin that returns errors for all operations
func NewErrorMock(errorMessage string) *MockAdmin {
	if errorMessage == "" {
		errorMessage = "mock server error"
	}
	return &MockAdmin{
		shouldReturnError: true,
		errorMessage:      errorMessage,
	}
}

// GetUser returns a mock user or error based on configuration
func (m *MockAdmin) GetUser(ctx context.Context, req *nyum.UserRequest) (*nyum.UserResponse, error) {
	if m.shouldReturnError {
		return nil, errors.New(m.errorMessage)
	}

	if req.UserId == "" {
		return nil, errors.New("userID is required")
	}

	return &nyum.UserResponse{
		UserResponse: nyumpb2.UserResponse{
			UserId:   req.UserId,
			Username: "test_user",
			Email:    "test@example.com",
		},
	}, nil
}

// GetAllUsers returns mock users or error based on configuration
func (m *MockAdmin) GetAllUsers(ctx context.Context) ([]*nyum.UserResponse, error) {
	if m.shouldReturnError {
		return nil, errors.New(m.errorMessage)
	}

	return []*nyum.UserResponse{
		{
			UserResponse: nyumpb2.UserResponse{
				UserId:   "user1",
				Username: "test_user1",
				Email:    "test1@example.com",
			},
		},
		{
			UserResponse: nyumpb2.UserResponse{
				UserId:   "user2",
				Username: "test_user2",
				Email:    "test2@example.com",
			},
		},
	}, nil
}

// DeleteUser returns mock delete response or error based on configuration
func (m *MockAdmin) DeleteUser(ctx context.Context, req *nyum.UserDeleteRequest) (*nyum.UserDeleteResponse, error) {
	if m.shouldReturnError {
		return nil, errors.New(m.errorMessage)
	}

	if req.UserId == "" {
		return nil, errors.New("userID is required")
	}

	return &nyum.UserDeleteResponse{
		UserDeleteResponse: nyumpb2.UserDeleteResponse{
			Message: fmt.Sprintf("User with ID %s deleted successfully", req.UserId),
		},
	}, nil
}

// GetHome returns a mock home or error based on configuration
func (m *MockAdmin) GetHome(ctx context.Context, req *nyum.HomeRequest) (*nyum.HomeResponse, error) {
	if m.shouldReturnError {
		return nil, errors.New(m.errorMessage)
	}

	if req.HomeId == "" {
		return nil, errors.New("homeID is required")
	}

	now := time.Now()
	return &nyum.HomeResponse{
		HomeResponse: nyumpb2.HomeResponse{
			HomeId:          req.HomeId,
			OwnerId:         "owner1",
			Name:            "Test Home",
			StreetAddress_1: "123 Test St",
			StreetAddress_2: "Apt 1",
			City:            "Test City",
			State:           "TS",
			ZipCode:         "12345",
			Country:         "USA",
			Description:     "A test home",
			Tags:            []string{"test", "mock"},
			ImageUrl:        "https://example.com/image.jpg",
			CreatedAt:       now.Format(time.RFC3339),
			UpdatedAt:       now.Format(time.RFC3339),
		},
	}, nil
}

// GetAllHomes returns mock homes or error based on configuration
func (m *MockAdmin) GetAllHomes(ctx context.Context) ([]*nyum.HomeResponse, error) {
	if m.shouldReturnError {
		return nil, errors.New(m.errorMessage)
	}

	now := time.Now()
	return []*nyum.HomeResponse{
		{
			HomeResponse: nyumpb2.HomeResponse{
				HomeId:          "home1",
				OwnerId:         "owner1",
				Name:            "Test Home 1",
				StreetAddress_1: "123 Test St",
				City:            "Test City",
				State:           "TS",
				ZipCode:         "12345",
				Country:         "USA",
				CreatedAt:       now.Format(time.RFC3339),
				UpdatedAt:       now.Format(time.RFC3339),
			},
		},
		{
			HomeResponse: nyumpb2.HomeResponse{
				HomeId:          "home2",
				OwnerId:         "owner2",
				Name:            "Test Home 2",
				StreetAddress_1: "456 Test Ave",
				City:            "Test City",
				State:           "TS",
				ZipCode:         "12346",
				Country:         "USA",
				CreatedAt:       now.Format(time.RFC3339),
				UpdatedAt:       now.Format(time.RFC3339),
			},
		},
	}, nil
}

// DeleteHome returns mock delete response or error based on configuration
func (m *MockAdmin) DeleteHome(ctx context.Context, req *nyum.HomeDeleteRequest) (*nyum.HomeDeleteResponse, error) {
	if m.shouldReturnError {
		return nil, errors.New(m.errorMessage)
	}

	if req.HomeId == "" {
		return nil, errors.New("homeID is required")
	}

	return &nyum.HomeDeleteResponse{
		HomeDeleteResponse: nyumpb2.HomeDeleteResponse{
			Message: fmt.Sprintf("Home with ID %s deleted successfully", req.HomeId),
		},
	}, nil
}

// GetHomesForUser returns mock homes for a user or error based on configuration
func (m *MockAdmin) GetHomesForUser(ctx context.Context, req *nyum.UserHomesRequest) ([]*nyum.HomeResponse, error) {
	if m.shouldReturnError {
		return nil, errors.New(m.errorMessage)
	}

	if req.UserId == "" {
		return nil, errors.New("userID is required")
	}

	now := time.Now()
	return []*nyum.HomeResponse{
		{
			HomeResponse: nyumpb2.HomeResponse{
				HomeId:          "home1",
				OwnerId:         req.UserId,
				Name:            "User's Home 1",
				StreetAddress_1: "123 User St",
				City:            "User City",
				State:           "US",
				ZipCode:         "54321",
				Country:         "USA",
				CreatedAt:       now.Format(time.RFC3339),
				UpdatedAt:       now.Format(time.RFC3339),
			},
		},
	}, nil
}
