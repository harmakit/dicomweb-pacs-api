package migrate

import (
	"fmt"

	"github.com/go-pg/migrations"
)

const studyTable = `
CREATE TABLE study (
id serial NOT NULL,
updated_at timestamp with time zone NOT NULL DEFAULT current_timestamp,
patient text NOT NULL,
PRIMARY KEY (id)
)`

func init() {
	up := []string{
		studyTable,
	}

	down := []string{
		`DROP TABLE study`,
	}

	migrations.Register(func(db migrations.DB) error {
		fmt.Println("create study table")
		for _, q := range up {
			_, err := db.Exec(q)
			if err != nil {
				return err
			}
		}
		return nil
	}, func(db migrations.DB) error {
		fmt.Println("drop study table")
		for _, q := range down {
			_, err := db.Exec(q)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
