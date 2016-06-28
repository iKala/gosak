package mysql

import (
	"errors"
	"flag"

	// register mysql driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	"straas.io/service/common"
)

func init() {
	common.Register(&service{})
}

type service struct {
	connStr string
}

func (s *service) Type() common.ServiceType {
	return common.MySQL
}

func (s *service) AddFlags() {
	flag.StringVar(&s.connStr, "common.mysql_connstr", "", "mysql connection string")
}

func (s *service) New(get common.ServiceGetter) (interface{}, error) {
	if s.connStr == "" {
		return nil, errors.New("empty mysql connection string")
	}

	db, err := gorm.Open("mysql", s.connStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (s *service) Dependencies() []common.ServiceType {
	return nil
}
