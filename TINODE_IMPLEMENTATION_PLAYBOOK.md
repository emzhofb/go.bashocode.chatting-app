# Tinode Chat Product — Implementation Playbook

Status: **planning source of truth**  
Target: MVP self-hosted chat untuk maksimum awal 5.000–8.000 akun  
Primary client: responsive web/PWA  
Primary backend language: Go  
Messaging engine: Tinode  
Last reviewed: 2026-08-19

---

## 0. Cara menggunakan dokumen ini

Dokumen ini sengaja sangat eksplisit agar implementasi dapat dikerjakan oleh AI dengan kemampuan terbatas tanpa banyak mengambil keputusan arsitektur sendiri.

Aturan untuk setiap AI/developer yang mengerjakan proyek:

1. Baca bagian **Keputusan yang sudah dikunci**, **Batas scope**, dan milestone aktif sebelum mengubah file.
2. Kerjakan hanya satu task checkbox dalam satu waktu.
3. Sebelum mengedit, sebutkan file yang akan diubah dan alasan perubahan.
4. Jangan mengganti teknologi, membuat abstraction baru, atau menambah dependency tanpa kebutuhan task yang jelas.
5. Jangan memodifikasi source Tinode server. Gunakan image/binary upstream yang dipin versinya.
6. Jangan membaca atau menulis langsung tabel internal database Tinode dari aplikasi kita.
7. Semua komunikasi client dengan Tinode harus melalui SDK/protokol resmi Tinode.
8. Semua business logic milik produk harus berada di `services/app-api` atau client kita, bukan di source Tinode.
9. Setelah satu task selesai, jalankan test yang relevan dan catat hasilnya di bagian **Progress log**.
10. Jika dokumentasi upstream dan dokumen ini berbeda, hentikan pekerjaan, tunjukkan bukti perbedaannya, lalu perbarui keputusan sebelum coding.
11. Jangan menandai task selesai jika acceptance criteria belum semuanya terpenuhi.
12. Jangan menyimpan secret, password, API key, private key, atau token nyata di Git.

Format laporan setiap selesai satu task:

```text
Task:
File berubah:
Perubahan:
Test yang dijalankan:
Hasil test:
Risiko/hal yang belum selesai:
Task berikutnya:
```

---

## 1. Tujuan produk

Membangun aplikasi chat bermerek sendiri dengan pengalaman dasar seperti WhatsApp/Telegram:

- akun dan profil pengguna;
- pencarian pengguna;
- percakapan personal;
- grup;
- pesan teks;
- gambar, file, video pendek, dan voice note;
- status sent/delivered/read;
- typing indicator;
- presence/online status;
- reply, forward, edit, dan pin;
- sinkronisasi beberapa perangkat;
- block dan report;
- voice/video call 1-to-1 setelah messaging stabil;
- deployment mandiri tanpa biaya lisensi per pengguna.

Produk bukan fork bermerek ulang yang membiarkan UI Tinode apa adanya. Tinode dipakai sebagai messaging engine dan referensi perilaku. Branding, business API, domain model produk, serta pengalaman UI tetap milik kita.

---

## 2. Keputusan yang sudah dikunci

AI tidak boleh mengubah keputusan berikut tanpa persetujuan manusia.

| Area | Keputusan |
| --- | --- |
| Messaging engine | Tinode self-hosted |
| Tinode server | Gunakan upstream tanpa modifikasi source |
| Versi awal | Pin ke Tinode `v0.25.3` atau patch stabil lebih baru yang telah diverifikasi sebelum bootstrap |
| Database | PostgreSQL, satu instance dengan database/user terpisah untuk Tinode dan app |
| Business backend | Go, satu service modular monolith bernama `app-api` |
| Auth source of truth | `app-api` melalui Tinode REST authenticator |
| Client pertama | Responsive web/PWA menggunakan `tinode-sdk` |
| Mobile | Android/iOS dikerjakan setelah web dan backend stabil |
| Reverse proxy | Caddy atau reverse proxy existing; satu public TLS entry point |
| Media development | Filesystem volume persisten |
| Media production | S3-compatible object storage setelah milestone media dasar lolos |
| Voice/video | WebRTC 1-to-1 milik Tinode + coturn; dikerjakan paling akhir pada MVP |
| Push | Web push lebih dulu; FCM/APNs hanya saat native client dimulai |
| Deployment awal | Docker Compose satu host |
| Scaling awal | Vertical scaling; cluster hanya setelah hasil load test membuktikan perlu |
| Observability | Structured logs, health checks, PostgreSQL metrics, Tinode exporter jika diperlukan |
| API style | JSON REST dengan version prefix `/v1` |
| App session | Opaque random access/refresh token; token yang disimpan di DB selalu dalam bentuk hash |
| Password | `bcrypt` dengan cost yang dikonfigurasi dan minimum awal 12 |
| ID aplikasi | UUID acak; jangan memakai sequential public IDs |
| Waktu | Simpan UTC di database; konversi timezone hanya di UI |

### Alasan keputusan utama

- Tinode menangani routing pesan, penyimpanan pesan, topic, subscription, access control, sinkronisasi, presence, receipts, attachment, dan signaling call.
- `app-api` tetap dibutuhkan untuk akun produk, sesi business API, report, moderation, audit, dan fitur non-chat pada masa depan.
- REST authenticator menjadikan database aplikasi sebagai sumber kebenaran akun tanpa menulis tabel Tinode secara langsung.
- Web/PWA dipilih lebih dulu karena `tinode-sdk` dan web client referensi resmi tersedia. Native mobile tidak boleh dimulai sebelum kontrak backend dan UX chat stabil.
- Source server Tinode tidak dimodifikasi untuk mengurangi beban maintenance dan risiko kewajiban distribusi GPLv3.

---

## 3. Batas scope

### 3.1 Wajib masuk MVP

- [ ] Local development dapat hidup dengan satu perintah.
- [ ] Register dengan username, email, password, dan display name.
- [ ] Login/logout aplikasi dan Tinode.
- [ ] Refresh app session.
- [ ] Profil dan avatar.
- [ ] Pencarian user berdasarkan username/display name yang diizinkan.
- [ ] Percakapan personal.
- [ ] Membuat grup, mengubah nama/foto grup, menambah/menghapus anggota.
- [ ] Pesan teks.
- [ ] Attachment gambar dan file dengan batas ukuran.
- [ ] Voice note.
- [ ] Sent/delivered/read state.
- [ ] Typing indicator.
- [ ] Presence.
- [ ] Reply, forward, edit, delete lokal/fitur yang memang tersedia upstream, dan pin.
- [ ] Daftar conversation dan unread count.
- [ ] Reconnect setelah jaringan putus.
- [ ] Sinkronisasi setelah login di browser/perangkat kedua.
- [ ] Block user.
- [ ] Report user/message ke `app-api`.
- [ ] Admin dapat melihat report dan suspend akun.
- [ ] Backup dan restore telah diuji.
- [ ] Load test sesuai profil awal telah dijalankan.
- [ ] Voice/video call 1-to-1 melalui coturn setelah chat core stabil.

### 3.2 Bukan bagian MVP

Jangan mengerjakan ini kecuali semua scope MVP selesai dan ada persetujuan eksplisit:

- End-to-end encryption. Tinode saat ini belum menyediakannya sebagai fitur stabil.
- Group voice/video call.
- Federation antarserver.
- Payment/wallet.
- Moments/status feed.
- Mini-app.
- Official account/channel monetization.
- Bot marketplace.
- SMS OTP.
- Import kontak telepon.
- Desktop native app.
- Kubernetes.
- Multi-region.
- Microservices.
- Custom database adapter Tinode.
- Mengubah protokol Tinode.
- Menyalin tampilan, nama, logo, ikon, atau aset milik WhatsApp, Telegram, WeChat, atau Tinode.

### 3.3 Batas keamanan yang harus dikomunikasikan

MVP tidak boleh dipasarkan sebagai messenger end-to-end encrypted. Gunakan kalimat yang akurat:

> Data ditransmisikan melalui TLS dan disimpan di infrastruktur yang kita kelola. Pesan belum end-to-end encrypted dan operator server secara teknis masih dapat mengakses data.

---

## 4. Sumber resmi yang wajib dijadikan referensi

Jangan mengandalkan blog pihak ketiga untuk kontrak protokol.

- Tinode server: <https://github.com/tinode/chat>
- Tinode server API: <https://github.com/tinode/chat/blob/master/docs/API.md>
- Tinode installation: <https://github.com/tinode/chat/blob/master/INSTALL.md>
- Tinode Docker: <https://github.com/tinode/chat/blob/master/docker/README.md>
- Tinode REST authenticator: <https://github.com/tinode/chat/blob/master/server/auth/rest/README.md>
- Tinode server config reference: <https://github.com/tinode/chat/blob/master/server/tinode.conf>
- Tinode JavaScript SDK: <https://github.com/tinode/tinode-js>
- Tinode web reference client: <https://github.com/tinode/webapp>
- Tinode Android client/SDK: <https://github.com/tinode/tindroid>
- Tinode iOS client/SDK: <https://github.com/tinode/ios>
- coturn: <https://github.com/coturn/coturn>

Saat memulai implementasi, catat commit SHA/tag upstream yang dipakai di `docs/UPSTREAM_VERSIONS.md`. Jangan menggunakan `latest` di production.

---

## 5. Konsekuensi lisensi

### 5.1 Tinode server

- Server Tinode berlisensi GPLv3.
- Jalankan sebagai process/container terpisah dan jangan menyalin source-nya ke `services/app-api`.
- Jangan statically link atau copy package internal Tinode server ke aplikasi proprietary.
- Jangan menghapus copyright dan license notice.
- Jika suatu hari binary Tinode yang telah dimodifikasi didistribusikan kepada pelanggan/on-premise, lakukan legal review dan penuhi kewajiban GPLv3.

### 5.2 Tinode client SDK dan reference client

- Client SDK/aplikasi resmi Tinode umumnya Apache 2.0.
- Pertahankan file LICENSE dan attribution upstream ketika melakukan fork.
- Buat `THIRD_PARTY_NOTICES.md` dan catat repository, commit, license, serta bagian yang digunakan.
- Jangan gunakan merek dagang dan logo Tinode sebagai merek produk.

### 5.3 Checklist sebelum release

- [ ] `THIRD_PARTY_NOTICES.md` lengkap.
- [ ] License scanner dijalankan terhadap dependency Go dan npm.
- [ ] Tidak ada dependency tanpa license.
- [ ] Tidak ada dependency GPL/AGPL baru yang di-link ke kode proprietary tanpa review.
- [ ] Attribution yang disyaratkan tersedia.
- [ ] Semua aset visual memiliki bukti hak penggunaan.

Dokumen ini bukan nasihat hukum. Untuk distribusi komersial/on-premise, lakukan review lisensi formal.

---

## 6. Arsitektur target

```text
Browser / PWA
   |
   | HTTPS + WSS
   v
+---------------------------+
| Caddy / reverse proxy     |
|                           |
| /             -> web     |
| /api/*        -> app-api |
| /v0/*         -> Tinode  |
+-------------+-------------+
              |
       Docker private network
              |
     +--------+---------+----------------+
     |                  |                |
     v                  v                v
+------------+    +------------+    +----------+
| app-api    |    | Tinode     |    | coturn   |
| Go         |<---| REST auth  |    | WebRTC   |
+------+-----+    +------+-----+    +----------+
       |                 |
       v                 v
+--------------------------------------------+
| PostgreSQL                                 |
|                                            |
| database app_chat  user app_chat           |
| database tinode    user tinode              |
+--------------------------------------------+
```

### 6.1 Ownership data

| Data | Pemilik | Aturan |
| --- | --- | --- |
| Username/password | `app-api` | Tinode memverifikasi melalui REST authenticator |
| App sessions | `app-api` | Opaque token, hash disimpan di DB |
| Public profile utama | `app-api` | Salinan minimum disinkronkan ke `public` Tinode |
| Tinode user ID | Tinode | Mapping disimpan di `app-api` setelah callback `link` |
| Chat messages | Tinode | Jangan diduplikasi ke app DB |
| Topics/groups/subscriptions | Tinode | Gunakan API/SDK Tinode |
| Attachments chat | Tinode media handler | Development filesystem; production S3-compatible |
| Reports/moderation cases | `app-api` | Referensikan Tinode user/topic/message IDs |
| Suspension | `app-api` + Tinode | Auth harus ditolak; status Tinode disinkronkan oleh admin worker |
| Audit events | `app-api` | Tidak berisi password, token, atau isi pesan penuh |

### 6.2 Aturan database

- Satu PostgreSQL instance boleh digunakan saat MVP, tetapi buat database dan role terpisah.
- Role `app_chat` tidak boleh memiliki permission ke database/schema Tinode.
- Role dedicated Tinode (`tinode_admin` pada development) tidak boleh memiliki permission ke database aplikasi.
- Tinode upstream membutuhkan database fallback dengan nama role ketika role bukan `postgres`; database target messaging tetap `tinode`.
- Migration aplikasi hanya mengubah database `app_chat`.
- Upgrade schema Tinode hanya menggunakan tool/prosedur resmi Tinode.
- Jangan membuat foreign key lintas database.

---

## 7. Struktur repository target

```text
.
├── AGENTS.md
├── README.md
├── TINODE_IMPLEMENTATION_PLAYBOOK.md
├── THIRD_PARTY_NOTICES.md
├── apps/
│   └── web/
│       ├── src/
│       ├── public/
│       ├── tests/
│       ├── package.json
│       └── README.md
├── services/
│   └── app-api/
│       ├── cmd/api/main.go
│       ├── internal/
│       │   ├── auth/
│       │   ├── config/
│       │   ├── health/
│       │   ├── httpapi/
│       │   ├── moderation/
│       │   ├── profile/
│       │   ├── session/
│       │   ├── storage/
│       │   └── tinodeauth/
│       ├── migrations/
│       ├── go.mod
│       └── README.md
├── infra/
│   ├── compose/
│   │   ├── compose.dev.yml
│   │   └── compose.prod.yml
│   ├── caddy/Caddyfile
│   ├── coturn/turnserver.conf.example
│   ├── postgres/init/
│   └── tinode/
│       ├── tinode.conf.example
│       └── ice-servers.json.example
├── docs/
│   ├── API.md
│   ├── ARCHITECTURE.md
│   ├── BACKUP_RESTORE.md
│   ├── DEPLOYMENT.md
│   ├── OPERATIONS.md
│   ├── SECURITY.md
│   ├── UPSTREAM_VERSIONS.md
│   └── adr/
├── scripts/
│   ├── dev-up.sh
│   ├── dev-down.sh
│   ├── smoke-test.sh
│   ├── backup.sh
│   └── restore-test.sh
└── .env.example
```

Jangan membuat semua folder sekaligus. Buat hanya saat milestone yang membutuhkan folder tersebut dimulai.

---

## 8. Konfigurasi dan secret

### 8.1 Prinsip

- `.env.example` hanya memuat nama variable dan nilai dummy aman.
- `.env`, certificate, database dump, dan file secret masuk `.gitignore`.
- Secret production disediakan deployment environment, bukan file Git.
- Production harus menggunakan secret berbeda dari development.
- Rotation procedure harus didokumentasikan sebelum public launch.

### 8.2 Environment contract awal

Nama dapat ditambah hanya jika benar-benar diperlukan.

```dotenv
APP_ENV=development
APP_HTTP_ADDR=:8080
APP_INTERNAL_HTTP_ADDR=:8081
APP_DATABASE_URL=postgresql://app_chat:change-me@postgres:5432/app_chat?sslmode=disable
APP_PUBLIC_BASE_URL=http://localhost:8080
APP_ALLOWED_ORIGINS=http://localhost:3000
APP_ACCESS_TOKEN_TTL=15m
APP_REFRESH_TOKEN_TTL=720h
APP_BCRYPT_COST=12
APP_LOG_LEVEL=debug

TINODE_PUBLIC_URL=http://localhost:6060
TINODE_INTERNAL_URL=http://tinode:6060
TINODE_API_KEY=replace-with-dev-key
TINODE_REST_AUTH_URL=http://app-api:8081/internal/tinode/auth/

POSTGRES_SUPERUSER_PASSWORD=development-only
POSTGRES_TINODE_USER=tinode_admin
POSTGRES_TINODE_PASSWORD=development-only
POSTGRES_APP_PASSWORD=development-only

TURN_REALM=localhost
TURN_USERNAME=development-only
TURN_PASSWORD=development-only

WEB_PUBLIC_API_BASE_URL=http://localhost:8080/api
WEB_PUBLIC_TINODE_URL=ws://localhost:6060/v0/channels
WEB_PUBLIC_TINODE_API_KEY=replace-with-dev-key
```

Catatan:

- Nama variable frontend harus disesuaikan dengan build tool yang benar-benar digunakan. Jangan mengasumsikan Vite/Next.js sebelum membaca `package.json` client yang dipilih.
- API key Tinode memang diperlukan client; perlakukan sebagai application identifier, bukan satu-satunya security boundary.
- `token.key`, `uid_key`, database password, TURN credential, dan TLS private key adalah secret sesungguhnya.

### 8.3 Tinode config wajib

Sebelum production:

- [ ] PostgreSQL adapter aktif.
- [ ] DSN tidak memakai default password.
- [ ] `token.key` diganti random minimal 32 byte.
- [ ] `uid_key` diganti random 16 byte.
- [ ] API key baru dibuat dengan utility resmi Tinode.
- [ ] REST authenticator menunjuk private internal listener `app-api`.
- [ ] `logical_names` memetakan external `basic` ke internal `rest` jika app DB sudah menjadi source of truth.
- [ ] REST auth original name disembunyikan dari client.
- [ ] CORS/origin dibatasi ke domain aplikasi.
- [ ] TLS hanya diterminasi pada reverse proxy atau Tinode, tidak keduanya tanpa alasan.
- [ ] Sample data dimatikan.
- [ ] Default/root/demo credentials dihapus.
- [ ] Upload size limit ditetapkan.
- [ ] Media volume/bucket persisten.
- [ ] WebRTC dimatikan sampai coturn valid.
- [ ] Metrics/health endpoint tidak dibuka ke internet.

---

## 9. Model data `app-api`

Schema berikut adalah minimum. Nama field boleh disesuaikan dengan konvensi Go/SQL, tetapi maknanya tidak boleh berubah tanpa migration dan ADR.

### 9.1 `users`

```text
id                    UUID primary key
username              varchar(32), normalized lowercase, unique
email                 varchar(320), normalized lowercase, unique
password_hash         text
display_name          varchar(80)
avatar_url            text nullable
status                enum-like varchar: active|suspended|deleted
tinode_uid_raw        varchar nullable unique
tinode_user_id        varchar nullable unique
email_verified_at     timestamptz nullable
created_at            timestamptz not null
updated_at            timestamptz not null
deleted_at            timestamptz nullable
```

Rules:

- Username hanya `[a-z0-9._]`, panjang 4–32, tidak boleh diawali/diakhiri titik, dan tidak boleh memiliki dua titik berurutan.
- Simpan username/email dalam normalized form.
- Display name boleh Unicode setelah validasi panjang rune.
- Jangan hard-delete user pada request biasa.
- `tinode_uid_raw` diisi idempotently oleh endpoint REST authenticator `link`.
- `tinode_user_id` adalah bentuk public Tinode jika dapat diturunkan/diterima secara resmi; jangan menebak format tanpa test.

### 9.2 `sessions`

```text
id                    UUID primary key
user_id               UUID foreign key users(id)
access_token_hash     char(64) unique
refresh_token_hash    char(64) unique
access_expires_at     timestamptz
refresh_expires_at    timestamptz
user_agent            text nullable
ip_address            inet nullable
last_seen_at          timestamptz
revoked_at            timestamptz nullable
created_at            timestamptz
```

Rules:

- Generate token dengan CSPRNG minimal 32 byte.
- Kirim token sebagai URL-safe base64 tanpa padding.
- Simpan SHA-256 token, bukan token asli.
- Refresh melakukan rotation: token lama langsung direvoke.
- Logout merevoke session aktif.
- Perubahan password merevoke seluruh session user.

### 9.3 `reports`

```text
id                    UUID primary key
reporter_user_id      UUID foreign key users(id)
reported_user_id      UUID nullable foreign key users(id)
tinode_topic_id       varchar nullable
tinode_message_seq    bigint nullable
category              varchar: spam|harassment|illegal|other
description           varchar(1000) nullable
status                varchar: open|reviewing|resolved|rejected
resolution_note       varchar(2000) nullable
created_at            timestamptz
updated_at            timestamptz
resolved_at           timestamptz nullable
```

Jangan menyalin isi pesan ke report secara otomatis. Jika bukti isi pesan memang dibutuhkan, buat kebijakan privasi dan retention terpisah.

### 9.4 `audit_logs`

```text
id                    bigserial primary key
actor_user_id         UUID nullable
actor_type            varchar: user|admin|system
action                varchar
target_type           varchar
target_id             varchar nullable
metadata              jsonb
created_at            timestamptz
```

Metadata tidak boleh memuat password, secret, raw token, atau isi pesan penuh.

---

## 10. Kontrak HTTP `app-api`

### 10.1 Response format

Success tunggal:

```json
{
  "data": {}
}
```

Success list:

```json
{
  "data": [],
  "pagination": {
    "next_cursor": null
  }
}
```

Error:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Input tidak valid",
    "fields": {
      "username": "Username sudah digunakan"
    },
    "request_id": "req_..."
  }
}
```

Jangan mengirim stack trace, SQL error, atau detail internal ke client.

### 10.2 Public endpoints

#### `POST /v1/auth/register`

Request:

```json
{
  "username": "alice",
  "email": "alice@example.com",
  "password": "correct horse battery staple",
  "display_name": "Alice"
}
```

Behavior:

1. Normalize dan validate input.
2. Pastikan username/email belum digunakan.
3. Hash password dengan bcrypt.
4. Insert user dalam transaction.
5. Buat app session.
6. Jangan membuat row langsung ke database Tinode.
7. Client kemudian melakukan first login ke Tinode; REST authenticator akan meminta Tinode membuat akun dan memanggil `link`.

Response `201`:

```json
{
  "data": {
    "user": {
      "id": "uuid",
      "username": "alice",
      "display_name": "Alice",
      "avatar_url": null
    },
    "session": {
      "access_token": "opaque-token",
      "access_expires_at": "2026-08-19T10:15:00Z",
      "refresh_token": "opaque-refresh-token",
      "refresh_expires_at": "2026-09-18T10:00:00Z"
    },
    "tinode_login": {
      "scheme": "basic",
      "username": "alice"
    }
  }
}
```

Password tidak boleh dipantulkan kembali.

#### `POST /v1/auth/login`

- Input username/email dan password.
- Generic error untuk akun tidak ada atau password salah.
- Terapkan rate limit per IP dan per normalized username.
- Buat session baru jika valid.
- Tolak user `suspended` atau `deleted`.

#### `POST /v1/auth/refresh`

- Terima refresh token.
- Cari berdasarkan hash.
- Tolak expired/revoked.
- Rotate access dan refresh token dalam satu transaction.

#### `POST /v1/auth/logout`

- Memerlukan access token.
- Revoke session saat ini.
- Client juga wajib disconnect Tinode dan menghapus Tinode token lokal.

#### `GET /v1/me`

- Mengembalikan profile aplikasi user aktif.

#### `PATCH /v1/me`

- Hanya menerima field whitelist: `display_name`, `avatar_url` pada fase awal.
- Setelah update app DB, client mengupdate Tinode `me` public profile melalui SDK.
- Jika sinkronisasi Tinode gagal, app DB tetap source of truth dan client menampilkan opsi retry.

#### `POST /v1/reports`

- Memerlukan user aktif.
- Validate referensi user/topic/message.
- Rate limit.
- Tidak menerima arbitrary JSON.

### 10.3 Admin endpoints

Semua endpoint admin memerlukan role admin yang disimpan dan diverifikasi `app-api`. Jangan menjadikan semua Tinode root sebagai admin aplikasi.

- `GET /v1/admin/reports`
- `GET /v1/admin/reports/{id}`
- `PATCH /v1/admin/reports/{id}`
- `POST /v1/admin/users/{id}/suspend`
- `POST /v1/admin/users/{id}/unsuspend`

Suspension harus:

1. Mengubah status user di app DB.
2. Merevoke semua app sessions.
3. Membuat login berikutnya ditolak REST authenticator.
4. Menjadwalkan perubahan status akun Tinode melalui interface admin/gRPC resmi.
5. Menulis audit log.

### 10.4 Health endpoints

- `GET /health/live`: process hidup; tidak menyentuh dependency.
- `GET /health/ready`: cek app DB dan dependency internal kritis dengan timeout pendek.
- Response tidak membocorkan DSN atau secret.
- Endpoint detail metrics tidak public.

---

## 11. Tinode REST authenticator

Ini bagian paling sensitif dari integrasi akun. Implementasi wajib mengikuti dokumentasi resmi dan fixture test upstream, bukan tebakan.

### 11.1 Listener

- Jalankan internal listener terpisah, misalnya `:8081`.
- Listener hanya tersedia di Docker private network.
- Reverse proxy tidak boleh meneruskan `/internal/*`.
- Gunakan request body size limit, timeout, structured logging, dan request ID.

### 11.2 Endpoint resmi yang harus ditangani

Jika `use_separate_endpoints=true`, sediakan:

- `POST /internal/tinode/auth/add`
- `POST /internal/tinode/auth/auth`
- `POST /internal/tinode/auth/checkunique`
- `POST /internal/tinode/auth/del`
- `POST /internal/tinode/auth/gen`
- `POST /internal/tinode/auth/link`
- `POST /internal/tinode/auth/upd`
- `POST /internal/tinode/auth/rtagns`

Gunakan nama field request/response persis seperti dokumentasi Tinode.

### 11.3 Behavior minimum

#### `auth`

1. Decode `secret` base64 menjadi bentuk yang diharapkan login produk.
2. Parse username dan password dengan aman; username tidak boleh mengandung colon.
3. Normalize username.
4. Cari user aktif.
5. Compare bcrypt secara constant-time melalui library resmi.
6. Jika `tinode_uid_raw` sudah ada, return `rec` dengan UID dan auth level yang sesuai.
7. Jika belum ada, return `newacc` agar Tinode membuat akun pada first login.
8. Isi `public` hanya dengan field yang memang aman dibaca user lain.
9. Jangan log `secret`, password, atau response token.

#### `link`

1. Hubungkan UID yang dibuat Tinode ke user aplikasi.
2. Operasi harus idempotent.
3. UID tidak boleh dapat dimiliki dua user.
4. Gunakan transaction dan unique constraint.
5. Jika callback diulang dengan pasangan sama, return success.
6. Jika UID/user konflik, return error resmi Tinode dan audit event.

#### `checkunique`

- Return apakah username tersedia.
- Hasil harus konsisten dengan normalization endpoint register.

#### Endpoint lain

- Jika app backend tidak mendukung perubahan melalui Tinode, return error `unsupported` sesuai kontrak resmi.
- Jangan return HTTP success dengan payload buatan sendiri.
- `rtagns` harus mengikuti namespace yang benar-benar dibatasi aplikasi.

### 11.4 Test wajib authenticator

- [ ] Secret valid, user existing dan linked.
- [ ] Secret valid, user existing tetapi belum linked.
- [ ] Password salah.
- [ ] Username tidak ada.
- [ ] User suspended.
- [ ] User deleted.
- [ ] Base64 invalid.
- [ ] Format username/password invalid.
- [ ] Link pertama berhasil.
- [ ] Link retry idempotent.
- [ ] Link UID konflik ditolak.
- [ ] Concurrent first login hanya menghasilkan satu link.
- [ ] Payload terlalu besar ditolak.
- [ ] Unsupported endpoint mengembalikan error resmi.
- [ ] Log tidak mengandung password/secret.

---

## 12. Client web/PWA

### 12.1 Strategi implementasi

Urutan yang harus dipakai:

1. Jalankan reference web client Tinode terhadap server lokal untuk memahami behavior.
2. Pin commit SHA `tinode/webapp` dan `tinode-sdk` yang kompatibel.
3. Pilih salah satu:
   - fork reference client lalu lakukan perubahan kecil bertahap; atau
   - buat client baru menggunakan `tinode-sdk` jika reference client terlalu sulit dirawat.
4. Untuk tim kecil/AI murah, default adalah fork reference client terlebih dahulu. Jangan rewrite sampai alur lengkap terbukti.
5. Simpan patch branding/produk terpisah dan kecil agar update upstream masih mungkin.

### 12.2 State yang harus dipisah

```text
App auth state:
- app access token
- app refresh token
- current app user

Tinode state:
- Tinode instance
- connection status
- Tinode auth token + expiry
- topics/subscriptions
- messages

UI state:
- active route/topic
- draft
- modal/dialog
- upload progress
- call state
```

Jangan menduplikasi seluruh message list ke state manager buatan sendiri jika SDK sudah menjadi pemilik data dan cache.

### 12.3 Penyimpanan token

- Access token app sebaiknya hanya di memory jika arsitektur memungkinkan.
- Refresh token idealnya secure `HttpOnly`, `SameSite`, `Secure` cookie untuk web.
- Jika reference client memaksa penyimpanan berbeda, dokumentasikan risiko dan buat ADR.
- Tinode token tidak boleh ditulis ke log atau analytics.
- Logout menghapus semua cache/token lokal yang relevan.

### 12.4 Routes minimum

```text
/login
/register
/chat
/chat/:topicId
/contacts
/groups/new
/settings/profile
/settings/privacy
/reports/new
/admin/reports        (admin only)
```

Gunakan route guard. Jangan render halaman chat sebelum app session dan Tinode session siap.

### 12.5 Connection state machine

```text
idle
  -> connecting
  -> authenticating
  -> syncing
  -> ready

failure path:
connecting/authenticating/syncing
  -> offline or auth_error or fatal_error

recovery:
offline -> reconnecting with bounded exponential backoff -> syncing -> ready
```

Rules:

- Hanya satu active Tinode connection per browser tab.
- Backoff memiliki maximum delay dan jitter.
- Jangan infinite-loop pada invalid credentials.
- Ketika Tinode token expired, lakukan re-auth dengan flow yang didukung; jangan diam-diam meminta password tersimpan plaintext.
- UI harus membedakan offline, connecting, syncing, dan ready.

### 12.6 Screen acceptance criteria

#### Register

- Client dan server validation konsisten.
- Duplicate username/email terlihat jelas.
- Submit button disabled selama request aktif.
- Tidak ada double submit.
- Setelah register, lakukan app login dan first Tinode login.
- Jika Tinode first login gagal, akun aplikasi tetap ada dan UI menyediakan retry.

#### Conversation list

- Urut berdasarkan activity terbaru.
- Menampilkan avatar, display name, preview, timestamp, unread count.
- Update realtime tanpa full refresh.
- Pinned conversation mengikuti dukungan Tinode.
- Empty/loading/error/offline state tersedia.

#### Chat detail

- Pagination history.
- Optimistic sending dengan pending state.
- Failed send dapat di-retry tanpa duplicate.
- Sent/delivered/read tampil sesuai data SDK.
- Composer tidak mengirim whitespace-only.
- Enter behavior dan mobile keyboard diuji.
- Reply, forward, edit, pin mengikuti permission.
- Attachment menampilkan progress, cancel, error, dan retry.
- Scroll tidak melompat ketika history lama dimuat.
- Incoming message saat user membaca history tidak memaksa scroll ke bawah.

#### Group management

- Hanya role yang berhak dapat mengubah metadata/anggota.
- Permission error ditampilkan dengan benar.
- Member list memiliki pagination/search jika diperlukan.
- User yang dikeluarkan kehilangan akses sesuai permission Tinode.

### 12.7 Accessibility minimum

- Semua action dapat digunakan keyboard.
- Focus state terlihat.
- Icon-only button memiliki accessible label.
- Form error terhubung ke input.
- Color contrast memenuhi WCAG AA sebisa mungkin.
- Status pesan tidak hanya dibedakan dengan warna.
- Reduced motion dihormati.

---

## 13. Mapping fitur ke Tinode

AI wajib membaca dokumentasi SDK sebelum menggunakan method. Tabel ini menjelaskan konsep, bukan nama method final.

| Fitur produk | Konsep Tinode |
| --- | --- |
| Current user | topic `me` |
| User profile | `public`, `private`, `trusted` fields |
| Personal chat | P2P topic |
| Group | group topic |
| Channel | channel/read-only subscriber access, bukan MVP awal |
| Conversation list | subscriptions pada `me` |
| Message | `{pub}`/SDK publish ke topic |
| History | subscribe/get data dengan pagination sesuai SDK |
| Delivered/read | note/receipt sesuai SDK |
| Typing | key press/note sesuai SDK |
| Presence | presence notifications/subscription |
| Attachment | out-of-band upload kemudian reference URL dalam content |
| User discovery | tags/query resmi Tinode |
| Block | access control/deny sesuai SDK |
| Group role | access mode/permissions |
| Voice/video 1-to-1 | Tinode WebRTC signaling + ICE config |

Jangan membuat topic name sendiri dengan format yang terlihat benar. Gunakan helper SDK dan response server.

---

## 14. Media dan attachment

### 14.1 Development

- Gunakan media handler filesystem Tinode.
- Mount volume persisten.
- Batasi upload awal, misalnya:
  - image 10 MiB;
  - voice note 20 MiB;
  - video/file umum 50 MiB.
- Nilai final harus ditetapkan konsisten di reverse proxy, Tinode, dan UI.

### 14.2 Production

- Gunakan S3-compatible handler yang didukung Tinode.
- Bucket private.
- Download melalui signed/authorized URL sesuai handler.
- Terapkan content type allowlist dan jangan percaya filename.
- Randomize storage key.
- Set retention/orphan cleanup.
- Backup metadata dan object storage secara konsisten.

### 14.3 Upload flow

1. Client validate size dan type untuk UX saja.
2. Server tetap melakukan enforcement.
3. Upload melalui endpoint resmi Tinode.
4. Tangani redirect upload jika dikembalikan server.
5. Dapatkan URL dari response resmi.
6. Publish message dengan reference attachment dan `extra.attachments` bila disyaratkan.
7. Jika publish gagal, tandai/retry dan pastikan orphan cleanup.

### 14.4 Security

- Jangan render HTML attachment inline.
- Set `Content-Disposition` yang aman.
- SVG diperlakukan sebagai file download atau disanitasi ketat.
- Strip metadata gambar hanya jika requirement privasi disetujui.
- Antivirus scanning dapat ditambahkan setelah MVP; jangan klaim file aman sebelum ada scanning.

---

## 15. Voice/video call 1-to-1

Jangan mulai bagian ini sebelum milestone messaging, reconnect, backup, dan load test dasar selesai.

### 15.1 Komponen

- Tinode untuk signaling.
- Browser WebRTC untuk media peer-to-peer.
- coturn untuk STUN/TURN relay.
- HTTPS/WSS wajib di production.

### 15.2 Batas

- Hanya 1-to-1.
- Tidak ada group call.
- Tidak ada recording.
- Tidak ada PSTN/SIP.
- Tidak ada background blur pada MVP.

### 15.3 Checklist coturn

- [ ] Realm menggunakan domain sendiri.
- [ ] TLS/DTLS dipertimbangkan untuk production.
- [ ] Credential bukan hardcoded di client repository.
- [ ] Port UDP/TCP yang diperlukan dibuka.
- [ ] Relay diuji dari dua jaringan NAT berbeda.
- [ ] Bandwidth dan abuse limit dikonfigurasi.
- [ ] Log tidak menyimpan data sensitif berlebihan.
- [ ] TURN metrics/traffic dapat dipantau.

### 15.4 Acceptance criteria call

- Caller dapat start/cancel.
- Callee dapat accept/reject.
- Timeout jika tidak dijawab.
- Audio-only dan video dapat dipilih.
- Mic/camera permission denial ditangani.
- Device selection ditangani jika browser mendukung.
- Call berakhir bersih saat salah satu tab ditutup/disconnect.
- Repeated call tidak meninggalkan ghost state.
- Call bekerja saat kedua pihak berada di NAT berbeda melalui TURN.
- Incoming message tetap berfungsi selama call.

---

## 16. Security baseline

### 16.1 Authentication

- Password minimum 12 karakter untuk MVP; maksimum masuk akal agar tidak menjadi DoS input.
- Bcrypt cost minimum 12, ukur latency pada hardware target.
- Generic login failure.
- Rate limit login, register, refresh, dan forgot-password.
- Session rotation dan revocation.
- Tidak ada password di log, analytics, URL, atau error.

### 16.2 Network

- Public production hanya port 80/443 dan port TURN yang diperlukan.
- PostgreSQL, Tinode internal admin/gRPC, metrics, dan app internal listener tidak public.
- TLS minimum modern.
- Security headers: HSTS setelah domain stabil, CSP, `X-Content-Type-Options`, frame protection, referrer policy.
- WebSocket origin divalidasi.

### 16.3 Authorization

- Jangan percaya user ID dari body tanpa mencocokkannya dengan authenticated session.
- Permission group/chat selalu ditegakkan oleh Tinode, bukan hanya disembunyikan di UI.
- Admin endpoint memiliki middleware terpisah.
- Report tidak memberi reporter akses ke data privat target.

### 16.4 Abuse prevention

- Rate limit account creation, user search, invite, message publish jika tersedia, upload, dan report.
- Batasi jumlah hasil search.
- Batasi ukuran `public`/`private` payload.
- Reserved usernames: admin, root, support, system, moderator, tinode, dan variasinya.
- Block/suspend harus efektif pada login dan discovery.

### 16.5 Logging

Log boleh berisi:

- request ID;
- route template;
- status code;
- duration;
- internal user UUID bila perlu;
- Tinode topic/user ID bila perlu untuk diagnosis.

Log tidak boleh berisi:

- password;
- auth secret base64;
- app/Tinode raw token;
- cookie;
- Authorization header;
- isi pesan;
- file content;
- full email jika tidak perlu.

---

## 17. Testing strategy

### 17.1 Test pyramid

1. Unit test untuk validation, normalization, password, token, dan mapping error.
2. Integration test `app-api` dengan PostgreSQL nyata dalam container.
3. Contract test REST authenticator terhadap payload fixture Tinode.
4. End-to-end browser test untuk alur utama.
5. Smoke test deployment Compose.
6. Load test dan soak test Tinode.
7. Backup/restore drill.

### 17.2 Backend unit test minimum

- Username normalization/validation.
- Email normalization.
- Password length dan bcrypt compare.
- Access/refresh token generation dan hashing.
- Session expiry/revocation/rotation.
- User status transitions.
- Error response tidak membocorkan internal detail.
- Report validation.

### 17.3 Integration test minimum

- Register sukses.
- Duplicate username.
- Duplicate email.
- Concurrent duplicate register hanya satu sukses.
- Login benar/salah.
- Suspended login ditolak.
- Refresh rotation.
- Reuse refresh token lama ditolak.
- Logout.
- Update profile.
- Membuat report.
- Admin suspend dan audit log.
- Seluruh test REST authenticator pada bagian sebelumnya.

### 17.4 E2E utama

Scenario A — first login:

1. User Alice register.
2. Alice login app.
3. Alice login Tinode melalui REST authenticator.
4. Tinode account dibuat dan link tersimpan.
5. Reload browser.
6. Session dipulihkan tanpa membuat akun Tinode baru.

Scenario B — personal chat:

1. Alice dan Bob login pada browser berbeda.
2. Alice menemukan Bob.
3. Alice memulai chat.
4. Alice mengirim teks.
5. Bob menerima realtime.
6. Receipt berubah sampai read.
7. Bob membalas.
8. Kedua browser reload dan history tetap ada.

Scenario C — offline:

1. Bob offline.
2. Alice mengirim tiga pesan.
3. Bob login kembali.
4. Pesan diterima urut, tanpa hilang/duplicate.
5. Unread count benar.

Scenario D — group:

1. Alice membuat grup dan mengundang Bob/Charlie.
2. Semua menerima pesan.
3. Permission perubahan metadata diuji.
4. Charlie dikeluarkan.
5. Charlie tidak dapat mengakses pesan baru.

Scenario E — attachment:

1. Upload gambar valid.
2. Progress terlihat.
3. Bob menerima dan dapat membuka.
4. File terlalu besar ditolak.
5. SVG/HTML berbahaya tidak dieksekusi inline.

Scenario F — reconnect:

1. Putuskan jaringan Alice.
2. UI menunjukkan offline.
3. Pulihkan jaringan.
4. Client reconnect dan sync.
5. Pesan selama offline muncul sekali dan berurutan.

### 17.5 Target load awal

Ini adalah baseline engineering, bukan janji SLA:

- 8.000 registered users.
- 500 concurrent WebSocket connections.
- 50 messages/second selama 10 menit.
- Burst 150 messages/second selama 60 detik.
- P95 server acknowledgement di bawah 500 ms pada satu region dan beban baseline.
- Tidak ada message loss.
- Tidak ada duplicate akibat retry normal.
- Error rate di bawah 1% dan penyebab seluruh error diketahui.
- CPU, memory, disk latency, DB connection pool, dan network direkam.

Jika gagal, jangan langsung mengaktifkan cluster. Profil bottleneck terlebih dahulu.

---

## 18. Observability

### 18.1 Structured log fields

```text
timestamp
level
service
environment
request_id
trace_id (jika tersedia)
route
method
status
duration_ms
user_id (opsional)
error_code (opsional)
```

### 18.2 Metrics minimum

`app-api`:

- request count/latency/error by route;
- active sessions;
- login success/failure;
- register success/conflict;
- REST auth request count/latency/error per endpoint;
- DB pool usage;
- report backlog.

Infrastructure:

- Tinode connections/messages/errors;
- PostgreSQL CPU, memory, connections, locks, slow queries, disk;
- host CPU, RAM, disk utilization/inode, network;
- reverse proxy request/WebSocket errors;
- coturn sessions dan bandwidth.

### 18.3 Alerts minimum

- Health check gagal lebih dari 2 menit.
- Disk di atas 80%.
- Backup gagal.
- Error rate login/auth meningkat abnormal.
- PostgreSQL connection pool hampir habis.
- WebSocket disconnect spike.
- Certificate mendekati expiry.
- TURN bandwidth abnormal.

---

## 19. Backup dan disaster recovery

### 19.1 Yang harus dibackup

- Database `tinode`.
- Database `app_chat`.
- Media filesystem atau object storage.
- Tinode config tanpa menaruh secret di artifact tak terenkripsi.
- Reverse proxy/coturn config.
- Secret inventory dan recovery procedure di secret manager/offline secure storage.

### 19.2 Jadwal awal

- PostgreSQL logical backup harian.
- Snapshot volume sesuai kemampuan provider.
- Media backup harian/incremental.
- Retention minimal 7 daily + 4 weekly untuk MVP, disesuaikan kebijakan data.
- Backup terenkripsi dan disimpan di failure domain berbeda.

### 19.3 Restore drill

Restore dianggap valid hanya jika:

1. Environment kosong dibuat.
2. App DB direstore.
3. Tinode DB direstore.
4. Media direstore.
5. Dua user dapat login.
6. Conversation/history dan attachment tersedia.
7. Pesan baru dapat dikirim.
8. Hasil dan waktu restore dicatat.

Target awal yang harus divalidasi, bukan dijanjikan:

- RPO: 24 jam.
- RTO: 4 jam.

---

## 20. Deployment environments

### 20.1 Development

- Docker Compose.
- HTTP diperbolehkan hanya localhost.
- Sample data default tetap dimatikan; fixture test dibuat sendiri.
- Database/media volume persisten tetapi boleh direset dengan perintah eksplisit.
- WebRTC default off.

### 20.2 Staging

- Domain dan TLS nyata.
- Konfigurasi sedekat mungkin production.
- Data sintetis, tidak memakai dump production mentah.
- Tempat E2E, load test terbatas, migration rehearsal, backup/restore.

### 20.3 Production

- Semua image pin exact version/digest.
- TLS dan security headers.
- Secret eksternal.
- Daily backup dan alerts.
- Database tidak public.
- Least-privilege DB roles.
- Resource limit container.
- Restart policy.
- Deployment dan rollback terdokumentasi.

### 20.4 Upgrade Tinode

1. Baca changelog dan migration notes.
2. Pin candidate version di staging.
3. Backup staging.
4. Upgrade dengan tool resmi.
5. Jalankan seluruh smoke/E2E.
6. Jalankan load test pendek.
7. Backup production.
8. Deploy pada maintenance window jika schema berubah.
9. Verify health, login, sync, send, attachment.
10. Rollback hanya sesuai compatibility schema yang terdokumentasi; jangan restore sebagian.

---

## 21. Milestone pengerjaan

Satu milestone tidak boleh dimulai sebelum exit criteria milestone sebelumnya terpenuhi.

### M0 — Repository dan keputusan dasar

Tasks:

- [x] Buat `README.md` singkat yang menunjuk dokumen ini.
- [x] Buat struktur minimum `infra/compose`, `infra/tinode`, dan `docs`.
- [x] Buat `.gitignore` dan `.env.example`.
- [x] Catat versi/commit upstream di `docs/UPSTREAM_VERSIONS.md`.
- [x] Buat `THIRD_PARTY_NOTICES.md` awal.
- [x] Buat ADR pemilihan Tinode, PostgreSQL, dan web-first.

Exit criteria:

- Repository tidak mengandung secret.
- Versi upstream exact tercatat.
- Tidak ada kode aplikasi sebelum keputusan dasar tercatat.

### M1 — Tinode local smoke environment

Tasks:

- [x] Buat Compose PostgreSQL + Tinode.
- [x] Pisahkan database target `tinode`/fallback role `tinode_admin` dan app database/role `app_chat`.
- [x] Buat Tinode config development tanpa sample data.
- [x] Mount volume DB dan media.
- [x] Jalankan reference web client Tinode dari static client image.
- [x] Buat dua akun test (`alice01` dan `bob01`) pada app-api dan Tinode.
- [x] Uji personal chat dua arah dan receipt dasar secara manual melalui dua tab reference client.
- [ ] Uji group, attachment, dan reconnect manual.
- [x] Buat `scripts/smoke-test.sh` untuk health dasar.

Exit criteria:

- Environment hidup dari checkout bersih dengan instruksi yang terdokumentasi.
- Restart container tidak menghilangkan account/history/media.
- [x] Dua user dapat chat lewat reference client.

### M2 — Skeleton `app-api`

Tasks:

- [x] Inisialisasi Go module.
- [x] Buat config loader dengan validation fail-fast.
- [x] Buat structured logging.
- [x] Buat public/internal HTTP listener.
- [x] Buat live/ready health endpoint.
- [x] Buat PostgreSQL connection pool.
- [x] Buat migration runner sederhana yang dipilih dan didokumentasikan.
- [x] Tambahkan graceful shutdown dan timeout.
- [x] Tambahkan unit/integration test harness.

Exit criteria:

- `go test ./...` lulus.
- Process gagal dengan pesan jelas jika config wajib kosong.
- Graceful shutdown menutup HTTP dan DB connection.
- Internal listener tidak terpublish ke host/public.

### M3 — App auth dan session

Tasks:

- [x] Migration `users` dan `sessions`.
- [x] Username/email normalization.
- [x] Password hashing bcrypt.
- [x] Register endpoint.
- [x] Login endpoint.
- [x] Opaque access/refresh tokens.
- [x] Refresh rotation.
- [x] Logout/revoke.
- [x] Auth middleware.
- [x] Rate limit auth endpoints.
- [x] Unit tests dan runtime integration smoke test dasar.

Exit criteria:

- Tidak ada raw token/password di database/log.
- Concurrent duplicate registration aman.
- Reuse refresh token lama ditolak.
- Suspended/deleted account tidak login.

### M4 — Tinode REST auth integration

Tasks:

- [x] Implement payload types sesuai upstream.
- [x] Implement `auth`.
- [x] Implement idempotent `link`.
- [x] Implement `checkunique`.
- [x] Implement behavior resmi endpoint lain.
- [x] Configure `logical_names` untuk map `basic` ke `rest` pada Compose external config.
- [x] Buat unit/contract tests dasar.
- [x] Uji first login dan repeated login path secara runtime terhadap internal listener.
- [ ] Uji concurrency first login.

Exit criteria:

- App DB menjadi source of truth password.
- Tinode account otomatis dibuat pada first login.
- Mapping user hanya satu dan stabil.
- REST auth tidak dapat diakses dari public network.

### M5 — Web client foundation

Tasks:

- [x] Bootstrap custom React/Vite client dengan `tinode-sdk@0.25.3`.
- [x] Pertahankan Apache license/notice pada dependency inventory.
- [x] Tambahkan app register/login/logout.
- [x] Implement dual session bootstrap: app + Tinode token.
- [x] Implement route guards untuk auth workspace.
- [x] Implement connection state UI.
- [x] Ganti branding/aset reference dengan aset BashoCode sendiri.
- [x] Tambahkan Indonesian/English strings.
- [x] Tambahkan error boundary dan offline state.

Exit criteria:

- [x] User register lalu masuk ke authenticated workspace tanpa langkah manual.
- [x] Reload memulihkan app session dan Tinode token yang valid.
- [x] Logout menghapus kedua session.
- [x] Tidak ada upstream logo/nama yang tertinggal kecuali attribution legal.

### M6 — Messaging core

Tasks:

- [ ] Conversation list.
- [ ] User discovery/contact flow.
- [ ] Personal chat.
- [ ] Group create/manage.
- [ ] Text messages.
- [ ] History pagination.
- [ ] Delivery/read receipts.
- [ ] Typing/presence.
- [ ] Reply/forward/edit/pin berdasarkan kemampuan upstream.
- [ ] Reconnect dan deduplication behavior.
- [ ] E2E scenarios A–D dan F.

Exit criteria:

- Dua browser dapat menyelesaikan seluruh E2E chat scenario.
- Offline/reconnect tidak menyebabkan pesan hilang/duplicate.
- Permission group ditegakkan server.

### M7 — Media dan voice note

Tasks:

- [ ] Attachment upload UI.
- [ ] Size/type limits konsisten.
- [ ] Image preview aman.
- [ ] File download aman.
- [ ] Voice note record/playback.
- [ ] Upload cancel/retry.
- [ ] Orphan cleanup diverifikasi.
- [ ] E2E attachment scenario.
- [ ] Siapkan migration plan filesystem ke S3-compatible.

Exit criteria:

- File valid terkirim dan bertahan setelah restart.
- File invalid/terlalu besar ditolak server-side.
- Tidak ada active content tak tepercaya dieksekusi.

### M8 — Moderation dan admin minimum

Tasks:

- [ ] Migration reports/audit logs/admin role.
- [ ] Report endpoint dan UI.
- [ ] Admin report list/detail/resolution.
- [ ] Suspend/unsuspend.
- [ ] Session revocation.
- [ ] Tinode account state sync.
- [ ] Audit log.

Exit criteria:

- Suspended user kehilangan session dan tidak dapat login ulang.
- Semua admin action tercatat.
- Admin UI tidak dapat dibuka user biasa.

### M9 — Production readiness

Tasks:

- [ ] Reverse proxy TLS.
- [ ] Production secret/config validation.
- [ ] Security headers dan origin restriction.
- [ ] Metrics/dashboard/alerts.
- [ ] Backup script.
- [ ] Restore drill.
- [ ] Load test baseline.
- [ ] Dependency/license scan.
- [ ] Security review.
- [ ] Deployment dan rollback runbook.

Exit criteria:

- Load target tercapai atau kapasitas nyata terdokumentasi.
- Restore berhasil.
- Critical/high security finding nol atau diberi blocking mitigation.
- Rollback rehearsal dilakukan.

### M10 — Voice/video 1-to-1

Tasks:

- [ ] Deploy coturn staging.
- [ ] Configure ICE file Tinode.
- [ ] Enable WebRTC staging.
- [ ] Implement call UI/state.
- [ ] Test permission/network failure.
- [ ] Test cross-NAT relay.
- [ ] Tambahkan TURN metrics/limits.
- [ ] Dokumentasikan bandwidth cost.

Exit criteria:

- Semua acceptance criteria call lulus di dua jaringan berbeda.
- Tidak ada credential production hardcoded di client/Git.
- Chat tetap stabil saat call.

### M11 — Native mobile (setelah MVP web)

Android dan iOS adalah workstream terpisah. Jangan memaksakan `tinode-sdk` browser ke React Native tanpa bukti kompatibilitas. Gunakan SDK native resmi Tinode atau binding yang secara resmi didukung.

Tasks Android:

- [ ] Pin SDK/client Android upstream.
- [ ] Pertahankan license/notice.
- [ ] Implement app auth.
- [ ] Implement Tinode session.
- [ ] Port UX messaging yang sudah stabil.
- [ ] FCM push.
- [ ] Secure local token storage.
- [ ] Background/reconnect testing.

Tasks iOS:

- [ ] Pin SDK/client iOS upstream.
- [ ] Pertahankan license/notice.
- [ ] Implement app auth.
- [ ] Implement Tinode session.
- [ ] Port UX messaging yang sudah stabil.
- [ ] APNs push.
- [ ] Keychain token storage.
- [ ] Background/reconnect testing.

---

## 22. Definition of Done per task

Sebuah task hanya selesai jika seluruh poin berikut terpenuhi:

- [ ] Perubahan hanya terkait task.
- [ ] Code mengikuti convention existing.
- [ ] Tidak ada secret atau debug credential.
- [ ] Error path ditangani.
- [ ] Test baru ditambahkan untuk behavior baru.
- [ ] Seluruh test relevan lulus.
- [ ] Lint/format lulus.
- [ ] Dokumentasi/config example diperbarui jika kontrak berubah.
- [ ] Tidak ada TODO tanpa issue/task referensi.
- [ ] Tidak ada dead code atau commented-out implementation.
- [ ] Acceptance criteria task dapat didemonstrasikan.
- [ ] Progress log diperbarui.

---

## 23. Aturan perubahan dependency

Sebelum menambah dependency:

1. Tunjukkan kebutuhan yang tidak layak diimplementasikan dengan standard library/dependency existing.
2. Periksa maintenance, release terakhir, security advisory, dan license.
3. Pilih dependency terkecil.
4. Pin versi melalui lockfile/go modules.
5. Tambahkan ke `THIRD_PARTY_NOTICES.md` jika perlu.
6. Tambahkan test untuk integration boundary.

Dependency yang diperkirakan memang diperlukan:

- PostgreSQL driver/pool Go.
- Migration tool/library atau migration runner yang sederhana.
- `golang.org/x/crypto/bcrypt`.
- Router/middleware hanya jika standard library menjadi tidak proporsional.
- `tinode-sdk` pada web client.

Jangan menambah ORM besar, message broker, Redis, Kafka, Elasticsearch, Kubernetes operator, atau observability framework penuh pada MVP tanpa bukti kebutuhan.

---

## 24. Stop conditions untuk AI

AI harus berhenti dan meminta keputusan manusia jika:

- Versi Tinode yang dipin tidak tersedia atau API berbeda dari dokumen.
- Implementasi memerlukan modifikasi source Tinode server.
- Ditemukan dependency GPL/AGPL baru pada kode proprietary.
- Harus memilih framework client baru.
- Ada kebutuhan E2EE.
- Ada kebutuhan group call atau federation.
- Migration berisiko menghapus/mengubah data existing.
- Harus membuka PostgreSQL/internal endpoint ke internet.
- Hasil load test jauh di bawah target dan penyebab belum diketahui.
- Restore backup gagal.
- Requirement membutuhkan pembayaran, SMS, atau external paid service.
- Ada ambiguity yang mengubah model data atau public API.

---

## 25. Template prompt task untuk AI murah

Gunakan template berikut untuk memberi task kecil:

```text
Baca TINODE_IMPLEMENTATION_PLAYBOOK.md, terutama bagian [nomor bagian].

Kerjakan hanya task ini:
[satu checkbox task]

Sebelum edit:
1. Periksa file terkait.
2. Sebutkan asumsi.
3. Sebutkan file yang akan diubah.

Batas:
- Jangan mengerjakan task berikutnya.
- Jangan refactor bagian lain.
- Jangan menambah dependency kecuali wajib; jelaskan jika wajib.
- Jangan mengubah source Tinode.
- Jangan menulis langsung database Tinode.

Acceptance criteria:
[salin acceptance criteria spesifik]

Setelah edit:
1. Jalankan format/lint/test relevan.
2. Laporkan command dan hasil.
3. Update Progress log di playbook.
```

Contoh task yang cukup kecil:

```text
Implementasikan hanya username normalization dan validation pada app-api.

Acceptance criteria:
- menerima lowercase a-z, digit, titik, underscore;
- panjang 4–32 rune/karakter ASCII;
- uppercase dinormalisasi lowercase;
- spasi dan simbol lain ditolak;
- tidak boleh titik di awal/akhir;
- tidak boleh dua titik berurutan;
- reserved usernames ditolak;
- table-driven unit tests lengkap;
- belum membuat HTTP endpoint atau database migration.
```

Contoh task yang terlalu besar dan harus dipecah:

```text
Buat backend dan frontend WhatsApp clone lengkap.
```

---

## 26. Risiko utama

| Risiko | Dampak | Mitigasi |
| --- | --- | --- |
| Tinode masih menyebut dirinya beta-quality | Bug/missing feature | PoC, pin version, E2E/load/soak test |
| Server GPLv3 | Distribusi/on-prem complicated | Jangan modifikasi/distribusikan tanpa legal review |
| Tidak ada E2EE stabil | Operator server dapat mengakses pesan | Transparansi produk; jangan klaim E2EE; evaluasi ulang sebelum use case sensitif |
| Group call/federation belum tersedia | Gap terhadap Telegram/WhatsApp | Tegaskan non-goal MVP; gunakan solusi lain jika menjadi hard requirement |
| Dual account/session app dan Tinode | Login/recovery lebih kompleks | State machine eksplisit, REST auth, idempotent link, test concurrency |
| Ketergantungan reference client | Upgrade conflict | Patch kecil, pin commit, catat deviations |
| TURN bandwidth | Biaya besar/abuse | 1-to-1 only, limits, metrics, rate controls |
| Database/media backup tidak konsisten | Attachment/history tidak sinkron | Restore drill dan runbook |
| Cheap AI memperlebar scope | Kode tidak terkontrol | Task checkbox kecil, stop conditions, Definition of Done |

---

## 27. Kriteria keputusan lanjut atau pindah stack

Setelah M6 dan load test awal, lakukan review. Lanjut menggunakan Tinode hanya jika:

- REST auth stabil dan mudah dioperasikan;
- web client dapat dikustomisasi tanpa patch upstream yang terlalu besar;
- message sync/reconnect bekerja pada kondisi jaringan buruk;
- load baseline memenuhi kebutuhan dengan margin aman;
- tidak ada blocker license untuk model bisnis;
- E2EE bukan hard requirement saat launch;
- group call/federation bukan hard requirement.

Pertimbangkan pindah sebelum investasi mobile besar jika salah satu kondisi berikut terjadi:

- E2EE menjadi requirement wajib;
- pelanggan meminta distribusi on-premise proprietary;
- group call menjadi core feature;
- custom patch server semakin banyak;
- update upstream sulit diadopsi;
- hasil load/soak test tidak memenuhi kebutuhan dan perbaikan memerlukan fork server besar.

---

## 28. Progress log

Tambahkan entry baru di bagian paling atas. Jangan menghapus history.

### 2026-08-19 — M5 custom web client foundation

- Menambahkan `apps/web` React/Vite dengan `tinode-sdk@0.25.3` yang dipin.
- Menambahkan auth page custom, register/login/logout, route guard berbasis app session, dan dual bootstrap menggunakan token app + token Tinode.
- Menambahkan status koneksi Tinode, indikator offline, error boundary, serta strings Indonesia/English.
- Mengganti tampilan reference client dengan branding BashoCode sendiri; conversation list/composer tetap menjadi pekerjaan M6.
- Menambahkan CORS exact-origin terkontrol pada `app-api` untuk `localhost:3000`.
- Verifikasi browser lulus: login Alice, koneksi Tinode, reload memulihkan session, pergantian bahasa, dan logout mengembalikan route login.
- `npm run build`, `npm audit --omit=dev`, dan `go test ./...` lulus.

### 2026-08-19 — M1 browser/WebSocket acceptance

- Membuat fixture `alice01` dan `bob01` dengan password development-only yang tidak disimpan di repository.
- Login kedua akun melalui dua tab reference Tinode Web `TinodeWeb/0.25.3` pada `localhost:6060`.
- Verifikasi first-login provisioning dan callback `link` menghasilkan mapping UID Tinode yang stabil di app DB.
- Alice memulai personal chat dengan Bob memakai public ID Tinode; pesan dua arah diterima di kedua tab.
- Receipt dasar terlihat pada UI (`done`/`done_all`).
- M1 masih menunggu group, attachment, dan reconnect manual.

### 2026-08-19 — M2 app-api skeleton

- Menambahkan Go module `services/app-api` dengan driver PostgreSQL `pgx` yang dipin.
- Menambahkan config loader fail-fast, structured JSON logging, public listener `:8080`, dan internal listener `:8081`.
- Menambahkan `/health/live` dan `/health/ready`, migration runner embedded, dan graceful shutdown.
- `go test ./...` lulus.
- Runtime test lulus: migration `0001_m2_foundation.sql` tercatat, ketiga health endpoint mengembalikan `{"status":"ok"}`, dan shutdown menerima SIGINT.
- Next: lanjutkan M3 auth/session dan tutup sisa acceptance M1 (group, attachment, reconnect).

### 2026-08-19 — M3 auth/session core

- Menambahkan migration `users`/`sessions`, normalisasi username/email, validasi password/display name, dan bcrypt.
- Menambahkan register/login/refresh/logout/`/v1/me` dengan opaque URL-safe token; database hanya menyimpan SHA-256 token hash.
- Menambahkan refresh rotation, revoke, auth middleware, request body limit, dan in-memory rate limit auth endpoint.
- `go test ./...` lulus.
- Runtime smoke lulus untuk register, login, refresh rotation, `/v1/me`, logout, dan penolakan access token yang sudah direvoke.
- M3 exit masih menunggu test concurrency duplicate registration, refresh-token reuse, serta integrasi status suspended/deleted dari moderation M8.

### 2026-08-19 — M4 REST authenticator core

- Menambahkan payload/response resmi Tinode REST authenticator pada internal listener `POST /internal/tinode/auth/{endpoint}`.
- `auth` memverifikasi bcrypt dari app DB dan mengembalikan `newacc` untuk user yang belum linked, lalu `rec` untuk user linked.
- `link` menyimpan UID Tinode dengan unique constraint dan idempotency; `checkunique` memakai normalisasi username yang sama dengan register.
- Endpoint `add/del/gen/upd` mengembalikan `unsupported`, `rtagns` mengembalikan namespace terbatas.
- Menambahkan config example `logical_names: ["basic:rest", "rest:"]` dan `server_url` internal.
- `go test ./...` lulus; runtime auth/checkunique/link/retry/transition lulus.
- M4 Compose wiring aktif: Tinode memuat external config dan dapat menjangkau `app-api:8081` melalui private network; runtime `auth` dari dalam Tinode container mengembalikan `newacc` untuk user baru.
- M4 exit masih menunggu first-login concurrency test.

### 2026-08-19 — M0/M1 foundation

- Membuat repository foundation, `.gitignore`, `.env.example`, README, upstream inventory, third-party notices, dan ADR.
- Membuat Compose PostgreSQL + Tinode `0.25.3` dengan database target `tinode`, fallback role/database `tinode_admin`, dan app database/role `app_chat` terpisah.
- Menambahkan volume PostgreSQL/media dan smoke script.
- Validasi shell dan `docker compose config` lulus.
- Docker Desktop aktif dan image `tinode/tinode:0.25.3` serta `postgres:16-alpine` berhasil dipull.
- Bootstrap pertama menemukan perilaku resmi `init-db`: Tinode membuat database sendiri dan membutuhkan role `CREATEDB`; init script diperbaiki untuk mengikuti itu.
- Fallback database `tinode_admin` membuat Tinode target database `tinode`; kedua service healthy dan HTTP smoke test lulus.
- Next: uji group, attachment, dan reconnect manual untuk menutup seluruh exit criteria M1.

---

## 29. Final release checklist MVP

Product:

- [ ] Register/login/logout/recovery bekerja.
- [ ] Personal chat dan group bekerja.
- [ ] History, unread, receipt, typing, presence bekerja.
- [ ] Attachment dan voice note bekerja.
- [ ] Block/report/suspend bekerja.
- [ ] Offline/reconnect diuji.
- [ ] User-facing disclosure bahwa belum E2EE tersedia.

Security:

- [ ] TLS production valid.
- [ ] Default secrets diganti.
- [ ] Internal ports tidak public.
- [ ] Auth/upload/report rate limit.
- [ ] No secrets atau message bodies di log.
- [ ] Dependency scan bersih dari critical issue.
- [ ] License/attribution review selesai.

Operations:

- [ ] Dashboard dan alert aktif.
- [ ] Backup harian aktif.
- [ ] Restore drill berhasil.
- [ ] Load test terdokumentasi.
- [ ] Deployment dan rollback diuji.
- [ ] Tinode/app/media versions dipin.
- [ ] Incident contact dan procedure tersedia.

Release:

- [ ] Privacy policy sesuai data yang benar-benar dikumpulkan.
- [ ] Terms dan acceptable-use policy tersedia.
- [ ] Data deletion/suspension procedure tersedia.
- [ ] Brand/aset bukan milik platform lain.
- [ ] Known limitations dipublikasikan.

---

## 30. Ringkasan satu halaman

Jika AI hanya mampu memahami satu bagian, gunakan ini:

1. Tinode adalah mesin chat; jangan ubah source dan jangan tulis tabelnya.
2. PostgreSQL memiliki database/user terpisah untuk Tinode dan aplikasi.
3. Go `app-api` memiliki akun/password/session/report/admin.
4. Tinode memverifikasi password ke internal REST authenticator `app-api`.
5. First Tinode login membuat account lalu callback `link` menyimpan mapping UID secara idempotent.
6. Web client memakai `tinode-sdk`; mulai dari reference client dan lakukan patch kecil.
7. Chat data tetap di Tinode; jangan diduplikasi.
8. MVP tidak memiliki E2EE, group call, federation, payment, Moments, atau mini-app.
9. Voice/video 1-to-1 dan coturn dikerjakan terakhir.
10. Kerjakan milestone berurutan, satu checkbox per task, selalu test, dan jangan memperlebar scope.
