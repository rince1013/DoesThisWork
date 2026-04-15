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
make run          # Start dev server at localhost:8090
make build        # Build binary for current platform
make build-linux  # Cross-compile ARM64 Linux binary (for deployment)
make clean        # Remove built binary
make deploy       # Build + deploy to VPS (see Deployment below)
make logs         # Tail logs on the VPS
```

## Deployment

The app compiles to a single self-contained binary — the server needs no Go runtime.

### 1. Build and copy

```bash
DEPLOY_HOST=your.vps.ip DEPLOY_USER=ubuntu make deploy
```

This cross-compiles for Linux ARM64, copies the binary + static files to `/opt/doesthiswork` on the server, and restarts the service.

### 2. First-time server setup

```bash
# Copy the systemd unit
scp doesthiswork.service ubuntu@your.vps.ip:/etc/systemd/system/
ssh ubuntu@your.vps.ip

# On the server
sudo mkdir -p /opt/doesthiswork/pb_data
sudo systemctl enable --now doesthiswork
```

### 3. Nginx + SSL

Edit `nginx.conf` — replace `your.domain.com` with your domain, then:

```bash
sudo cp nginx.conf /etc/nginx/sites-available/doesthiswork
sudo ln -s /etc/nginx/sites-available/doesthiswork /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx

# Free SSL via Certbot
sudo certbot --nginx -d your.domain.com
```

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
├── doesthiswork.service     # systemd unit
└── nginx.conf               # Reverse proxy config
```

## Identity (no login)

Identity is cookie-based — no accounts needed.

- **Creator** — gets a `dtw_c_{eventId}` cookie when they create the event
- **Participant** — gets a `dtw_p_{eventId}` cookie when they join

Tokens are stored in the database as hidden fields (never exposed via API). Losing your cookies means losing your identity for that event, but you can simply rejoin as a participant.
