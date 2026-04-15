// ── Date field management ───────────────────────────────────────
function addDateField() {
  const list = document.getElementById('dates-list');
  if (!list) return;
  const input = document.createElement('input');
  input.className = 'date-input';
  input.name = 'dates';
  input.type = 'date';
  list.appendChild(input);
  input.focus();
}

// ── Emoji picker ────────────────────────────────────────────────
function selectEmoji(btn, emoji) {
  document.querySelectorAll('.emoji-btn').forEach(b => b.classList.remove('selected'));
  btn.classList.add('selected');
  const input = document.getElementById('emoji-input');
  if (input) input.value = emoji;
}

// ── Share link ───────────────────────────────────────────────────
function copyLink() {
  const url = window.location.href;
  navigator.clipboard.writeText(url).then(() => {
    const btn = document.getElementById('share-btn');
    if (!btn) return;
    const orig = btn.textContent;
    btn.textContent = '✓ Copied!';
    btn.classList.add('copied');
    setTimeout(() => {
      btn.textContent = orig;
      btn.classList.remove('copied');
    }, 2000);
  }).catch(() => {
    // Fallback for browsers without clipboard API
    prompt('Copy this link:', url);
  });
}

// ── Highlight first emoji on page load ──────────────────────────
document.addEventListener('DOMContentLoaded', () => {
  const firstEmoji = document.querySelector('.emoji-btn');
  if (firstEmoji) firstEmoji.classList.add('selected');
});

// ── Realtime updates ────────────────────────────────────────────
document.addEventListener('DOMContentLoaded', () => {
  const el = document.querySelector('[data-event-id]');
  if (!el) return;

  const eventId = el.dataset.eventId;
  const pb = new PocketBase(window.location.origin);

  // Votes & date_options: smooth swap of just the results div
  let refreshTimer = null;
  function scheduleRefresh() {
    clearTimeout(refreshTimer);
    refreshTimer = setTimeout(() => {
      htmx.ajax('GET', `/events/${eventId}/results`, {
        target: '#results',
        swap: 'outerHTML',
      });
    }, 150);
  }

  pb.collection('date_options').subscribe('*', scheduleRefresh, {
    filter: `event_id = '${eventId}'`,
  });

  // Subscribe to votes only for date options belonging to this event.
  // PocketBase realtime doesn't support relation filters, so we fetch
  // the date option IDs first and build an explicit filter.
  pb.collection('date_options').getFullList({ filter: `event_id = '${eventId}'` })
    .then(opts => {
      if (!opts.length) return;
      const ids = opts.map(o => `date_option_id = '${o.id}'`).join(' || ');
      pb.collection('votes').subscribe('*', scheduleRefresh, { filter: ids });
    })
    .catch(() => {
      // fallback: subscribe to all votes if prefetch fails
      pb.collection('votes').subscribe('*', scheduleRefresh);
    });

  // Participants joining: reload so "Who's here?" updates
  pb.collection('participants').subscribe('*', () => {
    window.location.reload();
  }, { filter: `event_id = '${eventId}'` });

  // Event locked: reload so the locked banner + button states update
  pb.collection('events').subscribe(eventId, () => {
    window.location.reload();
  });
});
