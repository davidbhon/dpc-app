package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/CMSgov/dpc/attribution/middleware"
	"github.com/bxcodec/faker/v3"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CMSgov/dpc/attribution/model"
	"github.com/kinbiko/jsonassert"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockImplementerRepo struct {
	mock.Mock
}

func (m *MockImplementerRepo) Insert(ctx context.Context, body []byte) (*model.Implementer, error) {
	args := m.Called(ctx, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Implementer), args.Error(1)
}
func (m *MockImplementerRepo) Update(ctx context.Context, id string, body []byte) (*model.Implementer, error) {
	args := m.Called(ctx, id, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Implementer), args.Error(1)
}
func (m *MockImplementerRepo) FindByID(ctx context.Context, id string) (*model.Implementer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Implementer), args.Error(1)
}

type ImplementerServiceTestSuite struct {
	suite.Suite
	repo    *MockImplementerRepo
	service *ImplementerService
}

func TestImplementerServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ImplementerServiceTestSuite))
}

func (suite *ImplementerServiceTestSuite) SetupTest() {
	suite.repo = &MockImplementerRepo{}
	suite.service = NewImplementerService(suite.repo)
}

func (suite *ImplementerServiceTestSuite) TestPost() {
	ja := jsonassert.New(suite.T())

	impl := model.Implementer{}
	err := faker.FakeData(&impl)
	if err != nil {
		fmt.Printf("ERR %v\n", err)
	}
	suite.repo.On("Insert", mock.Anything, mock.Anything).Return(&impl, nil)

	req := httptest.NewRequest(http.MethodPost, "http://example.com/foo", strings.NewReader(`{"name":"test-name"}`))

	w := httptest.NewRecorder()

	suite.service.Post(w, req)

	res := w.Result()

	assert.Equal(suite.T(), http.StatusOK, res.StatusCode)

	resp, _ := ioutil.ReadAll(res.Body)

	b, _ := json.Marshal(impl)
	ja.Assertf(string(resp), string(b))
}

func (suite *ImplementerServiceTestSuite) TestPut() {
	ja := jsonassert.New(suite.T())

	impl := model.Implementer{}
	err := faker.FakeData(&impl)
	if err != nil {
		fmt.Printf("ERR %v\n", err)
	}
	implUpdated := model.Implementer{}
	err = faker.FakeData(&impl)
	if err != nil {
		fmt.Printf("ERR %v\n", err)
	}
	suite.repo.On("FindByID", mock.Anything, mock.Anything).Return(&impl, nil)

	suite.repo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(&implUpdated, nil)

	//Send update request
	req := httptest.NewRequest(http.MethodPut, "http://example.com/Implementer/123456789", strings.NewReader(`{"name":"test-name"}`))
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.ContextKeyImplementer, "123456789")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	suite.service.Put(w, req)

	res := w.Result()
	assert.Equal(suite.T(), http.StatusOK, res.StatusCode)
	resp, _ := ioutil.ReadAll(res.Body)

	b, _ := json.Marshal(implUpdated)
	ja.Assertf(string(resp), string(b))
}

func (suite *ImplementerServiceTestSuite) TestUpdateWithoutName() {
	impl := model.Implementer{}
	err := faker.FakeData(&impl)
	if err != nil {
		fmt.Printf("ERR %v\n", err)
	}
	suite.repo.On("FindByID", mock.Anything, mock.Anything).Return(&impl, nil)

	//Send update request
	req := httptest.NewRequest(http.MethodPut, "http://example.com/Implementer/123456789", strings.NewReader(`{}`))
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.ContextKeyImplementer, "123456789")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	suite.service.Put(w, req)

	res := w.Result()
	assert.Equal(suite.T(), http.StatusUnprocessableEntity, res.StatusCode)
	resp, _ := ioutil.ReadAll(res.Body)

	assert.Contains(suite.T(), string(resp), "Missing name in body")
}

func (suite *ImplementerServiceTestSuite) TestUpdateNotFound() {
	impl := model.Implementer{}
	err := faker.FakeData(&impl)
	if err != nil {
		fmt.Printf("ERR %v\n", err)
	}
	suite.repo.On("FindByID", mock.Anything, mock.Anything).Return(nil, nil)

	//Send update request
	req := httptest.NewRequest(http.MethodPut, "http://example.com/Implementer/123456789", strings.NewReader(`{}`))
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.ContextKeyImplementer, "123456789")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	suite.service.Put(w, req)

	res := w.Result()
	assert.Equal(suite.T(), http.StatusNotFound, res.StatusCode)
	resp, _ := ioutil.ReadAll(res.Body)

	assert.Contains(suite.T(), string(resp), "Implementer not found")
}

func (suite *ImplementerServiceTestSuite) TestSaveRepoError() {
	ja := jsonassert.New(suite.T())

	suite.repo.On("Insert", mock.Anything, mock.Anything).Return(nil, errors.New("error"))

	req := httptest.NewRequest(http.MethodPost, "http://example.com/foo", strings.NewReader("{\"name\":\"test-name\"}"))

	w := httptest.NewRecorder()

	suite.service.Post(w, req)

	res := w.Result()

	assert.Equal(suite.T(), http.StatusUnprocessableEntity, res.StatusCode)

	resp, _ := ioutil.ReadAll(res.Body)

	ja.Assertf(string(resp), `
    {
        "error": "Unprocessable Entity",
        "message": "error",
        "statusCode": 422
    }`)
}

func (suite *ImplementerServiceTestSuite) TestGetNotImplemented() {
	ja := jsonassert.New(suite.T())

	impl := model.Implementer{}
	err := faker.FakeData(&impl)
	if err != nil {
		fmt.Printf("ERR %v\n", err)
	}
	suite.repo.On("FindByID", mock.Anything, mock.Anything).Return(&impl, nil)

	//Send update request
	req := httptest.NewRequest(http.MethodGet, "http://example.com/Implementer/123456789", nil)
	ctx := req.Context()
	ctx = context.WithValue(ctx, middleware.ContextKeyImplementer, "123456789")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	suite.service.Get(w, req)

	res := w.Result()
	assert.Equal(suite.T(), http.StatusOK, res.StatusCode)
	resp, _ := ioutil.ReadAll(res.Body)

	b, _ := json.Marshal(impl)
	ja.Assertf(string(resp), string(b))
}

func (suite *ImplementerServiceTestSuite) TestDeleteNotImplemented() {
	req := httptest.NewRequest(http.MethodDelete, "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	suite.service.Delete(w, req)
	res := w.Result()
	assert.Equal(suite.T(), http.StatusNotImplemented, res.StatusCode)
}

func (suite *ImplementerServiceTestSuite) TestExportNotImplemented() {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	suite.service.Export(w, req)
	res := w.Result()
	assert.Equal(suite.T(), http.StatusNotImplemented, res.StatusCode)
}
