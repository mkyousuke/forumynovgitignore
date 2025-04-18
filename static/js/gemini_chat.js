const STORAGE_KEY = 'chatHistory';
let chatHistory = JSON.parse(sessionStorage.getItem(STORAGE_KEY) || '[]');

const container           = document.getElementById('chat-container');
const form                = document.getElementById('chat-form');
const input               = document.getElementById('message');
const newConversationBtn  = document.getElementById('new-conversation-btn');

function saveHistory() {
  sessionStorage.setItem(STORAGE_KEY, JSON.stringify(chatHistory));
}

function loadHistory() {
  chatHistory.forEach(({ text, role }) => renderBubble(text, role, false));
  container.scrollTop = container.scrollHeight;
}

function renderBubble(text, role, store = true) {
  const d = document.createElement('div');
  d.classList.add('chat-message', role);

  // Markdown **gras** et *italique*
  let html = text
    .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
    .replace(/(?<!\*)\*(?!\s)([^*]+?)(?<!\s)\*(?!\*)/g, '<em>$1</em>');

  d.innerHTML = html;
  container.appendChild(d);

  if (store) {
    chatHistory.push({ text, role });
    saveHistory();
  }
  container.scrollTop = container.scrollHeight;
}

document.addEventListener('DOMContentLoaded', loadHistory);

newConversationBtn.addEventListener('click', () => {
  chatHistory = [];
  saveHistory();
  container.innerHTML = '';
  input.focus();
});

form.addEventListener('submit', async function(e) {
  e.preventDefault();
  const msg = input.value.trim();
  if (!msg) return;
  renderBubble(msg, 'user');
  input.value = '';
  input.focus();
  try {
    const res = await fetch('/api/gemini-chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ message: msg }),
    });
    if (!res.ok) {
      const errText = await res.text();
      renderBubble(`Erreur ${res.status} : ${errText}`, 'bot');
      return;
    }
    const { reply } = await res.json();
    renderBubble(reply || 'Pas de réponse', 'bot');
  } catch (err) {
    renderBubble(`Erreur réseau : ${err.message}`, 'bot');
  }
});
