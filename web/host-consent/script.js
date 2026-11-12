"use strict";
// UI-only preview. Never enable approval controls based on URL parameters,
// localStorage, postMessage, or unauthenticated WebSocket payloads.
// A future authenticated host transport must supply trusted display data
// and an explicit user-action callback, with fail-closed behavior.
const approve = document.getElementById("approve");
const deny = document.getElementById("deny");
approve.disabled = true;
deny.disabled = true;
