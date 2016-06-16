package rest

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"straas.io/base/logger"

	"github.com/stretchr/testify/suite"
)

var log = logger.Get()

func TestRestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(RestHandlerTestSuite))
}

type RestHandlerTestSuite struct {
	suite.Suite
}

func (suite *RestHandlerTestSuite) TestSimpleRoute() {
	response := httptest.NewRecorder()
	response.Body = new(bytes.Buffer)

	req, err := http.NewRequest("GET", "http://localhost/healthcheck", nil)
	suite.Equal(nil, err)

	router := New(log)
	router.Route("GET", "/healthcheck", func(w http.ResponseWriter, req *http.Request) *Error {
		fmt.Fprintf(w, "OK")
		return nil
	})
	handler := router.GetHandler()
	handler.ServeHTTP(response, req)

	body, err := ioutil.ReadAll(response.Body)
	suite.Equal(nil, err)
	suite.Equal(http.StatusOK, response.Code)
	suite.Equal("OK", string(body))
}

func (suite *RestHandlerTestSuite) Test404ForNonexistingRoute() {
	response := httptest.NewRecorder()
	response.Body = new(bytes.Buffer)

	req, err := http.NewRequest("GET", "http://localhost/thereIsNoSuchRoute", nil)
	suite.Equal(nil, err)

	router := New(log)
	handler := router.GetHandler()
	handler.ServeHTTP(response, req)

	_, err = ioutil.ReadAll(response.Body)
	suite.Equal(nil, err)
	suite.Equal(http.StatusNotFound, response.Code)
}

// TODO: add test for Use() and for error conditions
