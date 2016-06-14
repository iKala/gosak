package rest_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"straas.io/pierce/rest"

	"github.com/codegangsta/negroni"
	"github.com/stretchr/testify/suite"
)

var n *negroni.Negroni
var tested rest.Handler

func TestRestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(RestHandlerTestSuite))
}

type RestHandlerTestSuite struct {
	suite.Suite
}

func (suite *RestHandlerTestSuite) SetupSuite() {
	// run test http server
	n = negroni.New()
	router := rest.BuildRouter(tested)
	n.UseHandler(router)
}

func (suite *RestHandlerTestSuite) TestHealthCheck() {
	response := httptest.NewRecorder()
	response.Body = new(bytes.Buffer)

	req, err := http.NewRequest("GET", "http://localhost/healthcheck", nil)
	suite.Equal(nil, err)
	n.ServeHTTP(response, req)

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
	n.ServeHTTP(response, req)

	_, err = ioutil.ReadAll(response.Body)
	suite.Equal(nil, err)
	suite.Equal(http.StatusNotFound, response.Code)
}
