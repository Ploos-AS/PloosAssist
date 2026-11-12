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
let stopped = false;
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
 if (stopped || !adapter || busy || !validPrompt(prompt)) {
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
 if (stopped || !adapter || !active || busy) return;
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

const bootstrapToken = document.querySelector('meta[name="ploos-token"]')?.content;
if (bootstrapToken && /^[A-Za-z0-9_-]{40,64}$/.test(bootstrapToken)) {
 const adapter = {
  async decide(ticketId, approved) {
   const response = await fetch("/api/consent/decision", {
    method: "POST",
    credentials: "omit",
    cache: "no-store",
    headers: {"Content-Type":"application/json","X-PloosAssist-Token":bootstrapToken},
    body: JSON.stringify({ticketId, approve: approved})
   });
   if (!response.ok) throw new Error("decision failed");
   const result = await response.json();
   if (result.type !== "decision_recorded") throw new Error("invalid response");
  }
 };
 if (window.PloosAssistHostConsent.installTrustedAdapter(adapter)) {
  let lastTicket = "";
  async function poll() {
   if (document.hidden || stopped) return;
   try {
    const response = await fetch("/api/consent", {
     credentials:"omit",cache:"no-store",
     headers: {"X-PloosAssist-Token":bootstrapToken}
    });
    if (!response.ok) throw new Error("request failed");
    const result = await response.json();
    if (result.type === "approval_none") {
     if (lastTicket) {lastTicket="";reset("Ingen ventende forespørsler.");}
     return;
    }
    if (result.type !== "approval_pending") throw new Error("invalid response");
    if (result.ticketId !== lastTicket) {
     lastTicket = result.ticketId;
     window.PloosAssistHostConsent.showPrompt(result);
    }
   } catch {
    lastTicket="";
    reset("Mistet forbindelsen til lokal klient. Ingen tilgang er gitt.");
   }
  }
  void poll();
  setInterval(() => void poll(), 2000);
 }
}

const stopButton = document.getElementById("stop");
stopButton.addEventListener("click", async () => {
 if (!bootstrapToken || stopButton.disabled) return;
 stopped = true;
 shareButton.disabled = true;
 shareStatus.textContent = "Skjermdeling stoppet.";
 stopButton.disabled = true;
 approve.disabled = true;
 deny.disabled = true;
 active = null;
 status.textContent = "Avslutter supportøkten …";
 try {
  const response = await fetch("/api/consent/stop", {
   method:"POST",credentials:"omit",cache:"no-store",
   headers:{"X-PloosAssist-Token":bootstrapToken}
  });
  if (!response.ok) throw new Error("stop failed");
  const result = await response.json();
  if (result.type !== "session_stopped") throw new Error("invalid stop response");
  reset("Serveren har bekreftet tilbakekalling. Supportøkten er avsluttet.");
 } catch {
  reset("Lokal tilgang er stoppet, men serverens tilbakekalling er ikke bekreftet. Lukk klienten.");
 }
 // Never enable the controls again in this browser session.
 stopButton.disabled = true;
});

const shareButton = document.getElementById("share-screen");
const shareStatus = document.getElementById("share-status");
shareButton.addEventListener("click", async () => {
 if (stopped || shareButton.disabled || !bootstrapToken) return;
 shareButton.disabled = true;
 shareStatus.textContent = "Ber Linux om skjermvalg. Dette starter ikke videooverføring.";
 try {
  const response = await fetch("/api/screen-share/start",{
   method:"POST",credentials:"omit",cache:"no-store",
   headers:{"X-PloosAssist-Token":bootstrapToken}
  });
  if (!response.ok) throw new Error("portal request failed");
  const result = await response.json();
  if (result.type !== "screen_share_requested") throw new Error("unexpected response");
  shareStatus.textContent = "Forespørselen er sendt til Linux. Ingen skjermbilder overføres.";
 } catch {
  shareStatus.textContent = "Kunne ikke be om skjermdeling. Ingen skjermbilder overføres.";
 }
});

const portalStatusText = Object.freeze({
 idle:"Skjermdeling er ikke startet.",
 waiting:"Venter på skjermvalg i Linux.",
 approved:"Linux har godkjent portaløkten. Ingen skjermbilder overføres.",
 capturing:"Skjermbilder mottas kun lokalt. Ingen bilder overføres til teknikeren.",
 finished:"Lokal skjermfangst er avsluttet.",
 denied:"Skjermdeling ble avslått eller avvist av Linux.",
 error:"Feil ved oppstart av skjermdeling. Ingen skjermbilder overføres.",
 stopped:"Skjermdeling stoppet.",
 unavailable:"Skjermdeling er ikke tilgjengelig."
});
async function pollScreenShareStatus(){
 if (stopped || !bootstrapToken || document.hidden) return;
 try {
  const response = await fetch("/api/screen-share/status",{
   credentials:"omit",cache:"no-store",
   headers:{"X-PloosAssist-Token":bootstrapToken}
  });
  if (!response.ok) throw new Error("status unavailable");
  const result = await response.json();
  if (result.type!=="screen_share_status" ||
      !Object.prototype.hasOwnProperty.call(portalStatusText,result.state))
   throw new Error("invalid status");
  if (!stopped) shareStatus.textContent=portalStatusText[result.state];
 } catch {
  if (!stopped) shareStatus.textContent="Kunne ikke hente status for skjermdeling.";
 }
}
void pollScreenShareStatus();
setInterval(() => void pollScreenShareStatus(),2000);
