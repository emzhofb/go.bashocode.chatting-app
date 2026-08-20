# OpenIM web client

M3 vertical slice for local development. It authenticates through `app-api`, requests an OpenIM user session, logs the browser into the pinned WebAssembly SDK, loads conversations, and sends/receives text messages.

## Run

```sh
npm install
npm run dev
```

Open `http://127.0.0.1:3000`. The OpenIM WASM assets are copied from the pinned SDK package into `public/` as part of the checked-in local PoC asset set.

Optional Vite variables:

- `VITE_APP_API_URL` (default `http://127.0.0.1:8080`)
- `VITE_OPENIM_API_URL` (default `http://127.0.0.1:10002`)
- `VITE_OPENIM_WS_URL` (default `ws://127.0.0.1:10001`)
- `VITE_OPENIM_PLATFORM_ID` (default `5`)

This is a development client. It does not contain the OpenIM admin secret. Recipient selection supports an OpenIM user ID directly or through authenticated user search; selecting a user loads the direct conversation history. Richer conversation controls remain later M3 work.
