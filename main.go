package main

import (
	"embed"
	"io/fs"
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/osutils"

	"doesthiswork/handlers"
	_ "doesthiswork/migrations"
)

//go:embed static
var staticFS embed.FS

func main() {
	app := pocketbase.New()

	// Register the migrate command.
	// Automigrate is enabled only during `go run` (dev) so that collection
	// changes in the Dashboard auto-generate migration files.
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: osutils.IsProbablyGoRun(),
	})

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		handlers.RegisterRoutes(e.Router, app)
		sub, err := fs.Sub(staticFS, "static")
		if err != nil {
			return err
		}
		e.Router.GET("/static/{path...}", apis.Static(sub, false)).Bind(apis.Gzip())
		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
