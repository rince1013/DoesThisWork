package handlers

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func toggleVoteHandler(app core.App) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		eventId := e.Request.PathValue("id")
		dateId := e.Request.PathValue("dateId")

		event, err := app.FindRecordById("events", eventId)
		if err != nil {
			return e.NotFoundError("event not found", nil)
		}

		// must be a participant
		participantToken := getParticipantToken(e.Request, eventId)
		if participantToken == "" {
			return e.ForbiddenError("join the event first", nil)
		}
		participant, err := app.FindFirstRecordByFilter("participants",
			"event_id={:eid} && token={:tok}",
			dbx.Params{"eid": eventId, "tok": participantToken},
		)
		if err != nil {
			return e.ForbiddenError("participant not found", nil)
		}

		// toggle: remove if exists, add if not
		// validate the date belongs to this event first
		dateOpt, err := app.FindRecordById("date_options", dateId)
		if err != nil || dateOpt.GetString("event_id") != eventId {
			return e.NotFoundError("date not found", nil)
		}

		existing, err := app.FindFirstRecordByFilter("votes",
			"date_option_id={:did} && participant_id={:pid}",
			dbx.Params{"did": dateId, "pid": participant.Id},
		)
		if err == nil {
			if !existing.GetBool("preferred") {
				// available → preferred
				existing.Set("preferred", true)
				if err := app.Save(existing); err != nil {
					return err
				}
			} else {
				// preferred → remove
				if err := app.Delete(existing); err != nil {
					return err
				}
			}
		} else {
			// no vote → available
			col, err := app.FindCollectionByNameOrId("votes")
			if err != nil {
				return err
			}
			vote := core.NewRecord(col)
			vote.Set("date_option_id", dateId)
			vote.Set("participant_id", participant.Id)
			vote.Set("preferred", false)
			if err := app.Save(vote); err != nil {
				return err
			}
		}

		data, err := buildEventPageData(app, e, event)
		if err != nil {
			return err
		}
		return renderFragment(e, "results", data)
	}
}

func resultsHandler(app core.App) func(*core.RequestEvent) error {
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
		return renderFragment(e, "results", data)
	}
}

func lockDateHandler(app core.App) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		eventId := e.Request.PathValue("id")
		dateId := e.Request.PathValue("dateId")

		event, err := app.FindRecordById("events", eventId)
		if err != nil {
			return e.NotFoundError("event not found", nil)
		}

		// only creator may lock; validate date belongs to this event
		creatorToken := getCreatorToken(e.Request, eventId)
		if creatorToken == "" || creatorToken != event.GetString("creator_token") {
			return e.ForbiddenError("only the creator can lock a date", nil)
		}

		dateOpt, err := app.FindRecordById("date_options", dateId)
		if err != nil || dateOpt.GetString("event_id") != eventId {
			return e.NotFoundError("date not found", nil)
		}

		event.Set("locked_date_id", dateId)
		if err := app.Save(event); err != nil {
			return err
		}

		data, err := buildEventPageData(app, e, event)
		if err != nil {
			return err
		}
		return renderFragment(e, "results", data)
	}
}
