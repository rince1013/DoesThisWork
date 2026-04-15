// ── Calendar ─────────────────────────────────────────────────────
class Calendar {
  constructor(container) {
    this.container = container;
    this.mode = container.dataset.calendar; // "multi" or "single"
    this.inputName = container.dataset.inputName || (this.mode === 'multi' ? 'dates' : 'date');
    if (!container.id) container.id = 'cal-' + Math.random().toString(36).slice(2, 8);
    this.id = container.id;
    this.selected = new Set();
    this.today = this._localDate(new Date());
    this.current = new Date(this.today.getFullYear(), this.today.getMonth(), 1);
    container._calendar = this;
    this.render();
  }

  _localDate(d) {
    return new Date(d.getFullYear(), d.getMonth(), d.getDate());
  }

  _fmt(d) {
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    return `${y}-${m}-${day}`;
  }

  _parse(str) {
    const [y, m, d] = str.split('-').map(Number);
    return new Date(y, m - 1, d);
  }

  render() {
    this.container.innerHTML = '';
    this.container.appendChild(this._renderGrid());
    if (this.mode !== 'single') this.container.appendChild(this._renderManual());
    if (this.selected.size > 0) this.container.appendChild(this._renderChips());
    this._syncInputs();
  }

  reset() {
    this.selected.clear();
    this.render();
  }

  _renderGrid() {
    const wrap = document.createElement('div');
    wrap.className = 'cal';

    // Header
    const header = document.createElement('div');
    header.className = 'cal-header';

    const prev = document.createElement('button');
    prev.type = 'button'; prev.className = 'cal-nav'; prev.textContent = '‹';
    prev.addEventListener('pointerdown', (e) => { e.preventDefault(); this.current = new Date(this.current.getFullYear(), this.current.getMonth() - 1, 1); this.render(); });

    const title = document.createElement('span');
    title.className = 'cal-title';
    title.textContent = this.current.toLocaleDateString('en-US', { month: 'long', year: 'numeric' });

    const next = document.createElement('button');
    next.type = 'button'; next.className = 'cal-nav'; next.textContent = '›';
    next.addEventListener('pointerdown', (e) => { e.preventDefault(); this.current = new Date(this.current.getFullYear(), this.current.getMonth() + 1, 1); this.render(); });

    header.appendChild(prev); header.appendChild(title); header.appendChild(next);
    wrap.appendChild(header);

    // Grid
    const grid = document.createElement('div');
    grid.className = 'cal-grid';

    ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa'].forEach(d => {
      const el = document.createElement('div');
      el.className = 'cal-day-label';
      el.textContent = d;
      grid.appendChild(el);
    });

    const year = this.current.getFullYear();
    const month = this.current.getMonth();
    const firstDay = new Date(year, month, 1).getDay();
    const daysInMonth = new Date(year, month + 1, 0).getDate();
    const todayStr = this._fmt(this.today);

    for (let i = 0; i < firstDay; i++) grid.appendChild(document.createElement('div'));

    for (let d = 1; d <= daysInMonth; d++) {
      const date = new Date(year, month, d);
      const dateStr = this._fmt(date);
      const isPast = date < this.today;
      const cell = document.createElement('button');
      cell.type = 'button';
      cell.className = 'cal-day';
      if (isPast) cell.classList.add('cal-past');
      if (dateStr === todayStr) cell.classList.add('cal-today');
      if (this.selected.has(dateStr)) cell.classList.add('cal-selected');
      cell.textContent = d;
      cell.dataset.date = dateStr;
      cell.disabled = isPast;
      if (!isPast) {
        cell.addEventListener('pointerdown', (e) => {
          e.preventDefault();
          this._toggle(dateStr, cell);
        });
      }
      grid.appendChild(cell);
    }

    wrap.appendChild(grid);
    return wrap;
  }

  _renderManual() {
    const row = document.createElement('div');
    row.className = 'cal-manual';

    const input = document.createElement('input');
    input.type = 'date';
    input.className = 'date-input';

    const btn = document.createElement('button');
    btn.type = 'button'; btn.className = 'btn-ghost'; btn.textContent = '+ Add date';
    btn.onclick = () => { if (input.value) { this._add(input.value); input.value = ''; } };

    row.appendChild(input); row.appendChild(btn);
    return row;
  }

  _renderChips() {
    const wrap = document.createElement('div');
    wrap.className = 'cal-chips';

    [...this.selected].sort().forEach(dateStr => {
      const chip = document.createElement('div');
      chip.className = 'cal-chip';

      const label = document.createElement('span');
      label.textContent = this._parse(dateStr).toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });

      const rm = document.createElement('button');
      rm.type = 'button'; rm.className = 'cal-chip-remove'; rm.textContent = '✕';
      rm.setAttribute('aria-label', 'Remove ' + label.textContent);
      rm.addEventListener('pointerdown', (e) => {
        e.preventDefault();
        this.selected.delete(dateStr);
        const cell = this.container.querySelector(`.cal-day[data-date="${dateStr}"]`);
        if (cell) cell.classList.remove('cal-selected');
        this._updateChips();
        this._syncInputs();
      });

      chip.appendChild(label); chip.appendChild(rm);
      wrap.appendChild(chip);
    });

    return wrap;
  }

  _toggle(dateStr, cell) {
    if (this.mode === 'single') {
      this.container.querySelectorAll('.cal-day.cal-selected').forEach(el => el.classList.remove('cal-selected'));
      this.selected.clear();
      this.selected.add(dateStr);
      cell.classList.add('cal-selected');
    } else {
      if (this.selected.has(dateStr)) {
        this.selected.delete(dateStr);
        cell.classList.remove('cal-selected');
      } else {
        this.selected.add(dateStr);
        cell.classList.add('cal-selected');
      }
    }
    this._updateChips();
    this._syncInputs();
  }

  _updateChips() {
    const existing = this.container.querySelector('.cal-chips');
    if (existing) existing.remove();
    if (this.selected.size > 0) this.container.appendChild(this._renderChips());
  }

  _add(dateStr) {
    if (this.mode === 'single') this.selected.clear();
    this.selected.add(dateStr);
    const d = this._parse(dateStr);
    this.current = new Date(d.getFullYear(), d.getMonth(), 1);
    this.render();
  }

  _syncInputs() {
    const form = this.container.closest('form');
    if (!form) return;
    form.querySelectorAll(`input[data-cal="${this.id}"]`).forEach(el => el.remove());
    this.selected.forEach(dateStr => {
      const input = document.createElement('input');
      input.type = 'hidden';
      input.name = this.inputName;
      input.value = dateStr;
      input.dataset.cal = this.id;
      form.appendChild(input);
    });
  }
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

  // Init all calendars
  document.querySelectorAll('[data-calendar]').forEach(el => new Calendar(el));
});

// Reset single-mode calendars after a successful HTMX date submission
document.addEventListener('htmx:afterRequest', (e) => {
  if (!e.detail.successful || e.detail.requestConfig.verb !== 'post') return;
  const form = e.detail.elt.closest('form') || e.detail.elt;
  if (form) {
    form.querySelectorAll('[data-calendar="single"]').forEach(el => {
      if (el._calendar) el._calendar.reset();
    });
  }
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
