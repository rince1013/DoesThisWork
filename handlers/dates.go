package handlers

import (
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func addDateHandler(app core.App) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		eventId := e.Request.PathValue("id")

		event, err := app.FindRecordById("events", eventId)
		if err != nil {
			return e.NotFoundError("event not found", nil)
		}

		// must be a participant or the creator
		creatorToken := getCreatorToken(e.Request, eventId)
		isCreator := creatorToken != "" && creatorToken == event.GetString("creator_token")

		participantToken := getParticipantToken(e.Request, eventId)
		var participant *core.Record
		if participantToken != "" {
			participant, _ = app.FindFirstRecordByFilter("participants",
				"event_id={:eid} && token={:tok}",
				dbx.Params{"eid": eventId, "tok": participantToken},
			)
		}

		if !isCreator && participant == nil {
			return e.ForbiddenError("join the event first", nil)
		}

		var form struct {
			Date string `form:"date"`
		}
		if err := e.BindBody(&form); err != nil {
			return e.BadRequestError("invalid form data", nil)
		}
		form.Date = strings.TrimSpace(form.Date)
		if form.Date == "" {
			return e.BadRequestError("date is required", nil)
		}

		col, err := app.FindCollectionByNameOrId("date_options")
		if err != nil {
			return err
		}
		proposedBy := ""
		if participant != nil {
			proposedBy = participant.Id
		}
		opt := core.NewRecord(col)
		opt.Set("event_id", eventId)
		opt.Set("proposed_by", proposedBy)
		opt.Set("date", form.Date)
		if err := app.Save(opt); err != nil {
			return err
		}

		// return updated results fragment
		data, err := buildEventPageData(app, e, event)
		if err != nil {
			return err
		}
		return renderFragment(e, "results", data)
	}
}

func deleteDateHandler(app core.App) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		eventId := e.Request.PathValue("id")
		dateId := e.Request.PathValue("dateId")

		event, err := app.FindRecordById("events", eventId)
		if err != nil {
			return e.NotFoundError("event not found", nil)
		}

		dateOpt, err := app.FindRecordById("date_options", dateId)
		if err != nil {
			return e.NotFoundError("date not found", nil)
		}
		if dateOpt.GetString("event_id") != eventId {
			return e.NotFoundError("date not found", nil)
		}

		// only the proposer or the creator may delete
		creatorToken := getCreatorToken(e.Request, eventId)
		isCreator := creatorToken != "" && creatorToken == event.GetString("creator_token")

		participantToken := getParticipantToken(e.Request, eventId)
		var participant *core.Record
		if participantToken != "" {
			participant, _ = app.FindFirstRecordByFilter("participants",
				"event_id={:eid} && token={:tok}",
				dbx.Params{"eid": eventId, "tok": participantToken},
			)
		}

		isProposer := participant != nil && participant.Id == dateOpt.GetString("proposed_by")
		if !isCreator && !isProposer {
			return e.ForbiddenError("not allowed", nil)
		}

		// delete associated votes first
		votes, _ := app.FindRecordsByFilter("votes", "date_option_id={:did}", "", 0, 0, dbx.Params{"did": dateId})
		for _, v := range votes {
			_ = app.Delete(v)
		}

		if err := app.Delete(dateOpt); err != nil {
			return err
		}

		data, err := buildEventPageData(app, e, event)
		if err != nil {
			return err
		}
		return renderFragment(e, "results", data)
	}
}
