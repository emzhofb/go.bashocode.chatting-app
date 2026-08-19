import { Tinode } from 'tinode-sdk';

const TINODE_HOST = import.meta.env.VITE_TINODE_HOST || 'localhost:6060';
const TINODE_API_KEY = import.meta.env.VITE_TINODE_API_KEY || 'AQEAAAABAAD_rAp4DJh05a1HAwFT3A6K';

export function createTinodeClient() {
  return new Tinode({
    appName: 'BashoCodeChat/0.1.0',
    host: TINODE_HOST,
    apiKey: TINODE_API_KEY,
    transport: 'ws',
    secure: false,
    persist: false,
  });
}

export async function connectTinode(credentials, callbacks = {}) {
  const tinode = createTinodeClient();
  tinode.onConnect = callbacks.onConnect;
  tinode.onDisconnect = callbacks.onDisconnect;
  tinode.onAutoreconnectIteration = callbacks.onAutoreconnectIteration;
  if (credentials.token) {
    tinode.setAuthToken({ token: credentials.token, expires: new Date(credentials.expires) });
  }
  await tinode.connect();
  if (credentials.token) {
    await tinode.loginToken(credentials.token);
  } else {
    await tinode.loginBasic(credentials.username, credentials.password);
  }
  return { client: tinode, authToken: tinode.getAuthToken() };
}
