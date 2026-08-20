# OpenIM Chat Product — Implementation Playbook

Status: **planning source of truth**  
Target: MVP self-hosted chat untuk 5.000–8.000 akun  
Primary client: responsive web/PWA, kemudian mobile  
Business backend: Go  
Messaging engine: OpenIM  
Last reviewed: 2026-08-20

---

## 0. Cara menggunakan dokumen ini

Dokumen ini adalah urutan kerja untuk membuat PoC OpenIM dan memutuskan apakah OpenIM layak menjadi fondasi produk proprietary.

Aturan pengerjaan:

1. Kerjakan satu task checkbox dalam satu waktu.
2. Jangan memodifikasi source OpenIM Server atau SDK Core sebelum ada kebutuhan yang dibuktikan oleh PoC.
3. Pin semua image, SDK, dan repository upstream ke versi atau commit yang eksplisit; jangan gunakan `latest` untuk production.
4. Jangan membaca atau menulis langsung database internal OpenIM dari service aplikasi.
5. Business logic produk tetap berada di service milik kita.
6. Jangan menambahkan fitur di luar milestone aktif.
7. Jangan menyimpan secret, token, password, private key, atau kredensial nyata di Git.
8. Jalankan test dan catat hasilnya sebelum menandai task selesai.
9. Jika dokumentasi upstream bertentangan dengan dokumen ini, hentikan task dan perbarui keputusan berdasarkan bukti terbaru.
10. Rilis proprietary tidak boleh dilakukan sebelum **License Gate** pada bagian 2 selesai.

Format laporan setelah menyelesaikan task:

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

Membangun aplikasi chat bermerek sendiri dengan kemampuan dasar seperti WhatsApp/Telegram:

- akun dan profil pengguna;
- pencarian dan kontak;
- chat personal;
- grup;
- pesan teks;
- gambar, file, video pendek, dan voice note;
- status sent/delivered/read;
- typing indicator dan presence;
- reply, forward, edit, delete, dan pin jika didukung kontrak OpenIM yang dipilih;
- sinkronisasi multi-device;
- block dan report;
- push notification;
- voice/video call setelah messaging stabil;
- self-hosted tanpa ketergantungan pada layanan OpenIM Cloud.

OpenIM dipakai sebagai messaging engine. Branding, autentikasi produk, business API, moderasi, audit, dan UI tetap milik kita.

---

## 2. License Gate — wajib selesai sebelum production

### 2.1 Fakta yang sudah diverifikasi

- OpenIM Server dan OpenIM SDK Core adalah komponen yang lisensinya harus diperiksa secara terpisah.
- Repository `openim-sdk-core` menggunakan model dual-license: AGPL-3.0-or-later **atau** commercial license dari OpenIMSDK.
- Jalur Open Source IMSDK ditampilkan sebagai gratis, tetapi tetap tunduk pada AGPL.
- Harga commercial license tidak dipublikasikan.
- Business menggunakan commercial license dan Enterprise menggunakan custom plan.
- Permintaan komersial diarahkan ke `contact@openim.io`.

Sumber resmi:

- OpenIM SDK Core: <https://github.com/openimsdk/openim-sdk-core>
- OpenIM Server: <https://github.com/openimsdk/open-im-server>
- OpenIM Enterprise: <https://openim.io/enterprise/>
- OpenIM documentation: <https://docs.openim.io/>

Dokumen ini bukan nasihat hukum. Interpretasi kewajiban AGPL dan commercial license harus disetujui oleh pihak legal perusahaan.

### 2.2 Keputusan yang harus diperoleh

- [ ] Tentukan apakah aplikasi akan closed-source/proprietary.
- [ ] Minta legal memeriksa lisensi pada versi OpenIM Server, SDK Core, SDK platform, demo client, dan plugin yang akan dipakai.
- [ ] Minta quotation commercial license untuk 5.000–8.000 akun.
- [ ] Pastikan apakah lisensi dihitung per perusahaan, aplikasi, deployment, domain, server, MAU, atau registered user.
- [ ] Pastikan apakah commercial license SDK Core dapat dibeli tanpa Business client UI.
- [ ] Pastikan apakah staging, disaster recovery, dan multiple environment memerlukan lisensi tambahan.
- [ ] Pastikan apakah update, security patch, dan support termasuk harga.
- [ ] Pastikan apakah modifikasi SDK Core diperbolehkan dan bagaimana hak distribusinya.
- [ ] Dokumentasikan keputusan final dalam `openim/LICENSE_DECISION.md`.

### 2.3 Aturan sebelum keputusan lisensi

Yang boleh dilakukan:

- PoC internal dengan data sintetis;
- menjalankan deployment lokal;
- menguji API dan SDK;
- mengukur kelengkapan fitur dan performa;
- membuat adapter business backend yang tidak menyalin kode OpenIM.

Yang belum boleh dilakukan:

- merilis aplikasi closed-source kepada pengguna eksternal;
- menyatakan OpenIM bebas kewajiban lisensi;
- menggabungkan atau menyalin kode AGPL ke service proprietary;
- membeli atau menyetujui kontrak tanpa persetujuan perusahaan;
- menghapus attribution atau license notice upstream.

---

## 3. Keputusan awal

Keputusan ini berlaku untuk PoC dan dapat diubah hanya setelah hasil milestone terdokumentasi.

| Area | Keputusan awal |
| --- | --- |
| Messaging engine | OpenIM self-hosted |
| Source upstream | Gunakan tanpa modifikasi untuk PoC |
| Versi | Ditentukan dan dipin pada M0 setelah verifikasi release stabil terbaru |
| Business backend | Go modular monolith bernama `app-api` |
| Auth source of truth | `app-api`; integrasi OpenIM menggunakan Server API/token resmi |
| Client pertama | Web/PWA untuk validasi kontrak dan UX |
| Mobile | Setelah chat web dan kontrak backend stabil |
| Deployment PoC | Docker Compose satu host |
| Database/infrastruktur | Gunakan komponen resmi yang disyaratkan versi OpenIM terpilih |
| Media | Object storage lokal/S3-compatible; jangan simpan blob besar di app DB |
| Voice/video | Milestone terpisah setelah messaging core lulus |
| Scaling | Load test lebih dulu; jangan mulai dengan Kubernetes |
| API aplikasi | JSON REST dengan prefix `/v1` |
| ID aplikasi | UUID acak |
| Waktu | Simpan UTC; konversi timezone hanya di UI |

---

## 4. Batas scope

### 4.1 Wajib masuk MVP

- [ ] Local development hidup dengan satu perintah terdokumentasi.
- [ ] Register, login, logout, dan refresh session.
- [ ] Mapping user aplikasi ke user OpenIM.
- [ ] Profil dan avatar.
- [ ] Pencarian pengguna/kontak.
- [ ] Chat personal.
- [ ] Grup dan pengelolaan anggota.
- [ ] Pesan teks.
- [ ] Gambar, file, dan voice note.
- [ ] Sent/delivered/read state.
- [ ] Typing indicator dan presence.
- [ ] Reply, forward, edit, delete, dan pin sesuai dukungan versi terpilih.
- [ ] Conversation list dan unread count.
- [ ] Reconnect setelah jaringan putus.
- [ ] Sinkronisasi multi-device.
- [ ] Block user.
- [ ] Report user/message ke `app-api`.
- [ ] Suspend user oleh admin.
- [ ] Push notification.
- [ ] Backup dan restore teruji.
- [ ] Load test terhadap profil 5.000–8.000 akun.
- [ ] Voice/video 1-to-1 setelah chat core stabil.

### 4.2 Di luar MVP

- Moments/status feed;
- payment/wallet;
- mini-app;
- federation;
- group video meeting;
- bot marketplace;
- SMS OTP;
- import kontak telepon;
- native desktop;
- Kubernetes;
- multi-region;
- microservices;
- fork besar OpenIM Server/SDK;
- menyalin merek, logo, ikon, atau tampilan WhatsApp, Telegram, WeChat, dan OpenIM.

### 4.3 Batas keamanan

Jangan mengklaim end-to-end encryption sebelum implementasi dan audit kriptografi benar-benar selesai. TLS, encryption at rest, dan E2EE adalah hal yang berbeda.

---

## 5. Arsitektur PoC

Arsitektur ini bersifat provisional sampai M0 memverifikasi requirement deployment versi OpenIM yang dipin.

```text
Browser / PWA
   |
   | HTTPS + WSS
   v
+----------------------------+
| Reverse proxy              |
|                            |
| /          -> web          |
| /api/*     -> app-api      |
| /openim/*  -> OpenIM       |
+-------------+--------------+
              |
     Docker private network
              |
     +--------+---------+
     |                  |
     v                  v
+------------+    +----------------+
| app-api    |    | OpenIM Server  |
| Go         |--->| API + WebSocket|
+------+-----+    +-------+--------+
       |                  |
       v                  v
+------------+    +--------------------------+
| App DB     |    | Upstream data services   |
| PostgreSQL |    | Mongo/Redis/Kafka/etc.   |
+------------+    | sesuai versi yang dipin  |
                  +--------------------------+
```

### 5.1 Ownership data

| Data | Pemilik | Aturan |
| --- | --- | --- |
| Username/password | `app-api` | Jangan disimpan ulang di OpenIM kecuali kontrak resmi mengharuskan data minimum |
| App session | `app-api` | Token disimpan dalam bentuk aman/hashed sesuai desain auth |
| Mapping user | `app-api` | Simpan `app_user_id` dan `openim_user_id` |
| Profil utama | `app-api` | Sinkronkan field minimum ke OpenIM |
| Pesan | OpenIM | Jangan diduplikasi ke app DB |
| Conversation/group | OpenIM | Kelola melalui SDK/Server API resmi |
| Attachment | Object storage/OpenIM | App DB hanya menyimpan metadata yang memang dimiliki aplikasi |
| Report/moderation | `app-api` | Referensikan ID OpenIM, jangan salin isi pesan tanpa kebutuhan legal yang jelas |
| Audit event | `app-api` | Jangan log password, access token, atau isi pesan penuh |

### 5.2 Aturan integrasi

- Client tidak boleh memiliki credential rahasia Server API.
- `app-api` menjadi satu-satunya pihak yang memanggil API administratif OpenIM.
- Token OpenIM untuk user diterbitkan melalui backend, bukan dibuat di browser.
- Gunakan kontrak SDK/API resmi; jangan mengandalkan tabel internal.
- Timeout, retry, idempotency, dan error mapping harus eksplisit.
- Jangan mencatat token atau isi pesan pada structured logs.

---

## 6. Struktur repository target

```text
openim/
├── OPENIM_IMPLEMENTATION_PLAYBOOK.md
├── LICENSE_DECISION.md              # dibuat pada M0
├── README.md                        # cara menjalankan stack
├── .env.example
├── infra/
│   ├── compose/
│   │   └── compose.dev.yml
│   └── openim/
│       └── README.md
├── services/
│   └── app-api/
│       ├── cmd/api/
│       ├── internal/
│       └── go.mod
├── web/
│   ├── src/
│   └── package.json
└── scripts/
    ├── dev-up.sh
    ├── dev-down.sh
    └── smoke-test.sh
```

Struktur boleh disederhanakan jika task belum memerlukannya. Jangan membuat seluruh folder sekaligus tanpa isi yang dipakai.

---

## 7. Milestone pengerjaan

## M0 — Due diligence dan version pin

Tujuan: memastikan lisensi, versi, dependency, dan jalur integrasi sebelum coding produk.

- [x] Catat tag/commit OpenIM Server yang dipilih.
- [ ] Catat tag/commit SDK Core dan SDK platform yang kompatibel.
- [ ] Salin license/NOTICE upstream yang relevan ke daftar third-party notices.
- [ ] Verifikasi requirement CPU, RAM, disk, database, Redis, Kafka, object storage, dan port.
- [ ] Verifikasi API login, user token, create user, update profile, dan revoke/block user.
- [ ] Verifikasi dukungan browser dan batasan WebAssembly/SDK web.
- [ ] Kirim permintaan quotation commercial license.
- [x] Buat `LICENSE_DECISION.md` berisi status `pending`, `approved-commercial`, atau `approved-AGPL`.
- [ ] Tulis ADR pemilihan OpenIM atau keputusan membatalkan OpenIM.

Acceptance criteria:

- Versi upstream dan compatibility matrix terdokumentasi.
- Tidak ada asumsi harga commercial license.
- Semua komponen copyleft yang masuk ke client atau backend sudah diidentifikasi.
- Tim mengetahui fitur mana yang open-source, Business, Enterprise, atau add-on.

## M1 — Local infrastructure

Tujuan: menjalankan OpenIM upstream secara repeatable tanpa kode produk.

- [x] Buat `.env.example` tanpa secret nyata.
- [x] Buat Compose berdasarkan deployment resmi versi yang dipin.
- [x] Pin seluruh image dengan tag/digest.
- [x] Expose hanya reverse proxy ke host jika memungkinkan.
- [x] Tambahkan health check untuk OpenIM dan dependency kritis.
- [x] Tambahkan persistent volume bernama jelas.
- [x] Buat `dev-up.sh`, `dev-down.sh`, dan smoke test read-only.
- [x] Dokumentasikan cara reset data sebagai operasi destruktif terpisah.

Acceptance criteria:

- Stack hidup dari clone bersih dengan langkah terdokumentasi.
- Restart container tidak menghilangkan data.
- Health check gagal ketika dependency kritis dihentikan.
- Tidak ada default password production di repository.

## M2 — App API foundation dan auth

Tujuan: aplikasi memiliki identity dan session sendiri serta dapat membuat sesi OpenIM dengan aman.

- [x] Bootstrap Go `app-api`.
- [x] Tambahkan config validation dan graceful shutdown.
- [x] Buat migration user, credential, session, dan OpenIM mapping.
- [x] Implement register/login/logout/refresh.
- [x] Implement provisioning user OpenIM melalui Server API.
- [x] Implement penerbitan token OpenIM melalui backend.
- [x] Pastikan provisioning idempotent.
- [x] Tambahkan rate limit endpoint auth.
- [x] Tambahkan unit dan integration test.

Acceptance criteria:

- Browser tidak pernah menerima Server API secret.
- User yang sama tidak terprovisi dua kali saat retry.
- Token dan password tidak muncul dalam log.
- Logout/revoke menghasilkan perilaku yang terdokumentasi pada app session dan OpenIM session.

## M3 — Personal chat

Tujuan: dua user dapat berkomunikasi secara stabil.

- [x] Login client ke OpenIM menggunakan token dari `app-api`.
- [x] Tampilkan conversation list.
- [x] Cari/pilih user.
- [x] Kirim dan terima pesan teks.
- [ ] Tampilkan sent/delivered/read.
- [ ] Tampilkan typing dan presence.
- [ ] Uji offline recipient.
- [ ] Uji reconnect setelah koneksi diputus.
- [ ] Uji history setelah reload.
- [ ] Uji login user sama pada dua perangkat/browser.

Acceptance criteria:

- Pesan tidak hilang atau tampil ganda pada skenario retry normal.
- Urutan pesan tetap konsisten setelah reconnect.
- Unread count dan read state konsisten di dua perangkat.

## M4 — Group chat dan contact controls

- [ ] Buat grup.
- [ ] Ubah nama dan avatar grup.
- [ ] Tambah/hapus anggota.
- [ ] Terapkan role owner/admin/member.
- [ ] Leave dan dissolve group sesuai izin.
- [ ] Block/unblock user.
- [ ] Pastikan user yang diblokir tidak dapat melewati pembatasan melalui API langsung.

Acceptance criteria:

- Izin grup diverifikasi server-side.
- Perubahan anggota tersinkron di semua client aktif.
- Block behavior terdokumentasi dan diuji.

## M5 — Media dan message actions

- [ ] Upload gambar dan file dengan allowlist MIME serta batas ukuran.
- [ ] Buat thumbnail di luar request kritis jika diperlukan.
- [ ] Tambahkan voice note.
- [ ] Implement reply dan forward.
- [ ] Verifikasi edit/delete/pin terhadap kemampuan upstream.
- [ ] Bersihkan upload yatim dengan job yang aman.
- [ ] Tambahkan malware scanning jika file publik diizinkan.

Acceptance criteria:

- File executable tidak dapat disamarkan hanya melalui ekstensi.
- URL media memiliki kontrol akses yang sesuai.
- Metadata pesan tetap konsisten setelah edit/delete.

## M6 — Moderation, push, dan operasi

- [ ] Report user/message ke `app-api`.
- [ ] Admin dapat melihat dan mengubah status report.
- [ ] Suspend/unsuspend akun tersinkron ke OpenIM.
- [ ] Integrasikan web push, kemudian FCM/APNs saat native client dimulai.
- [ ] Tambahkan structured logging dan correlation ID.
- [ ] Tambahkan metrics, alert dasar, dan dashboard operasional.
- [ ] Buat backup dan jalankan restore drill.
- [ ] Dokumentasikan upgrade dan rollback.

Acceptance criteria:

- Admin action diaudit.
- User suspended tidak dapat memperoleh token/sesi baru.
- Restore menghasilkan data yang dapat dipakai, bukan hanya backup yang berhasil dibuat.

## M7 — Load, reliability, dan security testing

- [ ] Tentukan profil concurrent connection realistis untuk 5.000–8.000 akun.
- [ ] Uji login ramp-up, idle connection, direct messages, group fan-out, reconnect storm, dan history sync.
- [ ] Ukur p50/p95/p99 latency, error rate, CPU, RAM, disk I/O, network, dan queue lag.
- [ ] Jalankan dependency/license scan.
- [ ] Jalankan vulnerability scan terhadap image.
- [ ] Review CORS, TLS, rate limit, secret rotation, dan admin endpoint exposure.
- [ ] Catat kapasitas satu node dan trigger scale-up/scale-out.

Acceptance criteria:

- Target performa ditentukan sebelum test dan hasilnya terdokumentasi.
- Tidak ada data loss pada restart/reconnect test yang disepakati.
- Bottleneck utama diketahui dan memiliki tindakan perbaikan.

## M8 — Voice/video dan production decision

- [ ] Verifikasi apakah modul voice/video termasuk lisensi yang dibeli atau add-on terpisah.
- [ ] Implement call signaling 1-to-1.
- [ ] Siapkan STUN/TURN dan observability media.
- [ ] Uji jaringan NAT, mobile, packet loss, dan reconnect.
- [ ] Hitung bandwidth dan biaya TURN.
- [ ] Buat keputusan `go`, `conditional go`, atau `no-go` untuk OpenIM.

Acceptance criteria:

- Keputusan production menyertakan lisensi, biaya, performa, risiko operasional, dan gap fitur.
- Tidak ada komponen berbayar yang diasumsikan gratis.

---

## 8. Environment dan secret

Nama final harus mengikuti dokumentasi versi yang dipin. Kelompok minimal:

```text
# App
APP_ENV=
APP_HTTP_ADDR=
APP_DATABASE_URL=
APP_SESSION_SECRET=

# OpenIM integration
OPENIM_API_URL=
OPENIM_WS_URL=
OPENIM_ADMIN_SECRET=

# Media/push
OBJECT_STORAGE_ENDPOINT=
OBJECT_STORAGE_BUCKET=
OBJECT_STORAGE_ACCESS_KEY=
OBJECT_STORAGE_SECRET_KEY=
FCM_CREDENTIALS_FILE=
```

Aturan:

- `.env.example` hanya berisi placeholder aman.
- Production secret berasal dari secret manager/deployment environment.
- Fail fast jika secret wajib kosong atau masih memakai nilai contoh.
- Jangan mengirim `OPENIM_ADMIN_SECRET` ke browser.
- Redact credential dari log dan error response.

---

## 9. Profil load test awal

Angka ini adalah starting point, bukan klaim kapasitas OpenIM.

| Skenario | Target awal |
| --- | --- |
| Registered accounts | 8.000 |
| Concurrent connected users | 1.500 |
| Login ramp-up | 100 user/menit |
| Direct message rate | 100 msg/detik burst |
| Group | 100 anggota/grup untuk baseline |
| Reconnect storm | 500 client dalam 60 detik |
| Message delivery p95 | < 1 detik pada kondisi normal |
| API error rate | < 1% di luar error yang disengaja |

Profil harus diperbarui berdasarkan pola pengguna yang sebenarnya. Jangan membeli infrastruktur hanya berdasarkan angka sintetis ini.

---

## 10. Risiko utama

| Risiko | Dampak | Mitigasi |
| --- | --- | --- |
| Harga commercial license tidak publik | Anggaran tidak dapat dipastikan | Selesaikan quotation pada M0 |
| SDK Core AGPL | Konflik dengan produk closed-source | Legal review atau commercial license tertulis |
| Client UI production termasuk Business | PoC demo tidak merepresentasikan biaya production | Bangun UI sendiri atau masukkan Business dalam TCO |
| Banyak dependency operasional | Deployment dan debugging kompleks | Ikuti deployment resmi, pin versi, observability sejak awal |
| Voice/video dapat menjadi add-on | Biaya dan scope bertambah | Pisahkan milestone dan minta harga tertulis |
| Perubahan upstream | Integrasi rusak saat upgrade | Pin versi dan regression test |
| Menulis database internal | Corruption dan upgrade gagal | Gunakan SDK/Server API resmi saja |
| Admin secret bocor | Pengambilalihan sistem | Backend-only, secret rotation, audit log |
| Tidak ada audit E2EE | Klaim keamanan keliru | Jangan klaim E2EE sebelum audit |

---

## 11. Pertanyaan quotation OpenIM

Kirim konteks berikut kepada OpenIM:

```text
We are evaluating OpenIM SDK Core for a proprietary, closed-source,
self-hosted messaging product with approximately 5,000–8,000 registered users.

Please provide:
1. Commercial license price and billing model.
2. Whether SDK Core can be licensed separately from Business client UI.
3. Limits by users, MAU, app, domain, deployment, or server count.
4. Rights to modify and redistribute the SDK in our client applications.
5. Coverage for development, staging, production, and disaster recovery.
6. Included updates, security patches, and support period.
7. Separate pricing for voice/video, meetings, HA, and Kubernetes support.
8. Renewal, termination, and perpetual-use terms.
```

Simpan balasan dan kontrak di penyimpanan perusahaan yang sesuai, bukan di repository publik.

---

## 12. Definition of Done MVP

MVP dianggap selesai hanya jika:

- [ ] License Gate disetujui tertulis.
- [ ] Seluruh acceptance criteria M1–M7 selesai.
- [ ] Backup dan restore drill berhasil.
- [ ] Load test memenuhi target yang disepakati.
- [ ] Security dan dependency review tidak memiliki temuan kritis terbuka.
- [ ] Runbook deploy, rollback, incident, dan secret rotation tersedia.
- [ ] Product tidak membuat klaim keamanan yang belum dibuktikan.
- [ ] Biaya lisensi dan infrastruktur masuk dalam TCO yang disetujui.

---

## 13. Progress log

Tambahkan entri terbaru di bagian paling atas.

```text
YYYY-MM-DD — Mx / task
Perubahan:
Test:
Hasil:
Risiko:
Berikutnya:
```

### 2026-08-20 — M3 web chat vertical slice

- Added `openim/web`, a React/Vite client using the pinned `@openim/wasm-client-sdk` WebAssembly package.
- Browser auth calls `app-api`, obtains the short-lived OpenIM session server-side, then logs the SDK in without exposing the admin secret.
- Added conversation loading, authenticated user search/picker, text send, incoming-message listener, and connection status.
- Standardized all visible web client copy to English.
- Copied the SDK's required `openIM.wasm`, `sql-wasm.wasm`, and `wasm_exec.js` assets into the web app's public directory.

Task:
M3 personal chat vertical slice
File berubah:
`openim/web/**`, `openim/services/app-api/**`, `openim/README.md`
Test yang dijalankan:
`npm install`; `npm run build`; `go test -race ./...`; `go vet ./...`; live register/search privacy/validation request; local Vite dev server; browser DOM load and console error check.
Hasil test:
Build dan Go checks lulus; authenticated user search returned results without email fields and rejected one-character queries; local page rendered the login/register UI with no browser console errors. End-to-end two-browser message/reconnect/offline tests remain open.
Risiko:
Delivery/read state, typing/presence, offline recipient, reconnect, and multi-device history acceptance are not complete.
Berikutnya:
Add delivery/read state and deterministic two-user browser tests before expanding to M4 group chat.

### 2026-08-20 — M2 runtime verification

- Docker Desktop memory was increased to about 3.05 GB and the full Compose stack reached healthy status, including OpenIM Server and OpenIM Chat.
- Runtime auth exposed that UUIDs with hyphens are rejected by OpenIM; app users now map to `app_` plus the UUID without hyphens before provisioning.
- The MinIO healthcheck uses the image-native `mc ready local` command.

Task:
M2 runtime smoke and auth acceptance
File berubah:
`openim/services/app-api/internal/auth/service.go`, `openim/services/app-api/internal/auth/service_test.go`, `openim/infra/compose/compose.dev.yml`
Test yang dijalankan:
Full Compose startup; all service healthchecks; HTTP probes for OpenIM API, OpenIM Chat API, MinIO, and app-api; register, `/v1/me`, OpenIM session, login, refresh rotation, revoked-token rejection, logout/force-logout, and duplicate registration.
Hasil test:
Lulus. Auth scenario returned `201/200/200/200/401/200/401/409` for the expected register, login, refresh, old-token rejection, logout, post-logout rejection, and duplicate-registration statuses. Tokens were not printed.
Risiko:
The local stack still uses development credentials from `.env`; License Gate, production secrets, immutable image digests, and M3 remain open.
Berikutnya:
Mulai M3 web client setelah keputusan lisensi dan deployment secret management disetujui.

### 2026-08-20 — M2 app-api foundation

- Added `openim/services/app-api` Go service with config validation, graceful shutdown, PostgreSQL migrations, opaque app sessions, bcrypt passwords, and auth rate limiting.
- Added separate PostgreSQL `app_chat` service; OpenIM MongoDB remains owned by OpenIM.
- Added backend-only OpenIM adapter for admin token caching, idempotent user provisioning, user-token issuance, and force logout.
- Added `GET /v1/openim/session`; only the user token and public API/WebSocket addresses reach the browser.
- Added unit tests for the OpenIM adapter, including operation IDs and response-body secret redaction.
- Replaced the MinIO healthcheck's unavailable `curl` dependency with the image-native `mc ready local` check.
- Runtime validation started with Docker Desktop: MongoDB, Redis, etcd, Kafka, MinIO, PostgreSQL, and `app-api` become healthy; OpenIM Server aborts its startup because the Docker VM exposes only about 0.51 GB to the container and the pinned image's `check-free-memory` gate requires more.

Task:
M2 app-api foundation
File berubah:
`openim/services/app-api/**`, `openim/infra/compose/compose.dev.yml`, `openim/.env.example`, `openim/scripts/smoke-test.sh`, `openim/README.md`
Perubahan:
Identity aplikasi, session, mapping OpenIM, provisioning, token issuance, and logout revocation.
Test yang dijalankan:
`cd openim/services/app-api && GOCACHE=/tmp/openim-go-cache go test -race ./...`; `GOCACHE=/tmp/openim-go-cache go vet ./...`; `docker compose config --quiet`; `bash -n openim/scripts/*.sh`; `./openim/scripts/dev-up.sh`; `./openim/scripts/smoke-test.sh`; `curl /health/live`; `curl /health/ready`; register failure-path request
Hasil test:
Unit, race, vet, Compose config, and app-api health checks pass. Full smoke test is blocked by `openim-chat` not starting; app-api correctly returns HTTP 502 with a generic `OPENIM_ERROR` when registration cannot provision OpenIM.
Risiko:
M2 integration acceptance remains open until Docker Desktop is allocated enough memory for the full OpenIM process set; OpenIM platform ID and server/API compatibility must then be verified against the pinned runtime.
Berikutnya:
Increase Docker Desktop memory, rerun the full smoke/auth/provisioning scenarios, then start M3 web client.

### 2026-08-20 — M0/M1 foundation

- OpenIM Docker distribution `v3.8` and its server/Chat image pair were pinned in `openim/UPSTREAM_VERSIONS.md`.
- Added pending License Gate record in `openim/LICENSE_DECISION.md` and the initial third-party inventory in `openim/THIRD_PARTY_NOTICES.md`.
- Added a local Compose wrapper with explicit image tags, private dependency network, persistent named volumes, health checks, loopback-only published ports, and no OpenIM source modifications.
- Added `.env.example`, lifecycle scripts, smoke test, and reset-data warning under `openim/`.
- `docker compose config` and shell syntax checks pass. Full pull/start smoke test is pending because no Docker daemon is available in this environment.
- License quotation, legal approval, immutable image digests, SDK compatibility, and all M2+ product work remain open.

Task:
M0/M1 foundation bootstrap
File berubah:
`openim/.env.example`, `openim/infra/compose/compose.dev.yml`, `openim/infra/openim/README.md`, `openim/scripts/*`, `openim/README.md`, `openim/LICENSE_DECISION.md`, `openim/UPSTREAM_VERSIONS.md`, `openim/THIRD_PARTY_NOTICES.md`
Perubahan:
Menyiapkan deployment OpenIM self-hosted lokal yang dapat diulang tanpa memodifikasi upstream.
Test yang dijalankan:
`docker compose config`; `bash -n openim/scripts/*.sh`
Hasil test:
Lulus. Runtime smoke test belum dapat dijalankan tanpa Docker daemon.
Risiko:
Lisensi AGPL/commercial belum diputuskan; digest image belum dicatat; compatibility SDK belum diverifikasi.
Berikutnya:
Selesaikan License Gate dan jalankan stack pada host dengan Docker daemon, lalu mulai M2 `app-api`.

### 2026-08-20 — Planning

- Playbook awal dibuat.
- License Gate ditetapkan sebagai milestone pertama.
- Belum ada versi upstream, quotation, atau arsitektur production yang disetujui.
