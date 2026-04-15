package handlers

import (
"github.com/pocketbase/pocketbase/core"
"github.com/pocketbase/pocketbase/tools/router"
)

// RegisterRoutes wires all app routes onto the PocketBase router.
func RegisterRoutes(r *router.Router[*core.RequestEvent], app core.App) {
r.GET("/", indexHandler(app))
r.POST("/events", createEventHandler(app))

// group all per-event routes under /events/{id}
events := r.Group("/events/{id}")
events.GET("", eventPageHandler(app))
events.POST("/join", joinHandler(app))
events.POST("/dates", addDateHandler(app))
events.DELETE("/dates/{dateId}", deleteDateHandler(app))
events.POST("/votes/{dateId}", toggleVoteHandler(app))
events.POST("/lock/{dateId}", lockDateHandler(app))

events.GET("/results", resultsHandler(app))
}
