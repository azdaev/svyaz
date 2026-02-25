# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make setup    # Install npm deps, copy HiQ CSS to static/css/
make run      # Dev server: go run ./cmd/server (localhost:3000)
make build    # Build binary to bin/server
```

No tests or linter configured yet.

## Required Environment

`BOT_TOKEN` and `CSRF_SECRET` must be set (via `.env` or environment). See `.env.example`.

Optional: `DEV_LOGIN=1` enables `/auth/dev` — a local user picker that bypasses Telegram auth.

## Architecture

Svyaz is a Go SSR web app for team formation. Telegram Login Widget for auth, SQLite for storage, no frontend framework.

**Request flow:** chi router → middleware (logging, recovery, auth, CSRF) → handler → repo → SQLite

**Key layers:**

- `cmd/server/main.go` — entry point, loads config, starts HTTP server
- `internal/config/` — env var loading and validation
- `internal/handler/` — all HTTP handlers and router setup
  - `handler.go` — `Handler` struct, route definitions, template rendering with FuncMap
  - `auth.go` — Telegram HMAC validation, session cookie management, dev login
  - `pages.go` — GET handlers that render templates
  - `api.go` — POST handlers for mutations (CSRF-protected)
- `internal/middleware/auth.go` — extracts session token from cookie, loads user into context
- `internal/models/models.go` — domain structs (User, Project, Response, Role with Count/Filled, Notification)
- `internal/repo/` — database layer, one file per entity (repo, users, sessions, projects, responses, notifications), raw SQL queries
- `migrations/` — goose SQL migrations, auto-run on startup
- `templates/` — Go html/template files, all extend `base.html`
- `static/css/style.css` — all custom styles (HiQ framework as base)
- `static/js/app.js` — notifications dropdown, user menu, CSRF helper, role card toggle/stepper, role picker

**Auth flow:** Telegram widget → `/auth/telegram` validates HMAC → upsert user → create session token → set httpOnly cookie → redirect to `/onboarding` (if new) or `/`. Dev login (`DEV_LOGIN=1`): `/auth/dev` shows user picker → `/auth/dev/{id}` creates session directly.

**Responses:** Each response optionally links to a project role via `role_id`. When responding, users pick a role from the project's role list. Accepted responses are counted per role to track filled positions (`Role.Filled`).

**CSRF:** HMAC-SHA256 of session cookie with `CSRF_SECRET`. Checked on all POST requests via form field `csrf_token` or header `X-CSRF-Token`.

**Database:** SQLite with WAL mode, foreign keys enabled. Connection opened in `repo.New()` which also runs goose migrations. JSON arrays (skills, stack) stored as TEXT columns. `project_roles` junction table has a `count` column for how many people are needed per role. `responses.role_id` is nullable (backwards-compat with old responses).

## Conventions

- UI language is Russian
- Repo methods return `(result, error)`, handlers log errors and return HTTP status codes
- User context: `middleware.UserFromContext(r.Context())`
- Route params: `chi.URLParam(r, "id")` parsed to int64
- Form arrays: `r.Form["roles"]` for checkbox groups, `r.Form["role_count_<id>"]` for role counts
- Comma-separated tags: skills and stack parsed by splitting on `,`
- Templates receive `map[string]any` with user, csrf_token, notification count, and page-specific data
- Template FuncMap includes: `join`, `hasRole`, `roleCount`, `initial`

## Deployment

Docker multi-stage build → Alpine. Deployed on Coolify at svyaz.fitra.tech. SQLite file persisted via `/data` volume mount.
