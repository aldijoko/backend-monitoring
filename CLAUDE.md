# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Related repository

The frontend lives in a sibling repo at `D:\Source Code\cctv\monitoring-cctv` (SvelteKit/Svelte 5 admin panel, also has its own `CLAUDE.md`). It is not a subdirectory of this repo — access it via its own absolute path. When a backend change affects the API contract (new endpoint, DTO/response shape change, RBAC rule), check/update the matching frontend code (`src/lib/api/*`, `src/lib/types/api.ts`) directly rather than guessing at how it's consumed; both repos are local and safe to access for development without asking for confirmation first.

## Commands

```bash
go run cmd/api/main.go        # run the API server (reads .env, defaults to :8000)
go build ./...                # build everything
go vet ./...                  # static checks

go run cmd/migrate/main.go up     # apply migrations (migrations/*.sql, golang-migrate)
go run cmd/migrate/main.go down   # roll back one step

go run cmd/createadmin/main.go    # bootstrap the first superadmin from ADMIN_* env vars (no-op if it already exists)

scripts/dev.sh                    # convenience: runs mediamtx + API together, stops both on Ctrl+C

# Or, run API + mediamtx in Docker instead of natively (see "Docker dev" below):
docker compose up --build            # API (hot-reload via air) + mediamtx
docker compose exec api go run cmd/migrate/main.go up
docker compose exec api go run cmd/createadmin/main.go
```

There are no automated tests in this repo.

### Docker dev

`docker-compose.yml` runs the API (via `air`, hot-reloading on any `.go` change — source is bind-mounted, not baked into the image) and `mediamtx` as containers, verified working end-to-end (login, migrations, and the mediamtx control-API call all confirmed from inside the containers). It deliberately does **not** include a Postgres service: the `api` container joins the external Docker network created by the sibling `backend-cpma` project (`backend-cpma_bnpt-cpma-network`, aliased `postgres`) and talks to the same `bnpt-cpma-postgres` container described below. Start that container first (`docker start bnpt-cpma-postgres`) — this repo's compose will fail to attach to the network otherwise. The frontend (`monitoring-cctv` repo) is not part of this compose; keep running it with `npm run dev`, which proxies to the port this compose publishes (`8000`).

The `mediamtx` service uses `network_mode: "service:api"` (shares the `api` container's network namespace) instead of joining a normal bridge network. This is load-bearing, not incidental: mediamtx.yml's control API (`action: api`) is only allowed unauthenticated from loopback (`127.0.0.1`/`::1`), and `internal/service/mediamtx_service.go` calls it with no auth header at all. On an ordinary bridge network the `api` container's requests would arrive from its own non-loopback IP and get rejected; sharing the namespace makes `http://localhost:9997` from inside `api` genuinely be loopback, so mediamtx.yml's existing allowlist doesn't need to be loosened. Consequence: `mediamtx` has no `ports`/`networks` of its own — all published ports (including mediamtx's) live on the `api` service, and `mediamtx` `depends_on: api` (not the other way around) since it needs `api`'s netns to exist first. `.env`'s `MEDIAMTX_API_URL=http://localhost:9997` / `MEDIAMTX_HLS_BASE_URL=http://localhost:8888` need no docker-specific override as a result — both are already correct for the API container (genuine loopback) and the browser on the host (published port), unlike `DB_HOST`, which docker-compose.yml does override to `postgres`.

Only the mediamtx ports this app actually uses (RTSP, RTP/RTCP, HLS, control API) are published, even though `mediamtx.yml`'s template defaults also enable webrtc/srt/rtmp/moq.

**`.air.toml` needs `poll = true`.** Confirmed the hard way: with the default `poll = false` (inotify), an edit to a bind-mounted `.go` file sat completely unpicked-up for minutes under Docker Desktop on Windows — the container never rebuilt. `backend-cpma`'s `.air.toml` (native Linux/macOS bind mount) doesn't need this; this one does. If a code change doesn't seem to take effect in the running container, check `docker compose logs api` for a `building...`/`running...` line before assuming the bug is elsewhere.

**Restarting the `api` container breaks `mediamtx` until it's also restarted.** Because `mediamtx` uses `network_mode: "service:api"`, it's attached to `api`'s network namespace at the moment `api`'s container process starts — `docker compose restart api` (or any full container recreate, not a normal `air` hot-reload, which only replaces the process inside the same still-running container) tears down and rebuilds that namespace, orphaning `mediamtx` from it. Symptom: every port on `api` (8000, 8888, 8554, 9997, …) stops responding even though `docker compose ps` shows both containers `Up`. Fix: `docker compose restart mediamtx` right after. If you ever automate container restarts here, restart both together.

**First-time setup gotcha:** this repo's docs/`.env.example` assume the `cctv_user` role and `monitoring_cctv` database already exist inside `bnpt-cpma-postgres` — on a machine where they don't (e.g. that container's volume was recreated), `docker compose up` will fail with a Postgres auth/does-not-exist error. Provision them once with `docker exec bnpt-cpma-postgres psql -U postgres -c "CREATE ROLE cctv_user WITH LOGIN PASSWORD '<DB_PASSWORD from .env>';"` then `-c "CREATE DATABASE monitoring_cctv OWNER cctv_user;"`, then run migrations as above. This only adds a role/database inside the shared container — it does not touch `bnpt_cpma`/`bnpt_perwabkeu` or anything else backend-cpma owns.

### Local dev prerequisites

This backend does **not** run its own Postgres container — it uses a database inside the existing `bnpt-cpma-postgres` container (see `.env.example` for the `cctv_user`/`monitoring_cctv` credentials already provisioned there). Don't spin up a second Postgres; just make sure that container is running (`docker ps | grep bnpt-cpma-postgres`).

Streaming requires `mediamtx` and `ffmpeg` (installed via `brew install mediamtx ffmpeg`):

```bash
mediamtx mediamtx/mediamtx.yml    # media-plane: pulls RTSP, serves HLS on :8888, API on :9997
```

The default homebrew mediamtx config binds its RTP listener to UDP `:8000`, which collides with this API's `:8000` — `mediamtx/mediamtx.yml` in this repo overrides `rtpAddress`/`rtcpAddress` to `:8100`/`:8101`. Always run mediamtx with **this repo's config file**, not the homebrew default, or you'll get mysterious port-kill collisions if anything ever does `lsof -ti:8000 | kill`.

`hlsVariant` is set to `mpegts`, not the homebrew default `lowLatency` — deliberately downgraded after LL-HLS (fMP4 + `#EXT-X-PART` blocking-reload) failed to play in the browser against a real relayed source (a public ATCS Dishub HLS feed pulled over the internet, not a local camera). The media data itself was verified valid (fMP4 segment + init concatenated decodes fine via `ffprobe`), so the failure was LL-HLS's strict client/server timing against a jittery upstream, not corrupt output. Plain MPEG-TS segments are self-contained and far more tolerant of that. If you ever revert to `lowLatency`, also flip `StreamPlayer.svelte`'s `lowLatencyMode` back to `true` — they must match — and expect it to work fine for local/simulated sources but potentially fail again for internet-relayed ones.

Paths registered via the config API (`RegisterPath`/`UpdatePath`) are **not persisted to `mediamtx.yml`** — they live in mediamtx's in-memory config only. A mediamtx restart forgets every camera path even though the DB still has the right `stream_url`; cameras will show `offline` until something re-triggers `UpdatePath` (e.g. a `PATCH /cameras/:id` with a changed `source_url` — `CameraStatusPoller` alone won't re-register anything, it only reads status). There's no automatic re-registration-on-startup in this codebase yet; if mediamtx restarts in dev, expect to re-`PATCH` existing cameras (or build a startup reconciliation step if this becomes annoying).

For local testing without real cameras, simulate an RTSP source by publishing into mediamtx itself:

```bash
ffmpeg -re -f lavfi -i "testsrc=size=1280x720:rate=25" -f lavfi -i "sine=frequency=1000" \
  -c:v libx264 -preset ultrafast -tune zerolatency -c:a aac -f rtsp rtsp://localhost:8554/simsource
```

Then create/patch a camera with `source_url: "rtsp://localhost:8554/simsource"` and the status poller will pick it up within ~10s.

## Architecture

Standard layered Go service on Gin + GORM: `handler -> service -> repository -> model`, wired together manually in `internal/router/router.go` (`Setup()`) — the composition root. When adding a feature, follow this same wiring pattern.

This was scaffolded referencing a sibling project (`backend-cpma`) for the layered structure and Postgres/migration tooling, but deliberately **without** its DB-backed RBAC-policy engine or its Integrahub-style generic integration — this app only needs 3 fixed roles (see RBAC below), and a permission-matrix system would be unused complexity.

- `cmd/api` — main server entrypoint. `cmd/migrate` — golang-migrate CLI wrapper. `cmd/createadmin` — bootstraps the first superadmin.
- `internal/handler` — Gin handlers; parse/validate, call service, respond via `pkg/response`.
- `internal/service` — business logic.
- `internal/repository` — GORM queries, one per domain table.
- `internal/model` — GORM models.
- `internal/dto` — request/response payload structs, separate from models (some models are also used directly in responses, e.g. `model.User`, `model.Edge`, `model.Recording` — only `Camera` has a dedicated response DTO, because of the `source_url` visibility rule below).
- `internal/middleware` — `AuthMiddleware` (JWT) and `RequireRole(...)` (plain allow-list against the JWT role claim — no DB lookup).
- `internal/router/router.go` — composition root and all route definitions/RBAC groups.
- `pkg/` — `database` (GORM connection), `jwt`, `response`.
- `config/config.go` — env-driven config via `.env` (godotenv), defaults baked in.

### Response contract — do not add a `{success, data}` envelope

The frontend's `client.ts` does `return (await res.json()) as TResponse` with **no unwrapping**. `pkg/response.OK(c, status, data)` therefore writes `data` directly at the top level, and `pkg/response.Fail(...)` writes `{code, message, details?}` matching the frontend's `ApiError` type. If you're used to `backend-cpma`'s `{success, message, data, error}` envelope, don't carry that pattern over here — it will silently break every frontend call site that expects the raw object.

### RBAC

Three fixed roles: `superadmin` (only role that can manage `/users`), `admin` (full CRUD on edges/cameras/recordings), `viewer` (read-only, GET-only routes). Enforced per-route in `router.go` via three groups (`auth` = any authenticated role, `admin` = admin+superadmin, `superadmin` = superadmin only) — there is no per-permission table; adding a new fixed role or capability means editing these groups directly, not a database migration.

### Cameras: `source_url` vs `stream_url`, and why there's a separate DTO

`model.Camera.SourceURL` (the RTSP source with embedded credentials) is tagged `json:"-"` and never serialized from the model directly. `dto.CameraResponse.SourceURL` is a pointer, populated only by `CameraService.Get` (the single-camera admin detail endpoint, gated to admin/superadmin in the router) and left nil by `CameraService.List` (used by both `/cameras` and `/live` in the frontend — must never leak credentials to a `viewer`-role session or a live-view-only consumer). If you add a new camera-returning endpoint, decide explicitly which of these two behaviors it needs — don't default to "include everything."

`cameras.ts`'s frontend `updateCamera()` PATCHes a **partial** body for two different UI actions (full-form edit and the lightweight enable/disable toggle) — that's why `CameraPatchPayload` (all pointer fields) exists separately from `CameraFormPayload` (all required, used only by create). The `CameraService.Patch` method only touches fields that are non-nil in the request.

### Streaming pipeline (mediamtx)

`internal/service/media_provider.go` defines the `MediaProvider` interface (`RegisterPath`/`UpdatePath`/`RemovePath`) that `CameraService` depends on — `NoopMediaProvider` (no-op, empty `stream_url`) lets camera CRUD work before mediamtx is deployed; `MediaMTXService` (`mediamtx_service.go`) is the real implementation, calling mediamtx's runtime config API to register one path per camera (`camera-{id}`). mediamtx itself handles the pull, reconnect-on-failure, and HLS transcode — there is no manual `ffmpeg` process management in this codebase, and there shouldn't be; if you're about to spawn an `ffmpeg` subprocess from Go, that's very likely the wrong approach here.

`UpdatePath` deletes-then-adds the path (`DELETE /v3/config/paths/delete/{name}` then `POST .../add/{name}`) rather than calling mediamtx's own `PATCH /v3/config/paths/patch/{name}`, which is what the mediamtx docs point you at. Reproduced repeatedly against these particular pulled HLS sources with `hlsAlwaysRemux: true` (see below): PATCH updates the path config fine, but the path's HLS muxer gets stuck reporting `"muxer is waiting to be created"` (HTTP 500 on every read) and mediamtx never actually recreates it — full delete+add doesn't have this problem. If mediamtx ever changes this behavior and PATCH becomes reliable again, reverting is a one-line change, but verify against a real flaky pulled source (a local test source that never drops won't reproduce it).

This app sets two non-default mediamtx path/server options, both because every camera here is a **pulled external/public feed** (government ATCS HLS streams today, potentially real internet-facing RTSP cameras later) that can drop and reconnect on its own, unlike a local camera on a LAN:
- **`alwaysAvailable: true`** + `alwaysAvailableTracks: [{codec: H264}]`, set per-path in `pathBody()` (mediamtx.yml's `pathDefaults` cannot carry this — mediamtx refuses to start with `'alwaysAvailable' cannot be used in a path with a regular expression`, because of the `all_others` catch-all in `paths:`). Without it, a reader sees a plain black tile with zero signal while the source is down, indistinguishable from "still loading". With it, mediamtx concatenates a placeholder segment (default: "STREAM IS OFFLINE" text, no re-encoding) onto the output and keeps the reader connected, switching back automatically the instant the source returns. Verified by pointing a live camera at a nonexistent path and confirming the HLS output kept returning 200 with an obviously-different placeholder resolution (1920x1080@30 vs the real feed's 640x480@25) instead of erroring.
- **`hlsAlwaysRemux: true`** (global, `mediamtx.yml`). Default mediamtx behavior only starts converting a path to HLS output when a viewer first requests it, so every fresh page load costs a multi-second black tile per camera while that muxer spins up — acceptable for occasional single-camera viewing, not for this app's actual use case (a live monitoring grid where most/all cameras have a viewer most of the time). Muxers now start as soon as a path is registered and stay warm regardless of readers. mediamtx doesn't transcode (H264 passthrough remux only, confirmed by upstream docs: "negligible computational power"), so this is cheap even with many cameras — the cost that matters at scale is network/bandwidth to each source, not CPU.

`camera.source_url` is **not RTSP-only** — mediamtx's `source` field accepts whatever scheme the underlying tool understands, and this has been verified against a real source, not just assumed: `rtsp://`/`rtsps://` (a real camera/NVR), and `http://`/`https://` pointing at an existing HLS `.m3u8` (mediamtx pulls it as an `hlsSource` and re-serves it under our own path — confirmed by pulling a live Bandung ATCS Dishub public feed). `rtmp://`/`rtmps://` should work the same way but hasn't been verified against a real source. Don't assume every `source_url` is a raw camera; some cameras in this system are actually *re-streams of someone else's public HLS feed* — that's a legitimate, supported case, not a hack.

`CameraStatusPoller` (`camera_status_poller.go`), started as a goroutine in `cmd/api/main.go`, polls mediamtx's *runtime* path status (`GET /v3/paths/get/{name}`, distinct from the *config* API used for registration) every 10s for every enabled camera, writes `camera.status`/`last_seen_at`, then calls `EdgeRepository.RecalculateStatus` for every touched edge.

### Edges have no real heartbeat — status is derived, not reported

This is a single-site/centralized deployment: there is no physical device at an "edge" location sending its own heartbeat. `edges.status` and `edges.last_heartbeat` are **derived** from the aggregate status of that edge's cameras (`EdgeRepository.RecalculateStatus`, called by the status poller) — online if any camera is online/recording, pending if it has zero cameras, offline otherwise. Don't add a literal heartbeat-ping endpoint or a "last seen" field that something external is expected to POST to; that would contradict the architecture this was built against (see the RTSP-source discussion this app's design came out of — cameras are pulled directly by the backend/mediamtx over the network, not pushed by on-site agents).

### Dashboard: honest zeros, not fabricated telemetry

`DashboardService.Summary()` populates only fields backed by real data (edge/camera counts, recording storage/duration from the `recordings` table). The frontend's `DashboardData.edges[].cpu/memory/disk/network_*` time series exist in the DTO shape (`dto/dashboard_dto.go`) for frontend compatibility but are always returned with an empty `Data` slice — there is no hardware telemetry source in this architecture. Don't fill these with plausible-looking random numbers; if real telemetry is ever needed, that's a scope discussion (it would likely mean introducing an actual per-site agent, which this architecture explicitly avoided).

### Database

Single Postgres database (`monitoring_cctv`), hosted inside the sibling `backend-cpma` project's already-running Postgres container, under its own role (`cctv_user`) with a separate `schema_migrations` table — fully isolated from that project's own database despite sharing the container. See "Local dev prerequisites" above. This was a deliberate resource-sharing choice for local dev, not a coupling of the two applications' code — don't import anything from `backend-cpma` or assume its models/config apply here.
