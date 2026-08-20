# OpenIM-aligned UI specification

Status: **implementation reference**  
Target: `openim/web`  
Primary runtime: responsive Web/PWA with `@openim/wasm-client-sdk`  
Last reviewed: 2026-08-20

---

## 1. Objective

Build the product UI around the interaction model and capabilities demonstrated by the official OpenIM clients, especially the Electron demo for desktop/web and the Flutter demo for mobile.

"Following OpenIM" in this document means:

- use the same core information architecture: conversations, contacts, active chat, profile/group details, and settings;
- expose only actions supported by the pinned OpenIM SDK and our `app-api` contract;
- reflect OpenIM connection, synchronization, unread, message, and conversation states accurately;
- keep the product's own name, colors, typography, icons, copy, spacing, and component implementation.

It does **not** mean copying OpenIM source code, logo, assets, screenshots, or exact visual styling. OpenIM SDK is the messaging capability layer; it does not require a particular UI.

## 2. Official references

Use only official OpenIMSDK sources as the upstream reference:

1. [OpenIMSDK GitHub organization](https://github.com/openimsdk)
2. [OpenIM Electron demo](https://github.com/openimsdk/openim-electron-demo) — primary desktop/web UX and feature reference
3. [Electron demo preview](https://github.com/openimsdk/openim-electron-demo/blob/main/docs/images/preview.png) — high-level layout reference only
4. [OpenIM Flutter demo](https://github.com/openimsdk/openim-flutter-demo) — mobile navigation and responsive behavior reference
5. [OpenIM WASM SDK](https://github.com/openimsdk/openim-sdk-js-wasm) — SDK used by this web client
6. [OpenIM SDK documentation](https://docs.openim.io/) — authoritative API and event behavior
7. [Local implementation playbook](OPENIM_IMPLEMENTATION_PLAYBOOK.md) — product scope, architecture, and license gate

Verified observations:

- the official Electron demo is a reference application for both PC web and desktop runtimes;
- its source separates the main layout, left navigation, top search, chat pages, contact pages, and login pages;
- the official demos expose conversations, contacts, groups, message history, unread state, typing, mute, pin, media messages, and one-to-one calling;
- demo availability does not automatically place every feature inside this product's current milestone;
- the Electron and Flutter demo repositories publish AGPL-3.0 terms with additional restrictions. Do not copy their implementation into proprietary product code without an approved license decision.

## 3. Product UI principles

1. **Conversation first.** After login, the user lands on the conversation workspace, not a dashboard.
2. **Selection over identifiers.** Users select a person or group by name/avatar. Never expose a required `OpenIM user ID` input in the production UI.
3. **State is visible.** Connecting, syncing, online, offline, sending, sent, failed, unread, muted, and typing states must be distinguishable.
4. **Progressive controls.** Keep frequent actions visible; put destructive and administrative actions in menus or detail panels.
5. **One responsive model.** Desktop uses simultaneous panels. Mobile shows one panel at a time with predictable back navigation.
6. **SDK truth.** Do not show an action as successful before the relevant SDK promise or event confirms it.
7. **Original branding.** Preserve the product's existing green visual direction unless a separate branding decision replaces it.

## 4. Information architecture

### 4.1 Desktop and wide tablet

Use a full-height application shell with three persistent areas:

```text
+----------+--------------------------+----------------------------------------+
| App rail | Conversation/contact list| Active content                         |
| 64–72 px | 300–340 px               | flexible, min 480 px                   |
|          |                          |                                        |
| Avatar   | Search                   | Chat header                            |
| Chats    | Filters                  | Message timeline                       |
| Contacts | Rows                     | Composer                               |
| Settings |                          | Optional details drawer                |
+----------+--------------------------+----------------------------------------+
```

The application shell fills `100dvh`. The page itself should not scroll; the list and message timeline scroll independently.

### 4.2 Mobile

At widths below `768px`, render one primary panel at a time:

```text
Conversation list -> Active chat -> Chat details
Contacts          -> Contact detail -> Active chat
```

Requirements:

- replace the left rail with a bottom navigation bar for Chats, Contacts, and Settings;
- opening a chat pushes the chat panel to the foreground;
- the chat header provides a back button that restores the previous list position;
- keep the composer above the virtual keyboard and safe-area inset;
- do not render desktop panels stacked vertically.

## 5. Screen specification

### 5.1 Authentication

Keep the existing `app-api` login and registration flow.

Required UI:

- product identity panel on desktop; compact logo/title on mobile;
- Sign in and Create account modes;
- username/email and password fields;
- display name and email fields only during registration;
- submitting, field error, server error, and OpenIM connection states;
- disable duplicate submission while authentication is in flight.

After `app-api` authentication succeeds, show a short **Connecting to chat…** state while obtaining the OpenIM session and calling SDK login. Do not reveal admin credentials or raw tokens.

### 5.2 Application rail

Desktop order:

1. current-user avatar;
2. Chats with total unread badge;
3. Contacts with pending-request badge when implemented;
4. Settings;
5. connection indicator at the bottom.

Use icons with accessible labels and tooltips. The active destination must be visible without relying only on color.

### 5.3 Conversation sidebar

Header:

- title **Chats**;
- new-chat action;
- search field;
- optional filter chips: All, Unread, Groups.

Each conversation row contains:

- avatar or group avatar;
- display name;
- latest-message preview normalized by message type;
- timestamp;
- unread badge;
- muted indicator;
- pinned indicator;
- draft prefix when a draft exists;
- selected, hover, and keyboard-focus states.

Sort according to the conversation order returned by OpenIM. Do not re-sort only by locally rendered messages.

Row context menu, when supported:

- mark read/unread;
- pin/unpin;
- mute/unmute;
- clear history;
- remove conversation.

Destructive actions require confirmation. Hidden or unsupported SDK actions must not be rendered as working controls.

### 5.4 Empty workspace

When no conversation is selected, show a quiet branded empty state in the main panel:

- product mark or original illustration;
- **Select a conversation to start messaging**;
- **New chat** action.

Do not show the logged-in OpenIM ID as the main content.

### 5.5 Chat header

Left side:

- mobile back action;
- avatar;
- display name or group name;
- presence, member count, connection status, or typing status as appropriate.

Right side, gated by milestone support:

- search in conversation;
- audio call;
- video call;
- details/info menu.

The display name is primary. Internal application IDs and OpenIM IDs belong in a diagnostic or profile detail view, not the normal header.

### 5.6 Message timeline

Required behavior:

- load the latest page on chat selection;
- load older history when scrolling near the top;
- group consecutive messages from the same sender when sensible;
- add date separators and a first-unread divider;
- preserve scroll position when older history is prepended;
- auto-scroll for the user's own message and when already near the bottom;
- show a **new messages** affordance instead of forcing scroll when the user is reading older messages;
- deduplicate messages by client/server message ID when history and real-time events overlap.

Message presentation:

| Message type | Presentation |
| --- | --- |
| Text | Wrapped text bubble with selectable text and preserved line breaks |
| Image | Thumbnail, dimensions reserved before load, open preview on activation |
| File | File name, size, type icon, download state |
| Voice | Play control, duration, progress, playback state |
| Video | Thumbnail, duration, play action |
| System/custom | Centered informational row using product copy |
| Unsupported | Safe fallback: **Unsupported message type**; never render raw payload JSON |

Outgoing message states:

- sending;
- sent;
- delivered/read when confirmed by the SDK contract;
- failed with a retry action.

Message actions, added only when their SDK flow is implemented:

- reply;
- copy;
- forward;
- revoke/delete;
- pin;
- report.

### 5.7 Composer

Layout:

```text
[attachment] [emoji] [ growing text input                ] [voice] [send]
```

Rules:

- hide the recipient-ID field; the active conversation determines the target;
- Enter sends and Shift+Enter inserts a line break on desktop;
- the text input grows to a bounded maximum height;
- disable send for blank content and while no conversation is selected;
- show reply/edit context above the input with a clear cancel action;
- retain an unsent draft per conversation when draft support is implemented;
- attachment, emoji, and voice controls stay hidden until their flows are functional.

### 5.8 Contacts

Provide these sections as their backend/SDK flows become available:

- user search;
- friend requests;
- friends grouped alphabetically;
- groups;
- blocked users inside Settings/Privacy.

Selecting a contact opens a profile detail with **Message** as the primary action. User search must use `app-api`; the browser must not call administrative OpenIM APIs.

### 5.9 Conversation details

Open as a right drawer on desktop and a full page on mobile.

Direct chat:

- avatar and display name;
- presence when supported;
- shared media/files entry point;
- mute, block, clear history, and report actions.

Group chat:

- group avatar, name, and member count;
- member preview and search;
- invite/manage members according to the user's role;
- mute, leave, transfer ownership, or dissolve actions when authorized.

Permissions must be enforced by OpenIM/`app-api`; hiding controls is not authorization.

## 6. Visual direction

The OpenIM demos guide structure and interaction, not brand imitation.

Use the current product direction as the initial token set:

| Token | Initial value |
| --- | --- |
| Primary | `#1d8a68` |
| Primary hover | `#166c52` |
| Text | `#17231f` |
| Muted text | `#65756d` |
| App background | `#f1f5f2` |
| Surface | `#ffffff` |
| Border | `#dbe7e0` |
| Error | `#b64242` |
| Focus ring | primary color at visible 20–30% opacity |

Additional rules:

- use compact, functional desktop spacing similar to a messaging client;
- prefer 8–14 px radii for controls and message bubbles; avoid making every application panel a floating card;
- use one icon family already available in the project; do not copy OpenIM icons;
- maintain WCAG AA text contrast;
- support keyboard navigation and visible focus;
- respect `prefers-reduced-motion`;
- truncate list metadata, never the user's message content in the active timeline.

## 7. SDK-to-UI mapping

The exact signature must be verified against the version pinned in `UPSTREAM_VERSIONS.md` before implementation.

| UI responsibility | Current OpenIM integration point |
| --- | --- |
| Connection indicator | SDK connecting, connect-success, and connect-failed events |
| Authenticate chat client | SDK `login` using the user session returned by `app-api` |
| Conversation sidebar | `getAllConversationList` plus conversation-change events |
| Open direct chat | `getOneConversation` with the selected application/OpenIM user mapping |
| Message timeline | `getAdvancedHistoryMessageList` with paginated history |
| Real-time messages | `OnRecvNewMessages`, merged and deduplicated into active chat and sidebar state |
| Text send | `createTextMessage`, then `sendMessage` using the active conversation target |
| Unread badges | unread values from conversation/account state and unread-change events |
| Read state | SDK read-report API after the active timeline is actually viewed |
| Typing status | OpenIM typing/custom-message capability for direct chats |
| Pin/mute/draft | matching conversation APIs and conversation-change events |
| User directory | authenticated `app-api` search, never OpenIM admin APIs from the browser |
| Group administration | trusted `app-api`/OpenIM Platform API plus client SDK state refresh |

Keep SDK calls behind small product-facing functions. React components should consume application state such as `connectionState`, `activeConversation`, `messages`, and `sendState`, rather than parse raw SDK payloads throughout the view tree.

## 8. Required UI states

Every main surface must define these states before it is considered complete:

| Surface | States |
| --- | --- |
| App connection | initializing, connecting, syncing, ready, offline, failed, kicked/token expired |
| Conversation list | loading, loaded, empty, search-empty, stale/offline, error |
| Active chat | none selected, loading history, ready, empty, loading older, error |
| Message | sending, sent, read when available, failed, revoked, unsupported |
| User search | idle, typing, loading, results, empty, error |
| Composer | disabled, ready, uploading/recording when added, send failed |

Use inline recovery near the failed surface. Reserve global toasts for cross-surface events; do not rely on a single floating error message for all failures.

## 9. Implementation sequence

### Phase UI-1 — restructure the current vertical slice

No new product capability is required in this phase.

- [ ] Replace the page-width card layout with the full-height app shell.
- [ ] Add the desktop app rail and responsive mobile bottom navigation.
- [ ] Move user search into a New chat flow.
- [ ] Remove the recipient OpenIM ID input from the composer.
- [ ] Upgrade conversation rows with avatar placeholders, preview, time, selection, and unread-ready slots.
- [ ] Upgrade chat header, empty state, message bubbles, and composer layout.
- [ ] Preserve current login/register, conversation loading, direct-chat history, real-time receive, and text-send behavior.
- [ ] Split the monolithic `App.jsx` only where needed to keep state ownership clear; do not introduce a new framework or dependency.

### Phase UI-2 — conversation correctness

- [ ] Paginate older history without scroll jumps.
- [ ] Deduplicate real-time and historical messages.
- [ ] Add unread badges and read reporting.
- [ ] Add reconnect/offline/token-expired UX.
- [ ] Add send progress, failure, and retry.
- [ ] Add per-conversation drafts and typing indicator if supported by the pinned SDK.

### Phase UI-3 — supported rich features

- [ ] Contacts and friend requests.
- [ ] Groups and role-aware group details.
- [ ] Image, file, voice, video, and emoji messages.
- [ ] Reply, forward, revoke, pin, and search.
- [ ] Block/report and moderation flows.
- [ ] One-to-one call UI only after the call service and HTTPS requirements are verified.

## 10. UI-1 acceptance criteria

- [ ] Desktop at `1440×900` shows rail, conversation sidebar, and active chat without document-level scrolling.
- [ ] Tablet at `1024×768` remains usable without clipped composer/header actions.
- [ ] Mobile at `390×844` shows one panel at a time with working back navigation and no horizontal overflow.
- [ ] A user can register or sign in, select/search another user, open history, send text, and receive text without entering an OpenIM ID manually.
- [ ] Switching conversations never leaks messages from the previous conversation.
- [ ] Empty, loading, offline/connecting, and error states are visible and recoverable.
- [ ] Long names, long messages, multiline drafts, and empty lists do not break layout.
- [ ] All interactive controls are reachable by keyboard and have an accessible name.
- [ ] Existing tests/build pass; add focused tests for any state extraction introduced by the UI split.
- [ ] No OpenIM logo, screenshot, icon, CSS, component code, or other demo asset is copied into the product.

## 11. Out of scope for this document

- changing OpenIM Server or SDK Core;
- adding dependencies solely to reproduce the OpenIM demo;
- implementing features that are not yet supported by the pinned SDK/backend contract;
- changing product authentication ownership;
- copying or rebranding an upstream demo;
- resolving the commercial/AGPL license decision.

If an upstream demo and the pinned SDK disagree, follow the pinned SDK documentation and record the discrepancy before changing the UI contract.
