package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// -----------------------------------------------------------------
// Cookie helpers
// -----------------------------------------------------------------

func getCreatorToken(r *http.Request, eventId string) string {
	c, err := r.Cookie("dtw_c_" + eventId)
	if err != nil {
		return ""
	}
	return c.Value
}

func getParticipantToken(r *http.Request, eventId string) string {
	c, err := r.Cookie("dtw_p_" + eventId)
	if err != nil {
		return ""
	}
	return c.Value
}

func setCreatorCookie(e *core.RequestEvent, eventId, token string) {
	e.SetCookie(&http.Cookie{
		Name:     "dtw_c_" + eventId,
		Value:    token,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   isSecure(e.Request),
		MaxAge:   60 * 60 * 24 * 365,
	})
}

func setParticipantCookie(e *core.RequestEvent, eventId, token string) {
	e.SetCookie(&http.Cookie{
		Name:     "dtw_p_" + eventId,
		Value:    token,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   isSecure(e.Request),
		MaxAge:   60 * 60 * 24 * 365,
	})
}

// isSecure returns true when the request arrived over HTTPS (directly or via proxy).
func isSecure(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}

func newToken() string {
	return uuid.NewString()
}

// -----------------------------------------------------------------
// View model types
// -----------------------------------------------------------------

// DateOptionView is a date option enriched with vote data.
type DateOptionView struct {
	Id              string
	Date            string
	DateLabel       string // human-friendly format
	Voters          []ParticipantView
	PreferredVoters []ParticipantView
	VoteCount       int
	PreferredCount  int
	UserVoted       bool
	UserPreferred   bool
}

// ParticipantView is a slim participant representation.
type ParticipantView struct {
	Id    string
	Name  string
	Emoji string
}

// DefaultEmojis is the curated list shown in the join picker.
var DefaultEmojis = []string{
	"😊", "😎", "🥳", "🤩", "😄", "😂", "🤔", "😴",
	"🌟", "🔥", "🎉", "🍕", "🌮", "🍻", "☕", "🎸",
	"⚽", "🏄", "🐶", "🐱", "🦊", "🐻", "🌈", "🦋",
	"🌸", "🍀", "🚀", "💎", "🎯", "🦄",
}

// EventPageData is the full data bundle for the event page.
type EventPageData struct {
	EventId            string
	EventName          string
	EventDescription   string
	DateOptions        []DateOptionView
	Participants       []ParticipantView
	IsCreator          bool
	IsParticipant      bool
	CurrentParticipant ParticipantView
	IsLocked           bool
	LockedDate         *DateOptionView
	EmojiOptions       []string
}

// -----------------------------------------------------------------
// Query helpers
// -----------------------------------------------------------------

// buildEventPageData assembles EventPageData for a given event, resolving the
// current viewer's identity from their cookies.
func buildEventPageData(app core.App, e *core.RequestEvent, event *core.Record) (EventPageData, error) {
	eventId := event.Id

	// --- participants ---
	pRecords, err := app.FindRecordsByFilter("participants", "event_id={:eid}", "+name", 0, 0, dbx.Params{"eid": eventId})
	if err != nil {
		pRecords = []*core.Record{}
	}
	participantMap := make(map[string]ParticipantView, len(pRecords))
	participants := make([]ParticipantView, 0, len(pRecords))
	for _, p := range pRecords {
		pv := ParticipantView{
			Id:    p.Id,
			Name:  p.GetString("name"),
			Emoji: p.GetString("emoji"),
		}
		participantMap[p.Id] = pv
		participants = append(participants, pv)
	}

	// --- identity ---
	isCreator := false
	creatorToken := getCreatorToken(e.Request, eventId)
	if creatorToken != "" && creatorToken == event.GetString("creator_token") {
		isCreator = true
	}

	var currentParticipant ParticipantView
	isParticipant := false
	participantToken := getParticipantToken(e.Request, eventId)
	if participantToken != "" {
		p, err := app.FindFirstRecordByFilter("participants",
			"event_id={:eid} && token={:tok}",
			dbx.Params{"eid": eventId, "tok": participantToken},
		)
		if err == nil {
			currentParticipant = ParticipantView{
				Id:    p.Id,
				Name:  p.GetString("name"),
				Emoji: p.GetString("emoji"),
			}
			isParticipant = true
		}
	}

	// --- date options + votes ---
	dateRecords, err := app.FindRecordsByFilter("date_options", "event_id={:eid}", "+date", 0, 0, dbx.Params{"eid": eventId})
	if err != nil {
		dateRecords = []*core.Record{}
	}

	dateOptions := make([]DateOptionView, 0, len(dateRecords))
	for _, d := range dateRecords {
		votes, err := app.FindRecordsByFilter("votes", "date_option_id={:did}", "", 0, 0, dbx.Params{"did": d.Id})
		if err != nil {
			votes = []*core.Record{}
		}
		voters := make([]ParticipantView, 0, len(votes))
		preferredVoters := make([]ParticipantView, 0)
		userVoted := false
		userPreferred := false
		for _, v := range votes {
			pid := v.GetString("participant_id")
			preferred := v.GetBool("preferred")
			if pv, ok := participantMap[pid]; ok {
				voters = append(voters, pv)
				if preferred {
					preferredVoters = append(preferredVoters, pv)
				}
			}
			if isParticipant && pid == currentParticipant.Id {
				userVoted = true
				if preferred {
					userPreferred = true
				}
			}
		}
		dateOptions = append(dateOptions, DateOptionView{
			Id:              d.Id,
			Date:            d.GetString("date"),
			DateLabel:       formatDate(d.GetString("date")),
			Voters:          voters,
			PreferredVoters: preferredVoters,
			VoteCount:       len(votes),
			PreferredCount:  len(preferredVoters),
			UserVoted:       userVoted,
			UserPreferred:   userPreferred,
		})
	}

	// sort by vote count desc, then preferred count desc, then date asc
	sort.Slice(dateOptions, func(i, j int) bool {
		if dateOptions[i].VoteCount != dateOptions[j].VoteCount {
			return dateOptions[i].VoteCount > dateOptions[j].VoteCount
		}
		if dateOptions[i].PreferredCount != dateOptions[j].PreferredCount {
			return dateOptions[i].PreferredCount > dateOptions[j].PreferredCount
		}
		return dateOptions[i].Date < dateOptions[j].Date
	})

	// --- locked date ---
	isLocked := false
	var lockedDate *DateOptionView
	if lid := event.GetString("locked_date_id"); lid != "" {
		for i, d := range dateOptions {
			if d.Id == lid {
				isLocked = true
				lockedDate = &dateOptions[i]
				break
			}
		}
	}

	return EventPageData{
		EventId:            eventId,
		EventName:          event.GetString("name"),
		EventDescription:   event.GetString("description"),
		DateOptions:        dateOptions,
		Participants:       participants,
		IsCreator:          isCreator,
		IsParticipant:      isParticipant,
		CurrentParticipant: currentParticipant,
		IsLocked:           isLocked,
		LockedDate:         lockedDate,
		EmojiOptions:       DefaultEmojis,
	}, nil
}

// formatDate converts "2006-01-02" to "Wednesday, 5/20".
func formatDate(raw string) string {
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return raw
	}
	return fmt.Sprintf("%s, %d/%d", t.Format("Monday"), int(t.Month()), t.Day())
}
