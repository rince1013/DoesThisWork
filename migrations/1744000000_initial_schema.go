package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		return createInitialCollections(app)
	}, func(app core.App) error {
		return dropInitialCollections(app)
	})
}

func createInitialCollections(app core.App) error {
	// events
	events := core.NewBaseCollection("events")
	events.ListRule = ptr("")
	events.ViewRule = ptr("")
	events.Fields.Add(
		&core.TextField{Name: "name", Required: true},
		&core.TextField{Name: "description"},
		&core.TextField{Name: "creator_token", Required: true, Hidden: true},
		&core.TextField{Name: "locked_date_id"},
	)
	if err := app.Save(events); err != nil {
		return err
	}

	// participants
	participants := core.NewBaseCollection("participants")
	participants.ListRule = ptr("")
	participants.ViewRule = ptr("")
	participants.Fields.Add(
		&core.TextField{Name: "event_id", Required: true},
		&core.TextField{Name: "name", Required: true},
		&core.TextField{Name: "emoji"},
		&core.TextField{Name: "token", Required: true, Hidden: true},
	)
	if err := app.Save(participants); err != nil {
		return err
	}

	// date_options
	dateOptions := core.NewBaseCollection("date_options")
	dateOptions.ListRule = ptr("")
	dateOptions.ViewRule = ptr("")
	dateOptions.Fields.Add(
		&core.TextField{Name: "event_id", Required: true},
		&core.TextField{Name: "proposed_by"},
		&core.TextField{Name: "date", Required: true},
	)
	if err := app.Save(dateOptions); err != nil {
		return err
	}

	// votes — unique constraint prevents duplicate votes from race conditions
	votes := core.NewBaseCollection("votes")
	votes.ListRule = ptr("")
	votes.ViewRule = ptr("")
	votes.Fields.Add(
		&core.TextField{Name: "date_option_id", Required: true},
		&core.TextField{Name: "participant_id", Required: true},
	)
	votes.Indexes = append(votes.Indexes, "CREATE UNIQUE INDEX idx_votes_unique ON votes (date_option_id, participant_id)")
	return app.Save(votes)
}

// ptr returns a pointer to s, used for nullable rule fields.
func ptr(s string) *string { return &s }

func dropInitialCollections(app core.App) error {
	for _, name := range []string{"votes", "date_options", "participants", "events"} {
		col, err := app.FindCollectionByNameOrId(name)
		if err != nil {
			continue // already gone
		}
		if err := app.Delete(col); err != nil {
			return err
		}
	}
	return nil
}
