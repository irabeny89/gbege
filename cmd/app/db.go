package main

import (
	"github.com/irabeny89/gosqlitex"
)

func dbConn() (*gosqlitex.DbClient, error) {
	db, err := gosqlitex.Open(new(gosqlitex.Config))
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
