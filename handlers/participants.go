package handlers

import (
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

func joinHandler(app core.App) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		eventId := e.Request.PathValue("id")

		if _, err := app.FindRecordById("events", eventId); err != nil {
			return e.NotFoundError("event not found", nil)
		}

		var form struct {
			Name  string `form:"name"`
			Emoji string `form:"emoji"`
		}
		if err := e.BindBody(&form); err != nil {
			return e.BadRequestError("invalid form data", nil)
		}

		form.Name = strings.TrimSpace(form.Name)
		if form.Name == "" {
			return e.BadRequestError("name is required", nil)
		}
		if form.Emoji == "" {
			form.Emoji = "😊"
		}

		col, err := app.FindCollectionByNameOrId("participants")
		if err != nil {
			return err
		}
		token := newToken()
		p := core.NewRecord(col)
		p.Set("event_id", eventId)
		p.Set("name", form.Name)
		p.Set("emoji", form.Emoji)
		p.Set("token", token)
		if err := app.Save(p); err != nil {
			return err
		}

		setParticipantCookie(e, eventId, token)

		// Use HX-Redirect to do a full page refresh so the new cookie is active
		e.Response.Header().Set("HX-Redirect", "/events/"+eventId)
		return e.NoContent(http.StatusNoContent)
	}
}
