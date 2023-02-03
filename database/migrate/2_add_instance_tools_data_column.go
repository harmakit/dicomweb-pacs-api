package migrate

import (
	"fmt"
	"github.com/go-pg/migrations"
)

const query = `
ALTER TABLE instance
ADD COLUMN tools_data jsonb NOT NULL DEFAULT '{}'::jsonb;
`

const rollbackQuery = `
ALTER TABLE instance
DROP COLUMN tools_data;
`

func init() {
	up := []string{
		query,
	}

	down := []string{
		rollbackQuery,
	}

	migrations.Register(func(db migrations.DB) error {
		fmt.Println("run migration")
		for _, q := range up {
			_, err := db.Exec(q)
			if err != nil {
				return err
			}
		}
		return nil
	}, func(db migrations.DB) error {
		fmt.Println("rollback migration")
		for _, q := range down {
			_, err := db.Exec(q)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
