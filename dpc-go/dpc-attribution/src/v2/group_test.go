package v2

import (
	"context"
	"encoding/json"
	"github.com/CMSgov/dpc/attribution/model"
	"github.com/bxcodec/faker"
	"github.com/kinbiko/jsonassert"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockGrpRepo struct {
	mock.Mock
}

func (m *MockGrpRepo) Insert(ctx context.Context, body []byte) (*model.Group, error) {
	args := m.Called(ctx, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Group), args.Error(1)
}

type GroupServiceTestSuite struct {
	suite.Suite
	repo    *MockGrpRepo
	service *GroupService
}

func TestGroupServiceTestSuite(t *testing.T) {
	suite.Run(t, new(GroupServiceTestSuite))
}

func (suite *GroupServiceTestSuite) SetupTest() {
	suite.repo = &MockGrpRepo{}
	suite.service = NewGroupService(suite.repo)
}

func (suite *GroupServiceTestSuite) TestPost() {
	ja := jsonassert.New(suite.T())

	o := model.Group{}
	_ = faker.FakeData(&o)
	suite.repo.On("Insert", mock.Anything, mock.Anything).Return(&o, nil)

	req := httptest.NewRequest("POST", "http://example.com/foo", nil)

	w := httptest.NewRecorder()

	suite.service.Post(w, req)

	res := w.Result()

	assert.Equal(suite.T(), http.StatusOK, res.StatusCode)

	resp, _ := ioutil.ReadAll(res.Body)

	b, _ := json.Marshal(o)
	ja.Assertf(string(resp), string(b))
}

func (suite *GroupServiceTestSuite) TestPostRepoError() {
	ja := jsonassert.New(suite.T())

	suite.repo.On("Insert", mock.Anything, mock.Anything).Return(nil, errors.New("error"))

	req := httptest.NewRequest("POST", "http://example.com/foo", nil)

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
