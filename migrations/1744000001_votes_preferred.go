package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("votes")
		if err != nil {
			return err
		}
		col.Fields.Add(&core.BoolField{Name: "preferred"})
		return app.Save(col)
	}, func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("votes")
		if err != nil {
			return err
		}
		col.Fields.RemoveByName("preferred")
		return app.Save(col)
	})
}
