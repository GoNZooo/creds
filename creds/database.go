package creds

import (
	"fmt"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

func connectToDatabase(options DatabaseOptions) *pg.DB {
	address := fmt.Sprintf("%s:%d", options.Host, options.Port)

	return pg.Connect(&pg.Options{
		Addr:     address,
		Database: options.Database,
		User:     options.User,
		Password: options.Password,
	})
}

func createSchema(db *pg.DB) error {
	models := []interface{}{(*User)(nil), (*Token)(nil)}

	for _, m := range models {
		err := db.Model(m).CreateTable(&orm.CreateTableOptions{Temp: false, IfNotExists: true})
		if err != nil {
			return err
		}
	}

	return nil
}
