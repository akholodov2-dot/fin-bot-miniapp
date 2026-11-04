const tg = window.Telegram.WebApp;
tg.ready();
tg.expand();

async function apiFetch(path, options = {}) {
  const initData = tg.initData;
  try {
    const res = await fetch(path, {
      ...options,
      headers: {
        "Authorization": "tma " + initData,
        "Content-Type": "application/json",
        ...(options.headers || {}),
      },
    });
    return res;
  } catch (e) {
    showError("Ошибка сети. Попробуйте ещё раз.");
    return { ok: false };
  }
}

function showError(msg) {
  const el = document.getElementById("error");
  if (el) { el.textContent = msg; el.style.display = "block"; }
}

function hideError() {
  const el = document.getElementById("error");
  if (el) el.style.display = "none";
}
