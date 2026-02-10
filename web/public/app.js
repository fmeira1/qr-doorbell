let sessionId = null;
let ws = null;
let role = "visitor"; // "visitor" | "owner"
let messageCount = 0;
let unitId = null;
let unitNumber = null;
let doorbellId = null;

const messagesDiv = document.getElementById("messages");
const startBtn = document.getElementById("start");
const closeBtn = document.getElementById("close");
const title = document.getElementById("title");
const roleBadge = document.getElementById("role-badge");
const roleLabel = document.getElementById("role-label");
const statusPill = document.getElementById("status-pill");
const statusMain = document.getElementById("status-main");
const sessionIndicator = document.getElementById("session-indicator");
const messageCountLabel = document.getElementById("message-count");
const hintText = document.getElementById("hint-text");
const sendBtn = document.getElementById("send");
const textInput = document.getElementById("text");
const unitLabel = document.getElementById("unit-label");
const wsStatus = document.getElementById("ws-status");
const backToUnitsBtn = document.getElementById("back-to-units");

function updateMessageCount() {
  if (messageCount === 1) {
    messageCountLabel.textContent = "1 mensaje";
  } else {
    messageCountLabel.textContent = `${messageCount} mensajes`;
  }
}

function addMessageBubble({ sender, content, meta }) {
  const wrap = document.createElement("div");
  wrap.className = "msg" + (sender === role ? " me" : "");

  const icon = document.createElement("span");
  icon.className = "msg-icon";
  icon.textContent = sender === "owner" ? "👤" : "🔔";

  const textWrap = document.createElement("div");
  const textEl = document.createElement("div");
  textEl.className = "msg-text";
  textEl.textContent = content;
  textWrap.appendChild(textEl);

  if (meta) {
    const metaEl = document.createElement("div");
    metaEl.className = "msg-meta";
    metaEl.textContent = meta;
    textWrap.appendChild(metaEl);
  }

  wrap.appendChild(icon);
  wrap.appendChild(textWrap);
  messagesDiv.appendChild(wrap);
  messagesDiv.scrollTop = messagesDiv.scrollHeight;

  messageCount += 1;
  updateMessageCount();
}

function getSessionIdFromURL() {
  const params = new URLSearchParams(window.location.search);
  return params.get("session_id");
}

function getParamsFromURL() {
  const params = new URLSearchParams(window.location.search);
  return {
    sessionId: params.get("session_id"),
    unitId: params.get("unit_id"),
    unitNumber: params.get("unit_number"),
    doorbellId: params.get("doorbell_id"),
  };
}

// Si no hay unit_id, redirigir a selección
function checkUnitSelection() {
  const params = getParamsFromURL();
  if (!params.sessionId && !params.unitId) {
    window.location.href = `select-unit.html?doorbell_id=${params.doorbellId}`;
    return false;
  }
  return true;
}

async function loadHistory(id) {
  try {
    const base = window.location.origin;
    const res = await fetch(`${base}/messages?session_id=${id}`);
    if (!res.ok) {
      console.error("Error cargando historial");
      return;
    }

    const messages = await res.json();
    messages.forEach(m => {
      addMessageBubble({
        sender: m.sender,
        content: m.content,
        meta: new Date(m.created_at).toLocaleTimeString()
      });
    });
  } catch (e) {
    console.error("Error historial:", e);
  }
}

function connectWS(id) {
  const protocol = window.location.protocol === "https:" ? "wss" : "ws";
  const host = window.location.host;

  ws = new WebSocket(`${protocol}://${host}/ws/${id}`);

  ws.onopen = () => {
    statusMain.textContent = "Conectado";
    if (wsStatus) {
      wsStatus.querySelector(".tag-dot").style.background = "#4ade80";
    }
  };

  ws.onmessage = e => {
    try {
      const msg = JSON.parse(e.data);
      addMessageBubble({
        sender: msg.sender,
        content: msg.content,
        meta: new Date(msg.created_at).toLocaleTimeString()
      });
    } catch {
      addMessageBubble({
        sender: "system",
        content: e.data
      });
    }
  };

  ws.onerror = () => {
    statusMain.textContent = "Error de conexión";
    statusPill.style.borderColor = "rgba(248,113,113,0.8)";
    if (wsStatus) {
      wsStatus.querySelector(".tag-dot").style.background = "#ef4444";
    }
  };

  ws.onclose = () => {
    statusMain.textContent = "Desconectado";
    if (wsStatus) {
      wsStatus.querySelector(".tag-dot").style.background = "#9ca3af";
    }
  };
}

function setOwnerUI() {
  role = "owner";
  title.textContent = "Mensajes";
  roleLabel.textContent = "Propietario";
  roleBadge.style.display = "inline-flex";
  backToUnitsBtn.style.display = "none";
  startBtn.style.display = "none";
  closeBtn.style.display = "inline-flex";
  hintText.textContent = "Respondé los mensajes que recibas.";
  statusPill.style.display = "inline-flex";
  if (wsStatus) {
    wsStatus.style.display = "inline-flex";
  }
  sessionIndicator.style.display = sessionId ? "inline-flex" : "none";
  if (messageCountLabel && messageCountLabel.parentElement) {
    messageCountLabel.parentElement.style.display = "inline-flex";
  }
  if (unitNumber) {
    unitLabel.textContent = `Departamento: ${unitNumber}`;
  }
}

function setVisitorUI() {
  role = "visitor";
  title.textContent = "Timbre";
  roleLabel.textContent = "Visitante";
  roleBadge.style.display = "none";
  backToUnitsBtn.style.display = "inline-flex";
  startBtn.style.display = "inline-flex";
  closeBtn.style.display = "none";
  hintText.textContent = "Tocá el timbre para comenzar a chatear.";
  statusPill.style.display = "none";
  if (wsStatus) {
    wsStatus.style.display = "none";
  }
  sessionIndicator.style.display = "none";
  if (messageCountLabel && messageCountLabel.parentElement) {
    messageCountLabel.parentElement.style.display = "none";
  }
  if (unitNumber) {
    unitLabel.textContent = `Departamento: ${unitNumber}`;
  }
}

function setSession(id) {
  sessionId = id;
  if (id) {
    if (role === "owner") {
      sessionIndicator.style.display = "inline-flex";
    }
  }
}

async function handleSend() {
  const text = textInput.value.trim();
  if (!sessionId || !text) return;

  try {
    await fetch("/message", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        session_id: sessionId,
        sender: role,
        content: text
      })
    });
    textInput.value = "";
  } catch (e) {
    console.error("Error enviando mensaje:", e);
  }
}

// --- INICIO ---
if (!checkUnitSelection()) {
  // Redirigido a select-unit.html, salir
  // (checkUnitSelection ya hace window.location.href)
}

const params = getParamsFromURL();
sessionId = params.sessionId;
unitId = params.unitId;
unitNumber = params.unitNumber ? decodeURIComponent(params.unitNumber) : null;
doorbellId = params.doorbellId;

if (sessionId) {
  setSession(sessionId);
  setOwnerUI();

  (async () => {
    await loadHistory(sessionId);
    connectWS(sessionId);
  })();
} else {
  setVisitorUI();
  
  // Si no hay unit_id, redirigir
  if (!unitId) {
    window.location.href = `select-unit.html?doorbell_id=${doorbellId}`;
  }
}

// --- VISITANTE: crear sesión (con ubicación para geovalla si está activa) ---
function getVisitorPosition() {
  return new Promise((resolve) => {
    if (!navigator.geolocation) {
      resolve(null);
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (pos) => resolve({ lat: pos.coords.latitude, lon: pos.coords.longitude }),
      () => resolve(null),
      { timeout: 10000, maximumAge: 60000 }
    );
  });
}

startBtn.onclick = async () => {
  if (!unitId) {
    window.location.href = `select-unit.html?doorbell_id=${doorbellId}`;
    return;
  }

  hintText.classList.remove("error");
  hintText.textContent = "Obteniendo ubicación…";

  const position = await getVisitorPosition();
  const body = { doorbell_id: doorbellId, unit_id: unitId };
  if (position) {
    body.lat = position.lat;
    body.lon = position.lon;
  }

  hintText.textContent = "Seleccioná el departamento y tocá el timbre para comenzar a chatear.";

  try {
    const res = await fetch("/session/start", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body)
    });

    if (!res.ok) {
      const msg = await res.text();
      hintText.textContent = msg || "Error iniciando sesión";
      hintText.classList.add("error");
      return;
    }

    const data = await res.json();
    setSession(data.id);
    messageCount = 0;
    updateMessageCount();
    messagesDiv.innerHTML = "";

    addMessageBubble({
      sender: "system",
      content: "Timbre activado",
      meta: null
    });
    connectWS(data.id);
  } catch (e) {
    console.error("Error al crear sesión:", e);
    hintText.textContent = "Error de conexión. Intentá de nuevo.";
    hintText.classList.add("error");
  }
};

if (backToUnitsBtn) {
  backToUnitsBtn.onclick = () => {
    window.location.href = `select-unit.html?doorbell_id=${doorbellId}`;
  };
}

// --- ENVIAR MENSAJE ---
sendBtn.onclick = handleSend;
textInput.addEventListener("keydown", (e) => {
  if (e.key === "Enter" && !e.shiftKey) {
    e.preventDefault();
    handleSend();
  }
});

// --- CERRAR SESIÓN (WS) ---
closeBtn.onclick = () => {
  if (ws) ws.close();
};

// --- BARRA DE USUARIO AUTENTICADO ---
(async () => {
  try {
    const res = await fetch('/auth/me', { credentials: 'include' });
    if (!res.ok) return;
    const user = await res.json();
    const bar = document.getElementById('user-bar');
    const emailEl = document.getElementById('user-bar-email');
    if (bar && emailEl) {
      emailEl.textContent = user.email;
      bar.style.display = 'flex';
      // Si es admin, agregar enlace al panel
      if (user.is_admin) {
        const adminLink = document.createElement('a');
        adminLink.href = 'admin.html';
        adminLink.textContent = 'Panel admin';
        adminLink.style.cssText = 'font-size:0.82rem; color:var(--accent); text-decoration:none;';
        bar.querySelector('div').prepend(adminLink);
      }
    }
    const logoutLink = document.getElementById('user-logout');
    if (logoutLink) {
      logoutLink.addEventListener('click', async (e) => {
        e.preventDefault();
        await fetch('/auth/logout', { method: 'POST', credentials: 'include' });
        document.getElementById('user-bar').style.display = 'none';
      });
    }
  } catch (e) { /* No logueado, no mostrar barra */ }
})();
