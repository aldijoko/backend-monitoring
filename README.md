# Monitoring CCTV — Backend

API backend untuk sistem monitoring CCTV. Go (Gin + GORM), layered `handler -> service -> repository -> model`.

## Prasyarat

- Go 1.25+
- Docker (container `bnpt-cpma-postgres` sudah jalan — backend ini **tidak** punya Postgres container sendiri, ia numpang database `monitoring_cctv` di container tersebut lewat role `cctv_user`)
- [mediamtx](https://github.com/bluenviron/mediamtx) dan `ffmpeg` untuk streaming:
  ```bash
  brew install mediamtx ffmpeg
  ```

Pastikan container Postgres sudah jalan sebelum start API:

```bash
docker ps | grep bnpt-cpma-postgres
```

## Setup

1. Copy env dan sesuaikan bila perlu (default sudah cocok untuk dev lokal):
   ```bash
   cp .env.example .env
   ```
2. Jalankan migrasi database:
   ```bash
   go run cmd/migrate/main.go up
   ```
3. Bootstrap akun superadmin pertama (baca env `ADMIN_USERNAME`/`ADMIN_EMAIL`/`ADMIN_PASSWORD`, no-op kalau sudah ada):
   ```bash
   go run cmd/createadmin/main.go
   ```

## Menjalankan

Jalankan media-plane (mediamtx) — **wajib pakai config repo ini**, bukan default homebrew, karena `rtpAddress`/`rtcpAddress` di-override ke `:8100`/`:8101` supaya tidak bentrok dengan port API (`:8000`):

```bash
mediamtx mediamtx/mediamtx.yml
```

Jalankan API server (default `:8000`):

```bash
go run cmd/api/main.go
```

## Perintah lain

```bash
go build ./...                    # build semua
go vet ./...                      # static check
go run cmd/migrate/main.go down   # rollback 1 migrasi
```

Tidak ada automated test di repo ini.

## Testing tanpa kamera fisik

Simulasikan sumber RTSP dengan publish langsung ke mediamtx:

```bash
ffmpeg -re -f lavfi -i "testsrc=size=1280x720:rate=25" -f lavfi -i "sine=frequency=1000" \
  -c:v libx264 -preset ultrafast -tune zerolatency -c:a aac -f rtsp rtsp://localhost:8554/simsource
```

Lalu buat/patch kamera dengan `source_url: "rtsp://localhost:8554/simsource"` — status poller akan mendeteksinya online dalam ~10 detik.

## Catatan penting

- **Path kamera di mediamtx tidak persisten.** Registrasi path (`RegisterPath`/`UpdatePath`) hanya hidup di memory mediamtx. Kalau mediamtx restart, semua path kamera hilang meski data di DB tetap ada — kamera akan tampil `offline` sampai ada `PATCH /cameras/:id` yang men-trigger registrasi ulang.
- `source_url` mendukung `rtsp://`, `rtsps://`, `http(s)://` (HLS `.m3u8` yang mediamtx pull ulang), bukan cuma kamera fisik langsung — sebagian kamera di sistem ini adalah re-stream dari feed HLS publik pihak lain.
- Detail arsitektur lengkap (RBAC, response contract, DTO camera, dashboard, dsb.) ada di [CLAUDE.md](CLAUDE.md).
