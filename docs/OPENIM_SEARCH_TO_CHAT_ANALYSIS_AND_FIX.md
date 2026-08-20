# Analisis dan Rencana Fix: Hasil Search Tidak Membuka Chat

## Ringkasan

Masalah ini **bukan karena tombol hasil pencarian gagal menerima klik**. Klik sudah diproses oleh React, tetapi handler saat ini hanya:

1. mengisi `recipient` dengan OpenIM user ID;
2. mengganti isi input pencarian;
3. menutup daftar hasil pencarian.

Aplikasi belum memiliki state untuk percakapan yang sedang aktif dan belum memanggil SDK OpenIM untuk mengambil riwayat pesan. Akibatnya, setelah hasil search diklik, area chat tetap menampilkan keadaan sebelumnya sehingga pengguna merasa tidak masuk ke kolom chat.

Status: **fix sudah diterapkan** pada `openim/web/src/App.jsx` dan `openim/web/src/styles.css`. `npm run build` berhasil dijalankan setelah perubahan.

## Hasil Reproduksi

Pengujian dilakukan pada `http://127.0.0.1:3000/` dengan user login `codebasho`.

1. Cari `ba` pada input **Search username or name**.
2. Hasil `basho code` / `@bashocode` muncul.
3. Klik hasil tersebut.
4. Hasil pencarian hilang.
5. Input **Recipient OpenIM user ID** terisi `app_6bdb430121574f36a15018999f199896`.
6. Header dan isi area pesan tidak berubah; tetap menampilkan user yang sedang login dan `Send the first message to get started.`
7. Console browser tidak mencatat error atau warning.

Kesimpulan reproduksi: handler klik berhasil, tetapi tidak ada proses untuk membuka dan merender percakapan yang dipilih.

## Akar Masalah

### 1. Handler hasil search hanya mengubah `recipient`

Lokasi: `openim/web/src/App.jsx:136`

Handler saat ini:

```jsx
onClick={() => {
  setRecipient(user.openim_user_id)
  setUserQuery(user.username)
  setUserResults([])
}}
```

Tidak ada pemanggilan fungsi seperti `openConversation`, tidak ada perubahan identitas chat aktif, dan tidak ada pengambilan history.

### 2. Tidak ada state `activeChat` atau `selectedConversation`

Lokasi state: `openim/web/src/App.jsx:11-22`

State yang ada hanya menyimpan `recipient`, `messages`, dan `conversations`. `recipient` berupa string user ID tidak cukup untuk merender:

- nama user yang sedang diajak chat;
- conversation ID;
- status loading/error percakapan;
- daftar pesan yang khusus milik chat aktif.

### 3. Riwayat pesan tidak pernah dimuat

`sdk.getAllConversationList()` sudah dipanggil pada `refreshConversations`, tetapi tidak ada pemanggilan:

```js
sdk.getOneConversation(...)
sdk.getAdvancedHistoryMessageList(...)
```

State `messages` hanya bertambah ketika ada event pesan baru atau setelah user mengirim pesan. Karena itu, memilih user atau conversation tidak dapat menampilkan history.

### 4. Klik conversation list memiliki masalah yang sama

Lokasi: `openim/web/src/App.jsx:135`

Klik conversation yang sudah ada juga hanya memanggil `setRecipient(...)`. Jadi perbaikannya sebaiknya memakai satu fungsi pemilih chat yang sama untuk hasil search dan conversation list.

### 5. Pesan masuk tidak difilter berdasarkan chat aktif

Lokasi: `openim/web/src/App.jsx:30`

Semua pesan dari `OnRecvNewMessages` langsung ditambahkan ke satu array `messages`. Setelah pemilihan chat diperbaiki, perilaku ini dapat mencampur pesan dari user lain ke chat yang sedang dibuka.

## Fix yang Direkomendasikan

Lakukan perubahan kecil dan terpusat di `openim/web/src/App.jsx`, dengan penyesuaian visual minimal di `openim/web/src/styles.css`. Tidak perlu dependency baru dan tidak perlu mengubah API backend.

### 1. Tambahkan state chat aktif

```jsx
const [activeChat, setActiveChat] = useState(null)
const [isLoadingMessages, setIsLoadingMessages] = useState(false)
```

Untuk direct message, bentuk minimalnya:

```js
{
  userID,
  username,
  displayName,
  conversationID,
}
```

`recipient` dapat tetap dipertahankan untuk kompatibilitas dengan composer yang ada, tetapi nilainya harus disinkronkan dari `activeChat.userID`.

### 2. Buat satu fungsi untuk membuka direct chat

Gunakan fungsi yang sama dari hasil search dan conversation list.

```jsx
async function openDirectChat(user) {
  const userID = user.openim_user_id || user.userID
  if (!userID) return

  setError('')
  setIsLoadingMessages(true)
  setRecipient(userID)
  setActiveChat({
    userID,
    username: user.username || '',
    displayName: user.display_name || user.showName || user.username || userID,
    conversationID: user.conversationID || '',
  })
  setUserResults([])

  try {
    const conversationResult = user.conversationID
      ? { data: user }
      : await sdk.getOneConversation({ sourceID: userID, sessionType: 1 })

    const conversation = conversationResult.data
    const historyResult = await sdk.getAdvancedHistoryMessageList({
      conversationID: conversation.conversationID,
      startClientMsgID: '',
      count: 50,
      viewType: 0,
    })

    setActiveChat((current) => current && {
      ...current,
      conversationID: conversation.conversationID,
      displayName: current.displayName || conversation.showName,
    })
    setMessages((historyResult.data?.messageList || []).map(toMessage))
  } catch (cause) {
    setMessages([])
    setError(`Could not open conversation: ${cause.message || cause}`)
  } finally {
    setIsLoadingMessages(false)
  }
}
```

Catatan SDK yang sudah diverifikasi dari versi terpasang `@openim/wasm-client-sdk@3.8.3-patch.15.1`:

- direct/single session memakai `sessionType: 1`;
- history memakai `viewType: 0`;
- hasil history berada di `result.data.messageList`.

Jika urutan hasil history dari SDK tampil terbalik, urutkan berdasarkan `sendTime` secara ascending sebelum `setMessages`.

### 3. Sambungkan kedua sumber pilihan ke fungsi yang sama

Hasil search:

```jsx
<button key={user.id} type="button" onClick={() => openDirectChat(user)}>
```

Conversation list:

```jsx
<button
  className="conversation"
  key={conversation.conversationID}
  type="button"
  onClick={() => openDirectChat(conversation)}
>
```

Memberi `type="button"` membuat maksud tombol eksplisit dan mencegah submit form bila struktur JSX berubah di kemudian hari.

### 4. Tampilkan identitas chat aktif pada header

Header chat sekarang selalu menampilkan user yang sedang login. Ubah agar ada feedback yang jelas setelah hasil search diklik.

```jsx
<p className="eyebrow">{activeChat ? 'CHAT WITH' : 'LOGGED IN AS'}</p>
<h2>{activeChat?.displayName || currentUser.username}</h2>
<span className="pill">{activeChat?.username ? `@${activeChat.username}` : activeChat?.userID || currentUser.openim_user_id}</span>
```

Area pesan sebaiknya membedakan tiga kondisi:

- belum memilih user: `Select a user to start chatting.`;
- sedang memuat: `Loading messages…`;
- chat aktif tanpa history: `Send the first message to get started.`

### 5. Batasi pesan baru ke chat aktif

Event `OnRecvNewMessages` harus hanya menambahkan pesan yang berasal dari atau ditujukan ke `activeChat.userID`. Karena callback SDK didaftarkan sekali, simpan user ID aktif pada `useRef`, atau daftar ulang listener saat active chat berubah dengan cleanup yang benar.

Contoh kondisi filter direct message:

```js
const belongsToActiveChat = (message, activeUserID, currentUserID) =>
  (message.sendID === activeUserID && message.recvID === currentUserID) ||
  (message.sendID === currentUserID && message.recvID === activeUserID)
```

Pesan untuk chat lain tidak perlu dimasukkan ke layar aktif; cukup refresh conversation list agar preview dan unread state dapat diperbarui.

### 6. Normalisasi teks pesan OpenIM

Untuk text message, prioritaskan `textElem.content`. Field `content` dari OpenIM dapat berupa payload JSON string, sehingga menampilkannya langsung berisiko menghasilkan JSON mentah.

```js
function toMessage(message) {
  return {
    ...message,
    text: message.textElem?.content || message.text || message.msg || parseMessageContent(message.content),
  }
}
```

Parser harus aman: bila `JSON.parse` gagal, kembalikan string `content` asli.

## Perubahan File

| File | Perubahan |
| --- | --- |
| `openim/web/src/App.jsx` | Tambah chat aktif, fungsi pembuka chat, load history, filter pesan masuk, dan feedback UI. |
| `openim/web/src/styles.css` | Opsional: state aktif pada conversation dan state loading/empty. |

Backend search sudah mengembalikan `openim_user_id`, `username`, dan `display_name`, sehingga `openim/services/app-api` tidak perlu diubah.

## Acceptance Criteria

1. Search minimal dua karakter menampilkan user lain.
2. Klik hasil search mengubah header chat ke user yang dipilih.
3. Recipient ID terset otomatis dan tidak perlu diedit manual.
4. History direct message untuk user tersebut dimuat.
5. Bila belum ada history, empty state khusus chat aktif ditampilkan.
6. Klik conversation di sidebar membuka history conversation yang benar.
7. Berpindah antar-chat tidak mencampur pesan.
8. Pesan baru hanya muncul pada chat aktif yang sesuai.
9. Mengirim pesan tetap memakai `activeChat.userID` dan conversation list ter-refresh.
10. Tidak ada error/warning di console browser pada alur search lalu klik.

## Verifikasi Setelah Implementasi

Jalankan:

```sh
cd openim/web
npm run build
```

Lalu uji manual pada `http://127.0.0.1:3000/`:

1. login sebagai user A;
2. search user B;
3. klik user B dan pastikan header serta history berubah;
4. kirim satu pesan;
5. pindah ke conversation lain lalu kembali ke user B;
6. pastikan pesan user B tetap benar dan tidak tercampur;
7. kirim pesan dari user B ke user A pada sesi/browser terpisah dan pastikan event real-time masuk ke chat yang sesuai.

## Risiko

- Pemanggilan history dapat selesai setelah user cepat berpindah ke chat lain. Cegah stale response dengan request sequence/ref atau cek bahwa user ID yang selesai dimuat masih sama dengan chat aktif sebelum memanggil `setMessages`.
- Group chat memiliki `sessionType: 3` dan `groupID`; fungsi di atas sengaja fokus pada direct message dari hasil user search. Bila conversation list juga memuat group, buat dispatcher kecil berdasarkan `conversation.conversationType` tanpa memperluas perubahan pada alur search.
- Conversation preview `latestMsg` pada SDK bertipe string, bukan object. Perbaikan parsing preview adalah isu terpisah dan tidak perlu memblokir fix pemilihan chat.

## Keputusan

Fix terbaik adalah memperkenalkan satu state chat aktif dan satu fungsi `openDirectChat` yang dipakai oleh hasil search serta conversation list. Mengandalkan `recipient` saja tidak cukup karena recipient ID hanya dibutuhkan untuk pengiriman pesan, bukan untuk mengelola lifecycle sebuah layar percakapan.
