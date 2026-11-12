"use strict";
// This view is intentionally passive until a trusted local host application
// installs an adapter. Never derive authority from query strings or messages.
const approve = document.getElementById("approve");
const deny = document.getElementById("deny");
const status = document.getElementById("status");
const fingerprint = document.getElementById("fingerprint");
const session = document.getElementById("session");
const footnote = document.querySelector(".footnote");
let active = null;
let busy = false;
let adapter = null;
function reset(message) {
 active = null;
 approve.disabled = true;
 deny.disabled = true;
 fingerprint.textContent = "Ikke tilgjengelig";
 session.textContent = "Ikke tilgjengelig";
 status.textContent = message;
}
function validPrompt(prompt) {
 return prompt && typeof prompt.ticketId === "string" &&
  /^[A-Za-z0-9_-]{20,128}$/.test(prompt.ticketId) &&
  typeof prompt.sessionId === "string" && prompt.sessionId.length > 0 &&
  prompt.sessionId.length <= 128 &&
  typeof prompt.technicianFingerprint === "string" &&
  /^[A-Za-z0-9:_-]{16,256}$/.test(prompt.technicianFingerprint) &&
  typeof prompt.expiresAt === "string" &&
  Number.isFinite(Date.parse(prompt.expiresAt)) &&
  Date.parse(prompt.expiresAt) > Date.now();
}
function showPrompt(prompt) {
 if (!adapter || busy || !validPrompt(prompt)) {
  reset("Ingen gyldig forespørsel tilgjengelig.");
  return;
 }
 active = Object.freeze({...prompt});
 fingerprint.textContent = active.technicianFingerprint;
 session.textContent = active.sessionId;
 status.textContent = "En tekniker ber om fjernhjelp. Du må velge selv.";
 approve.disabled = false;
 deny.disabled = false;
 footnote.textContent = "Godkjenning gir ikke skjermdeling eller fjernstyring i denne prototypen.";
}
async function decide(approved) {
 if (!adapter || !active || busy) return;
 const prompt = active;
 if (!validPrompt(prompt)) {
  reset("Forespørselen har utløpt.");
  return;
 }
 busy = true;
 approve.disabled = true;
 deny.disabled = true;
 status.textContent = "Sender avgjørelsen …";
 try {
  // The adapter MUST be implemented by the trusted host process and bind
  // the decision to its authenticated socket, never to browser-supplied IDs.
  await adapter.decide(prompt.ticketId, approved);
  reset(approved ? "Godkjenning registrert." : "Forespørselen er avslått.");
 } catch {
  reset("Kunne ikke registrere avgjørelsen. Ingen tilgang er gitt.");
 } finally {
  busy = false;
 }
}
approve.addEventListener("click", () => void decide(true));
deny.addEventListener("click", () => void decide(false));
// The only integration point. A trusted local host application must call
// this with a verified adapter; this script cannot authenticate its caller.
// It MUST NOT be exposed on a publicly served, attacker-controlled page.
Object.defineProperty(window, "PloosAssistHostConsent", {
 value: Object.freeze({
  installTrustedAdapter(value) {
   if (adapter || !value || typeof value.decide !== "function") return false;
   adapter = value;
   status.textContent = "Venter på forespørsel fra den lokale klienten.";
   return true;
  },
  showPrompt
 }),
 writable: false,
 configurable: false
});
reset("Ingen autentisert lokal klient er tilkoblet.");
