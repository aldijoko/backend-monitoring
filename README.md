# Monitoring CCTV — Backend

API backend untuk sistem monitoring CCTV. Go (Gin + GORM), layered `handler -> service -> repository -> model`.

## Cara menjalankan (Docker — direkomendasikan)

Cara ini yang sudah diverifikasi jalan di lingkungan dev (Windows + Docker Desktop). Menjalankan API (hot-reload via `air`) dan mediamtx sebagai container, tanpa bikin Postgres container baru — keduanya numpang ke container `bnpt-cpma-postgres` milik proyek sibling `backend-cpma`.

### Prasyarat

- Docker Desktop
- Container `bnpt-cpma-postgres` (dari proyek `backend-cpma`) harus sudah jalan:
  ```bash
  docker start bnpt-cpma-postgres
  ```
  Network eksternal `backend-cpma_bnpt-cpma-network` harus sudah ada (otomatis ada selama proyek `backend-cpma` pernah di-`docker compose up` minimal sekali).

Go SDK **tidak wajib** di-install di host untuk menjalankan backend (semuanya jalan di dalam container) — hanya perlu kalau mau IntelliSense/gopls di editor.

### Setup (pertama kali saja)

1. Copy env:
   ```bash
   cp .env.example .env
   ```
2. Pastikan role `cctv_user` dan database `monitoring_cctv` sudah ada di dalam `bnpt-cpma-postgres`. Kalau container Postgres itu baru/volume-nya baru dibuat ulang, role dan database ini belum ada dan perlu dibuat sekali:
   ```bash
   docker exec bnpt-cpma-postgres psql -U postgres -c "CREATE ROLE cctv_user WITH LOGIN PASSWORD 'cctv_dev_password';"
   docker exec bnpt-cpma-postgres psql -U postgres -c "CREATE DATABASE monitoring_cctv OWNER cctv_user;"
   ```
   (password harus sama dengan `DB_PASSWORD` di `.env`)
3. Build & start:
   ```bash
   docker compose up --build -d
   ```
4. Jalankan migrasi database:
   ```bash
   docker compose exec api go run cmd/migrate/main.go up
   ```
5. Bootstrap akun superadmin pertama (baca env `ADMIN_USERNAME`/`ADMIN_EMAIL`/`ADMIN_PASSWORD`, no-op kalau sudah ada):
   ```bash
   docker compose exec api go run cmd/createadmin/main.go
   ```

### Menjalankan sehari-hari

```bash
docker compose up -d       # start API + mediamtx (kalau belum jalan)
docker compose logs -f api # lihat log, termasuk auto-rebuild air tiap ada perubahan .go
docker compose down        # stop semua (data Postgres tidak kepengaruh, itu di container terpisah)
```

API tersedia di `http://localhost:8000` (dipakai frontend via `npm run dev`'s Vite proxy), mediamtx HLS di `http://localhost:8888`, mediamtx control API di `http://localhost:9997`.

Detail kenapa desainnya begini (network, `network_mode: service:api`, dll) ada di [CLAUDE.md](CLAUDE.md#docker-dev).

## Cara menjalankan (native, tanpa Docker)

Alternatif kalau tidak mau pakai Docker sama sekali (misal di macOS dengan `brew`).

### Prasyarat

- Go 1.25+
- Docker hanya untuk container `bnpt-cpma-postgres` (backend ini tetap **tidak** punya Postgres sendiri, numpang database `monitoring_cctv` di container tersebut lewat role `cctv_user`)
- [mediamtx](https://github.com/bluenviron/mediamtx) dan `ffmpeg`:
  ```bash
  brew install mediamtx ffmpeg
  ```

Pastikan container Postgres sudah jalan sebelum start API:

```bash
docker ps | grep bnpt-cpma-postgres
```

### Setup

1. Copy env:
   ```bash
   cp .env.example .env
   ```
2. Jalankan migrasi database:
   ```bash
   go run cmd/migrate/main.go up
   ```
3. Bootstrap akun superadmin pertama:
   ```bash
   go run cmd/createadmin/main.go
   ```

### Menjalankan

Jalankan media-plane (mediamtx) — **wajib pakai config repo ini**, bukan default homebrew, karena `rtpAddress`/`rtcpAddress` di-override ke `:8100`/`:8101` supaya tidak bentrok dengan port API (`:8000`):

```bash
mediamtx mediamtx/mediamtx.yml
```

Jalankan API server (default `:8000`):

```bash
go run cmd/api/main.go
```

Atau jalankan keduanya bareng sekaligus:

```bash
scripts/dev.sh
```

## Perintah lain

```bash
go build ./...                    # build semua
go vet ./...                      # static check
go run cmd/migrate/main.go down   # rollback 1 migrasi
```

Tidak ada automated test di repo ini.

## Testing tanpa kamera fisik

Simulasikan sumber RTSP dengan publish langsung ke mediamtx (port `8554` sama saja baik mediamtx jalan native maupun di container, karena di-publish ke host):

```bash
ffmpeg -re -f lavfi -i "testsrc=size=1280x720:rate=25" -f lavfi -i "sine=frequency=1000" \
  -c:v libx264 -preset ultrafast -tune zerolatency -c:a aac -f rtsp rtsp://localhost:8554/simsource
```

Lalu buat/patch kamera dengan `source_url: "rtsp://localhost:8554/simsource"` — status poller akan mendeteksinya online dalam ~10 detik.

## Catatan penting

- **Path kamera di mediamtx tidak persisten.** Registrasi path (`RegisterPath`/`UpdatePath`) hanya hidup di memory mediamtx. Kalau mediamtx restart (atau container-nya di-recreate), semua path kamera hilang meski data di DB tetap ada — kamera akan tampil `offline` sampai ada `PATCH /cameras/:id` yang men-trigger registrasi ulang.
- `source_url` mendukung `rtsp://`, `rtsps://`, `http(s)://` (HLS `.m3u8` yang mediamtx pull ulang), bukan cuma kamera fisik langsung — sebagian kamera di sistem ini adalah re-stream dari feed HLS publik pihak lain.
- Frontend (repo sibling `monitoring-cctv`) **tidak** ikut di-docker di sini — dijalankan native lewat `npm run dev`, lihat README repo tersebut.
- Detail arsitektur lengkap (RBAC, response contract, DTO camera, dashboard, setup Docker, dsb.) ada di [CLAUDE.md](CLAUDE.md).
