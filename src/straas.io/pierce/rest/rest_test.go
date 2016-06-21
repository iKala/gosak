package rest

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"straas.io/base/logger"
	"straas.io/base/metric"

	"github.com/stretchr/testify/suite"
)

var log = logger.Get()

func TestRestTestSuite(t *testing.T) {
	suite.Run(t, new(RestTestSuite))
}

type RestTestSuite struct {
	suite.Suite
}

func (suite *RestTestSuite) TestHealthCheck() {
	response := httptest.NewRecorder()
	response.Body = new(bytes.Buffer)

	req, err := http.NewRequest("GET", "/healthcheck", nil)
	suite.Equal(nil, err)

	handler := BuildHTTPHandler(log, metric.New("test"))
	handler.ServeHTTP(response, req)

	body, err := ioutil.ReadAll(response.Body)
	suite.Equal(nil, err)
	suite.Equal(http.StatusOK, response.Code)
	suite.Equal("OK", string(body))
}
