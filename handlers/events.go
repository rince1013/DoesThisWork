package handlers

import (
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

func indexHandler(app core.App) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		return renderPage(e, "index", nil)
	}
}

func createEventHandler(app core.App) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var form struct {
			Name        string   `form:"name"`
			Description string   `form:"description"`
			Dates       []string `form:"dates"`
		}
		if err := e.BindBody(&form); err != nil {
			return e.BadRequestError("invalid form data", nil)
		}

		form.Name = strings.TrimSpace(form.Name)
		if form.Name == "" {
			return e.BadRequestError("event name is required", nil)
		}

		// filter empty date strings
		var dates []string
		for _, d := range form.Dates {
			if d = strings.TrimSpace(d); d != "" {
				dates = append(dates, d)
			}
		}
		if len(dates) == 0 {
			return e.BadRequestError("at least one date is required", nil)
		}

		// create event
		eventCol, err := app.FindCollectionByNameOrId("events")
		if err != nil {
			return err
		}
		creatorToken := newToken()
		event := core.NewRecord(eventCol)
		event.Set("name", form.Name)
		event.Set("description", form.Description)
		event.Set("creator_token", creatorToken)
		if err := app.Save(event); err != nil {
			return err
		}

		// create date options
		dateCol, err := app.FindCollectionByNameOrId("date_options")
		if err != nil {
			return err
		}
		for _, d := range dates {
			opt := core.NewRecord(dateCol)
			opt.Set("event_id", event.Id)
			opt.Set("proposed_by", "")
			opt.Set("date", d)
			if err := app.Save(opt); err != nil {
				return err
			}
		}

		setCreatorCookie(e, event.Id, creatorToken)
		return e.Redirect(http.StatusSeeOther, "/events/"+event.Id)
	}
}

func eventPageHandler(app core.App) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		eventId := e.Request.PathValue("id")
		event, err := app.FindRecordById("events", eventId)
		if err != nil {
			return e.NotFoundError("event not found", nil)
		}

		data, err := buildEventPageData(app, e, event)
		if err != nil {
			return err
		}

		return renderPage(e, "event", data)
	}
}
