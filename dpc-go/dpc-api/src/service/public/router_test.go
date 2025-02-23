package public

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CMSgov/dpc/api/client"
	"github.com/CMSgov/dpc/api/constants"
	"github.com/CMSgov/dpc/api/model"

	"github.com/CMSgov/dpc/api/apitest"
	"github.com/go-chi/chi/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockSsasClient struct {
	mock.Mock
}

func (mc *MockSsasClient) CreateSystem(ctx context.Context, request client.CreateSystemRequest) (client.CreateSystemResponse, error) {
	args := mc.Called(ctx, request)
	return args.Get(0).(client.CreateSystemResponse), args.Error(1)
}

func (mc *MockSsasClient) CreateGroup(ctx context.Context, request client.CreateGroupRequest) (client.CreateGroupResponse, error) {
	args := mc.Called(ctx, request)
	return args.Get(0).(client.CreateGroupResponse), args.Error(1)
}

func (mc *MockSsasClient) Authenticate(ctx context.Context, request []byte) ([]byte, error) {
	args := mc.Called(ctx, request)
	return args.Get(0).([]byte), args.Error(1)
}

func (mc *MockSsasClient) GetSystem(ctx context.Context, systemID string) (client.GetSystemResponse, error) {
	args := mc.Called(ctx, systemID)
	return args.Get(0).(client.GetSystemResponse), args.Error(1)
}

func (mc *MockSsasClient) CreateToken(ctx context.Context, systemID string, label string) (string, error) {
	args := mc.Called(ctx, systemID, label)
	return args.Get(0).(string), args.Error(1)
}

func (mc *MockSsasClient) DeleteToken(ctx context.Context, systemID string, tokenID string) error {
	args := mc.Called(ctx, systemID, tokenID)
	return args.Error(0)
}

func (mc *MockSsasClient) AddPublicKey(ctx context.Context, systemID string, request model.ProxyPublicKeyRequest) (map[string]string, error) {
	args := mc.Called(ctx, systemID, request)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (mc *MockSsasClient) DeletePublicKey(ctx context.Context, systemID string, keyID string) error {
	args := mc.Called(ctx, systemID, keyID)
	return args.Error(0)
}

func (mc *MockSsasClient) GetOrgIDFromToken(ctx context.Context, token string) (string, error) {
	args := mc.Called(ctx, token)
	return args.Get(0).(string), args.Error(1)
}

func (mc *MockSsasClient) ValidateToken(ctx context.Context, reqBytes []byte) ([]byte, error) {
	args := mc.Called(ctx, reqBytes)
	return args.Get(0).([]byte), args.Error(1)
}

type MockController struct {
	mock.Mock
}

func (c *MockController) Update(w http.ResponseWriter, r *http.Request) {
	c.Called(w, r)
}

func (c *MockController) Read(w http.ResponseWriter, r *http.Request) {
	c.Called(w, r)
}

func (c *MockController) Create(w http.ResponseWriter, r *http.Request) {
	c.Called(w, r)
}

func (c *MockController) Delete(w http.ResponseWriter, r *http.Request) {
	c.Called(w, r)
}

func (c *MockController) Export(w http.ResponseWriter, r *http.Request) {
	c.Called(w, r)
}

type MockFileController struct {
	mock.Mock
}

func (mfc *MockFileController) GetFile(w http.ResponseWriter, r *http.Request) {
	mfc.Called(w, r)
}

type MockJobController struct {
	mock.Mock
}

func (mjc *MockJobController) Status(w http.ResponseWriter, r *http.Request) {
	mjc.Called(w, r)
}

type MockSsasController struct {
	mock.Mock
}

func (mjc *MockSsasController) GetSystem(w http.ResponseWriter, r *http.Request) {
	mjc.Called(w, r)
}

func (mjc *MockSsasController) CreateSystem(w http.ResponseWriter, r *http.Request) {
	mjc.Called(w, r)
}

func (mjc *MockSsasController) GetAuthToken(w http.ResponseWriter, r *http.Request) {
	mjc.Called(w, r)
}

func (mjc *MockSsasController) CreateToken(w http.ResponseWriter, r *http.Request) {
	mjc.Called(w, r)
}

func (mjc *MockSsasController) DeleteToken(w http.ResponseWriter, r *http.Request) {
	mjc.Called(w, r)
}

func (mjc *MockSsasController) AddKey(w http.ResponseWriter, r *http.Request) {
	mjc.Called(w, r)
}

func (mjc *MockSsasController) DeleteKey(w http.ResponseWriter, r *http.Request) {
	mjc.Called(w, r)
}

func (mjc *MockSsasController) ValidateToken(w http.ResponseWriter, r *http.Request) {
	mjc.Called(w, r)
}

type MockExportController struct {
	mock.Mock
}

func (mec *MockExportController) Export(w http.ResponseWriter, r *http.Request) {
	mec.Called(w, r)
}

type RouterTestSuite struct {
	suite.Suite
	router         http.Handler
	mockOrg        *MockController
	mockHealth     *MockController
	mockMeta       *MockController
	mockGroup      *MockController
	mockData       *MockFileController
	mockJob        *MockJobController
	mockSsas       *MockSsasController
	mockSassClient *MockSsasClient
	mockPatient    *MockExportController
}

func (suite *RouterTestSuite) SetupTest() {
	suite.mockOrg = &MockController{}
	suite.mockMeta = &MockController{}
	suite.mockHealth = &MockController{}
	suite.mockGroup = &MockController{}
	suite.mockData = &MockFileController{}
	suite.mockJob = &MockJobController{}
	suite.mockSsas = &MockSsasController{}
	suite.mockSassClient = &MockSsasClient{}
	suite.mockPatient = &MockExportController{}

	c := controllers{
		Org:      suite.mockOrg,
		Metadata: suite.mockMeta,
		Health:   suite.mockHealth,
		Group:    suite.mockGroup,
		Data:     suite.mockData,
		Job:      suite.mockJob,
		Ssas:     suite.mockSsas,
		Patient:  suite.mockPatient,
	}

	suite.router = buildPublicRoutes(c, suite.mockSassClient)
}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}

func (suite *RouterTestSuite) TestMetadataRoute() {
	suite.mockMeta.On("Read", mock.Anything, mock.Anything).Once()

	ts := httptest.NewServer(suite.router)

	res, _ := http.Get(fmt.Sprintf("%s/%s", ts.URL, "api/v2/metadata"))

	assert.Equal(suite.T(), "application/fhir+json; charset=UTF-8", res.Header.Get("Content-Type"))
	assert.Equal(suite.T(), http.StatusOK, res.StatusCode)
}

func (suite *RouterTestSuite) TestHealthRoute() {
	suite.mockHealth.On("Read", mock.Anything, mock.Anything).Once()

	ts := httptest.NewServer(suite.router)

	res, _ := http.Get(fmt.Sprintf("%s/%s", ts.URL, "api/v2/_health"))

	assert.Equal(suite.T(), "application/fhir+json; charset=UTF-8", res.Header.Get("Content-Type"))
	assert.Equal(suite.T(), http.StatusOK, res.StatusCode)
}

func (suite *RouterTestSuite) TestGroupExportRoute() {
	var capturedRequestID string
	suite.mockGroup.On("Export", mock.Anything, mock.Anything).Once().Run(func(arg mock.Arguments) {
		r := arg.Get(1).(*http.Request)
		capturedRequestID = r.Header.Get(middleware.RequestIDHeader)
		w := arg.Get(0).(http.ResponseWriter)
		w.WriteHeader(http.StatusAccepted)
	})

	suite.mockSassClient.On("GetOrgIDFromToken", mock.Anything, mock.Anything).Return("12345", nil)

	ts := httptest.NewServer(suite.router)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", ts.URL, "api/v2/Group/9876/$export"), nil)
	req.Header.Add("Authorization", "Bearer hello")

	req.Header.Set("Content-Type", "application/fhir+json")
	req.Header.Set("Prefer", "respond-async")
	req.Header.Set(middleware.RequestIDHeader, "54321")
	req.Header.Set(constants.OrgHeader, "12345")
	res, _ := http.DefaultClient.Do(req)

	b, _ := ioutil.ReadAll(res.Body)
	var v map[string]interface{}
	_ = json.Unmarshal(b, &v)

	assert.Equal(suite.T(), "application/fhir+json; charset=UTF-8", res.Header.Get("Content-Type"))
	assert.NotNil(suite.T(), res.Header.Get("Content-Location"))
	assert.Equal(suite.T(), http.StatusAccepted, res.StatusCode)
	assert.Equal(suite.T(), "54321", capturedRequestID)
	assert.Nil(suite.T(), v)
}

func (suite *RouterTestSuite) TestOrganizationGetRoutes() {
	var capturedRequestID string
	suite.mockOrg.On("Read", mock.Anything, mock.Anything).Once().Run(func(arg mock.Arguments) {
		r := arg.Get(1).(*http.Request)
		capturedRequestID = r.Header.Get(middleware.RequestIDHeader)
		w := arg.Get(0).(http.ResponseWriter)
		_, _ = w.Write(apitest.AttributionOrgResponse())
	})

	suite.mockSassClient.On("GetOrgIDFromToken", mock.Anything, mock.Anything).Return("12345", nil)

	ts := httptest.NewServer(suite.router)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", ts.URL, "api/v2/Organization/12345"), nil)
	req.Header.Add("Authorization", "Bearer hello")

	req.Header.Set(middleware.RequestIDHeader, "54321")
	res, _ := http.DefaultClient.Do(req)

	b, _ := ioutil.ReadAll(res.Body)
	var v map[string]interface{}
	_ = json.Unmarshal(b, &v)

	assert.Equal(suite.T(), "application/fhir+json; charset=UTF-8", res.Header.Get("Content-Type"))
	assert.Equal(suite.T(), http.StatusOK, res.StatusCode)
	assert.Equal(suite.T(), "54321", capturedRequestID)
	assert.NotNil(suite.T(), v)
	assert.NotContains(suite.T(), v, "info")
	assert.Contains(suite.T(), v, "resourceType")
	assert.Equal(suite.T(), v["resourceType"], "Organization")
}

func (suite *RouterTestSuite) TestGetAuthTokenProxyRoute() {
	suite.mockSsas.On("GetAuthToken", mock.Anything, mock.Anything).Once().Run(func(arg mock.Arguments) {
		w := arg.Get(0).(http.ResponseWriter)
		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(suite.router)

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/%s", ts.URL, "api/v2/Token/auth"), strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	fmt.Println(err)
	assert.Equal(suite.T(), http.StatusOK, res.StatusCode)
}
