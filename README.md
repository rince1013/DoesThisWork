# Does This Work?

A no-login, mobile-first event planner for friends. Someone picks a name and some dates, shares a link, and everyone votes. No accounts, no friction.

## How it works

1. **Create** — name your event, pick some candidate dates
2. **Share** — send the link to your friends
3. **Vote** — everyone picks the dates that work for them (or suggests new ones)
4. **Lock** — the creator locks in the winner

Results update in real time for everyone on the page.

## Stack

- **[PocketBase](https://pocketbase.io/) v0.36.9** — Go backend + embedded SQLite + realtime SSE
- **[HTMX](https://htmx.org/) v2** — HTML-driven interactivity, no JS framework
- **Go templates** — server-rendered HTML
- **[PocketBase JS SDK](https://github.com/pocketbase/js-sdk)** — realtime subscriptions

## Development

### Prerequisites

- [mise](https://mise.jdx.dev/) — manages the Go version

### Setup

```bash
# Install the Go version pinned in mise.toml
mise install

# Run the dev server (auto-migrates DB on first run)
make run
```

Open [http://localhost:8090](http://localhost:8090).

The PocketBase admin dashboard is at [http://localhost:8090/_/](http://localhost:8090/_/).

### Available commands

```bash
make run      # Start dev server at localhost:8090
make build    # Build binary for current platform
make deploy   # Deploy to Fly.io
make clean    # Remove built binary
```

## Deployment

Hosted on [Fly.io](https://fly.io) — free tier with persistent storage. HTTPS is automatic.

### Prerequisites

```bash
brew install flyctl
fly auth login
```

### First-time setup

```bash
# Create the app (uses fly.toml config)
fly apps create does-this-work

# Create the persistent volume for SQLite data (free up to 3GB)
fly volumes create dtw_data --region iad --size 1

# Add your custom domain
fly certs add dtw.raishadandlisa.com
# Fly gives you an A record and AAAA record to add on Namecheap (Advanced DNS)

# Deploy
make deploy
```

### Subsequent deploys

```bash
make deploy
```

Fly builds the Docker image and does a rolling deploy. No downtime.

## Project structure

```
.
├── main.go                  # Entry point
├── mise.toml                # Go version pin
├── Makefile
├── migrations/              # PocketBase schema migrations
├── handlers/
│   ├── routes.go            # Route registration
│   ├── helpers.go           # Cookie identity, view models, DB queries
│   ├── events.go            # Create event, event page
│   ├── participants.go      # Join flow
│   ├── dates.go             # Add / delete date options
│   ├── votes.go             # Vote toggle, lock date, results fragment
│   ├── templates.go         # Template parsing and rendering
│   └── templates/           # Go HTML templates
├── static/
│   ├── style.css            # Mobile-first CSS
│   └── app.js               # Realtime subscriptions, emoji picker, clipboard
├── doesthiswork.service     # systemd unit (VPS alternative)
├── Dockerfile               # For Fly.io builds
├── fly.toml                 # Fly.io app config
├── Caddyfile                # Reverse proxy (VPS alternative)
```

## Identity (no login)

Identity is cookie-based — no accounts needed.

- **Creator** — gets a `dtw_c_{eventId}` cookie when they create the event
- **Participant** — gets a `dtw_p_{eventId}` cookie when they join

Tokens are stored in the database as hidden fields (never exposed via API). Losing your cookies means losing your identity for that event, but you can simply rejoin as a participant.
