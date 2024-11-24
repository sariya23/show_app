package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"show/internal/handler"
	"show/internal/lib/slogdiscard"
	"testing"

	"github.com/gin-gonic/gin"
	ssov1 "github.com/sariya23/sso_proto/gen/sso"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MockGrpcClient struct {
	mock.Mock
}

func (m *MockGrpcClient) Register(ctx context.Context, in *ssov1.RegisterRequest, opts ...grpc.CallOption) (*ssov1.RegisterResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*ssov1.RegisterResponse), args.Error(1)
}

func (m *MockGrpcClient) Login(ctx context.Context, in *ssov1.LoginRequest, opts ...grpc.CallOption) (*ssov1.LoginResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*ssov1.LoginResponse), args.Error(1)
}

func (m *MockGrpcClient) IsAdmin(ctx context.Context, in *ssov1.IsAdminRequest, opts ...grpc.CallOption) (*ssov1.IsAdminResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*ssov1.IsAdminResponse), args.Error(1)
}

type MockGrpcClientErr struct {
	mock.Mock
}

func (m *MockGrpcClientErr) Register(ctx context.Context, in *ssov1.RegisterRequest, opts ...grpc.CallOption) (*ssov1.RegisterResponse, error) {
	args := m.Called(ctx, in)
	return nil, args.Error(1)
}

func (m *MockGrpcClientErr) Login(ctx context.Context, in *ssov1.LoginRequest, opts ...grpc.CallOption) (*ssov1.LoginResponse, error) {
	args := m.Called(ctx, in)
	return nil, args.Error(1)
}

func (m *MockGrpcClientErr) IsAdmin(ctx context.Context, in *ssov1.IsAdminRequest, opts ...grpc.CallOption) (*ssov1.IsAdminResponse, error) {
	args := m.Called(ctx, in)
	return nil, args.Error(1)
}

// TestRegisterHandlerSuccess - проверяет, что
// пользователь может зарегестрироваться, если
// указал логин и пароль.
func TestRegisterHandlerSuccess(t *testing.T) {
	mockGrpcClient := new(MockGrpcClient)
	mockGrpcClient.On("Register", mock.Anything, &ssov1.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}).Return(&ssov1.RegisterResponse{
		UserId: 12345,
	}, nil)

	r := gin.Default()
	logger := slogdiscard.NewDiscardLogger()

	h := &handler.Handler{
		GrpcClient: mockGrpcClient,
		Log:        logger,
	}

	r.POST("/register", h.Register)

	registerReq := map[string]string{
		"login":    "test@example.com",
		"password": "password123",
	}
	jsonData, _ := json.Marshal(registerReq)

	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewReader(jsonData))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Equal(t, "register success", response["message"])
	require.Equal(t, 12345, int(response["user_id"].(float64)))
	mockGrpcClient.AssertExpectations(t)
}

// TestCannotRegisterWithoutEmail - проверяет, что
// если клиент не указал email, то ему вернется код 400
// и json с сообщением и деталями.
func TestCannotRegisterWithoutEmail(t *testing.T) {
	mockGrpcClient := new(MockGrpcClientErr)
	mockGrpcClient.On("Register", mock.Anything, &ssov1.RegisterRequest{
		Email:    "",
		Password: "password123",
	}).Return(nil, status.Error(codes.InvalidArgument, "email is invalid"))

	r := gin.Default()
	logger := slogdiscard.NewDiscardLogger()

	h := &handler.Handler{
		GrpcClient: mockGrpcClient,
		Log:        logger,
	}

	r.POST("/register", h.Register)

	registerReq := map[string]string{
		"login":    "",
		"password": "password123",
	}
	jsonData, _ := json.Marshal(registerReq)

	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewReader(jsonData))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	d, err := io.ReadAll(w.Body)
	require.NoError(t, err)
	err = json.Unmarshal(d, &response)
	require.NoError(t, err)
	require.Equal(t, "blank password or email", response["details"])
	require.Equal(t, "invalid arguments", response["message"])
	mockGrpcClient.AssertExpectations(t)
}

// TestCannotRegisterWithoutPassword - проверяет, что
// если пользователь не указал пароль ему вернется код 400
// и json с сообщением и деталями.
func TestCannotRegisterWithoutPassword(t *testing.T) {
	mockGrpcClient := new(MockGrpcClientErr)
	mockGrpcClient.On("Register", mock.Anything, &ssov1.RegisterRequest{
		Email:    "test@gmail.com",
		Password: "",
	}).Return(nil, status.Error(codes.InvalidArgument, "password is required"))

	r := gin.Default()
	logger := slogdiscard.NewDiscardLogger()

	h := &handler.Handler{
		GrpcClient: mockGrpcClient,
		Log:        logger,
	}

	r.POST("/register", h.Register)

	registerReq := map[string]string{
		"login":    "test@gmail.com",
		"password": "",
	}
	jsonData, _ := json.Marshal(registerReq)

	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewReader(jsonData))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	d, err := io.ReadAll(w.Body)
	require.NoError(t, err)
	err = json.Unmarshal(d, &response)
	require.NoError(t, err)
	require.Equal(t, "blank password or email", response["details"])
	require.Equal(t, "invalid arguments", response["message"])
	mockGrpcClient.AssertExpectations(t)
}

// TestCannotRegisterWithInvalidEmail проверяет, что
// если клиент указал невалидный email, то ему вернется код 400
// и json с сообщением и деталями.
func TestCannotRegisterWithInvalidEmail(t *testing.T) {
	mockGrpcClient := new(MockGrpcClientErr)
	mockGrpcClient.On("Register", mock.Anything, &ssov1.RegisterRequest{
		Email:    "qweqweqweqweqwe",
		Password: "qwerty",
	}).Return(nil, status.Error(codes.InvalidArgument, "password is required"))

	r := gin.Default()
	logger := slogdiscard.NewDiscardLogger()

	h := &handler.Handler{
		GrpcClient: mockGrpcClient,
		Log:        logger,
	}

	r.POST("/register", h.Register)

	registerReq := map[string]string{
		"login":    "qweqweqweqweqwe",
		"password": "qwerty",
	}
	jsonData, _ := json.Marshal(registerReq)

	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewReader(jsonData))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	d, err := io.ReadAll(w.Body)
	require.NoError(t, err)
	err = json.Unmarshal(d, &response)
	require.NoError(t, err)
	require.Equal(t, "blank password or email", response["details"])
	require.Equal(t, "invalid arguments", response["message"])
	mockGrpcClient.AssertExpectations(t)
}
