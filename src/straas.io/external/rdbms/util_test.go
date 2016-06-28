package rdbms

import (
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/suite"
)

func TestUtil(t *testing.T) {
	suite.Run(t, new(UtilTestSuite))
}

type UtilTestSuite struct {
	suite.Suite
}

func (s *UtilTestSuite) TestUnique() {
	s.False(IsErrConstraintUnique(nil))
	s.False(IsErrConstraintUnique(fmt.Errorf("some error")))

	s.True(IsErrConstraintUnique(&mysql.MySQLError{
		Number: 1062,
	}))
	s.True(IsErrConstraintUnique(sqlite3.Error{
		ExtendedCode: sqlite3.ErrConstraintUnique,
	}))
}
