package rdbms

import (
	"github.com/go-sql-driver/mysql"
	sqlite3 "github.com/mattn/go-sqlite3"
)

// IsErrConstraintUnique checks if violate unique constrain
// only supports mysql and sqlite3
func IsErrConstraintUnique(err error) bool {
	if err == nil {
		return false
	}
	// type assertion
	switch v := err.(type) {
	// mysql
	case *mysql.MySQLError:
		return v.Number == 1062

	// sqlite
	case sqlite3.Error:
		return v.ExtendedCode == sqlite3.ErrConstraintUnique

	default:
		return false
	}
}
