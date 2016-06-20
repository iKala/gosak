package rest

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"straas.io/base/logger"
	"straas.io/base/metric"

	cmetric "github.com/csigo/metric"
	"github.com/stretchr/testify/suite"
)

var log = logger.Get()

func TestRestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(RestHandlerTestSuite))
}

type RestHandlerTestSuite struct {
	suite.Suite
	response *httptest.ResponseRecorder
}

func (suite *RestHandlerTestSuite) SetupTest() {
	suite.initResponse()
}

func (suite *RestHandlerTestSuite) TestSimpleRoute() {
	req, err := http.NewRequest("GET", "/healthcheck", nil)
	suite.Equal(nil, err)

	router := New(log, metric.New("test"))
	router.Route("GET", "/healthcheck", okEchoer)
	handler := router.GetHandler()
	handler.ServeHTTP(suite.response, req)

	body, err := ioutil.ReadAll(suite.response.Body)
	suite.Equal(nil, err)
	suite.Equal(http.StatusOK, suite.response.Code)
	suite.Equal("OK", string(body))
}

func (suite *RestHandlerTestSuite) Test404ForNonexistingRoute() {
	req, err := http.NewRequest("GET", "/thereIsNoSuchRoute", nil)
	suite.Equal(nil, err)

	router := New(log, metric.New("test"))
	router.Route("GET", "/ok", okEchoer)
	handler := router.GetHandler()
	handler.ServeHTTP(suite.response, req)

	_, err = ioutil.ReadAll(suite.response.Body)
	suite.Equal(nil, err)
	suite.Equal(http.StatusNotFound, suite.response.Code)
}

func (suite *RestHandlerTestSuite) TestSimpleMiddleware() {
	req, err := http.NewRequest("GET", "/ok", nil)
	suite.Equal(nil, err)

	router := New(log, metric.New("test"))
	router.Use(karaInserter)
	router.Route("GET", "/ok", okEchoer)
	handler := router.GetHandler()
	handler.ServeHTTP(suite.response, req)

	body, err := ioutil.ReadAll(suite.response.Body)
	suite.Equal(nil, err)
	suite.Equal(http.StatusOK, suite.response.Code)
	suite.Equal("karaOKe", string(body))
}

func (suite *RestHandlerTestSuite) TestMultipleRoutes() {
	router := New(log, metric.New("test"))
	router.Route("GET", "/ok", okEchoer)
	router.Route("GET", "/foo", fooEchoer)
	handler := router.GetHandler()

	req, err := http.NewRequest("GET", "/foo", nil)
	suite.Equal(nil, err)
	handler.ServeHTTP(suite.response, req)
	body, err := ioutil.ReadAll(suite.response.Body)
	suite.Equal(nil, err)
	suite.Equal(http.StatusOK, suite.response.Code)
	suite.Equal("foo", string(body))

	suite.initResponse()
	req, err = http.NewRequest("GET", "/ok", nil)
	suite.Equal(nil, err)
	handler.ServeHTTP(suite.response, req)
	body, err = ioutil.ReadAll(suite.response.Body)
	suite.Equal(nil, err)
	suite.Equal(http.StatusOK, suite.response.Code)
	suite.Equal("OK", string(body))
}

func (suite *RestHandlerTestSuite) TestHandlerReturnError() {
	req, err := http.NewRequest("GET", "/ok", nil)
	suite.Equal(nil, err)

	router := New(log, metric.New("test"))
	errorStr := "user sees this string so I can't say anything here"
	router.Route("GET", "/ok", func(rw http.ResponseWriter, req *http.Request) *Error {
		return &Error{
			Error:  errors.New(errorStr),
			Detail: "stack overflow!",
			Code:   http.StatusInternalServerError,
		}
	})
	handler := router.GetHandler()
	handler.ServeHTTP(suite.response, req)

	body, err := ioutil.ReadAll(suite.response.Body)
	suite.Equal(nil, err)
	suite.Equal(http.StatusInternalServerError, suite.response.Code)
	suite.Equal(errorStr+"\n", string(body))
}

func (suite *RestHandlerTestSuite) TestMiddlewareReturnError() {
	req, err := http.NewRequest("GET", "/ok", nil)
	suite.Equal(nil, err)

	router := New(log, metric.New("test"))
	errorStr := "user sees this string so I can't say anything here"
	router.Use(func(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		http.Error(rw, errorStr, http.StatusInternalServerError)
	})
	router.Route("GET", "/ok", okEchoer)
	handler := router.GetHandler()
	handler.ServeHTTP(suite.response, req)

	body, err := ioutil.ReadAll(suite.response.Body)
	suite.Equal(nil, err)
	suite.Equal(http.StatusInternalServerError, suite.response.Code)
	suite.Equal(errorStr+"\n", string(body))
}

func (suite *RestHandlerTestSuite) initResponse() {
	suite.response = httptest.NewRecorder()
	suite.response.Body = new(bytes.Buffer)
}

func karaInserter(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	fmt.Fprintf(rw, "kara")
	next(rw, req)
	fmt.Fprintf(rw, "e")
}

func okEchoer(rw http.ResponseWriter, req *http.Request) *Error {
	fmt.Fprintf(rw, "OK")
	return nil
}

func fooEchoer(rw http.ResponseWriter, req *http.Request) *Error {
	fmt.Fprintf(rw, "foo")
	return nil
}

// test middleware error blocks handler call
