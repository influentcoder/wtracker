'use strict';

const REFRESH_INTERVAL = 60_000; // 60 seconds
let refreshTimer = null;

// ── Utilities ─────────────────────────────────

function fmt(n, decimals = 4) {
  if (n == null) return '—';
  return n.toLocaleString('en-US', { minimumFractionDigits: decimals, maximumFractionDigits: decimals });
}

function fmtUSD(n) {
  if (!n) return '$0';
  return n.toLocaleString('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 });
}

function fmtBTC(n) {
  if (n == null) return '—';
  const abs = Math.abs(n);
  if (abs >= 1000) return fmt(n, 2);
  if (abs >= 1) return fmt(n, 4);
  return fmt(n, 6);
}

function shortAddr(addr) {
  if (!addr) return '';
  if (addr.length <= 20) return addr;
  return addr.slice(0, 10) + '…' + addr.slice(-8);
}

function timeAgo(dateStr) {
  if (!dateStr) return '';
  const d = new Date(dateStr);
  if (isNaN(d)) return '';
  const sec = Math.floor((Date.now() - d) / 1000);
  if (sec < 60) return `${sec}s ago`;
  if (sec < 3600) return `${Math.floor(sec / 60)}m ago`;
  if (sec < 86400) return `${Math.floor(sec / 3600)}h ago`;
  return `${Math.floor(sec / 86400)}d ago`;
}

function formatDate(dateStr) {
  if (!dateStr) return 'Pending';
  const d = new Date(dateStr);
  if (isNaN(d) || d.getFullYear() < 2009) return 'Pending';
  return d.toLocaleString('en-US', { month: 'short', day: 'numeric', year: 'numeric', hour: '2-digit', minute: '2-digit' });
}

// ── Skeleton Loaders ──────────────────────────

function renderSkeletons(count = 6) {
  const grid = document.getElementById('whale-grid');
  grid.innerHTML = Array.from({ length: count }, () => `
    <div class="skeleton-card">
      <div class="skeleton skeleton-line medium"></div>
      <div class="skeleton skeleton-line full" style="margin-top:6px;height:10px"></div>
      <div style="margin:20px 0 10px">
        <div class="skeleton skeleton-line large"></div>
        <div class="skeleton skeleton-line short" style="margin-top:6px;height:10px"></div>
      </div>
      <div class="skeleton skeleton-line full" style="height:1px"></div>
      <div class="skeleton skeleton-line medium" style="margin-top:12px;height:10px"></div>
    </div>
  `).join('');
}

// ── Price ─────────────────────────────────────

async function fetchPrice() {
  try {
    const res = await fetch('/api/price');
    if (!res.ok) return null;
    return await res.json();
  } catch {
    return null;
  }
}

function renderPrice(data) {
  const el = document.getElementById('btc-price');
  if (!el) return;
  if (!data || !data.btc_usd) {
    el.textContent = 'BTC: —';
    return;
  }
  el.textContent = `BTC ${fmtUSD(data.btc_usd)}`;
}

// ── Whale Cards ───────────────────────────────

async function fetchWhales() {
  const res = await fetch('/api/whales');
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return await res.json();
}

function renderWhaleCards(whales) {
  const grid = document.getElementById('whale-grid');

  if (!whales || whales.length === 0) {
    grid.innerHTML = '<p style="color:var(--text-muted);grid-column:1/-1">No whales configured.</p>';
    return;
  }

  grid.innerHTML = whales.map(w => {
    const hasError = !!w.error;
    return `
      <div class="whale-card ${hasError ? 'error' : ''}" data-address="${w.address}" role="button" tabindex="0">
        <div class="card-header">
          <span class="card-label">${escHtml(w.label)}</span>
          <span class="chain-badge">${escHtml(w.chain)}</span>
        </div>
        <div class="card-address">${escHtml(w.address)}</div>
        ${hasError ? `
          <div class="error-msg">${escHtml(w.error)}</div>
        ` : `
          <div class="card-balance">
            <div class="balance-btc">
              ${fmtBTC(w.btc_balance)}
              <span class="unit">BTC</span>
            </div>
            <div class="balance-usd">${fmtUSD(w.usd_value)}</div>
          </div>
          <div class="card-footer">
            <span class="tx-count">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="17 1 21 5 17 9"></polyline>
                <path d="M3 11V9a4 4 0 0 1 4-4h14"></path>
                <polyline points="7 23 3 19 7 15"></polyline>
                <path d="M21 13v2a4 4 0 0 1-4 4H3"></path>
              </svg>
              ${w.tx_count?.toLocaleString() ?? '—'} txns
            </span>
            <span class="view-details">View details →</span>
          </div>
        `}
      </div>
    `;
  }).join('');

  // Bind click & keyboard handlers
  grid.querySelectorAll('.whale-card:not(.error)').forEach(card => {
    const open = () => openModal(card.dataset.address);
    card.addEventListener('click', open);
    card.addEventListener('keydown', e => { if (e.key === 'Enter' || e.key === ' ') open(); });
  });
}

// ── Summary Stats Bar ─────────────────────────

function renderStats(whales, btcPrice) {
  const tracked = whales.length;
  const totalBTC = whales.reduce((s, w) => s + (w.btc_balance || 0), 0);
  const totalUSD = totalBTC * (btcPrice || 0);

  document.getElementById('stat-whales').textContent = tracked;
  document.getElementById('stat-btc').textContent = fmtBTC(totalBTC) + ' BTC';
  document.getElementById('stat-usd').textContent = fmtUSD(totalUSD);
}

// ── Modal ─────────────────────────────────────

async function openModal(address) {
  const overlay = document.getElementById('modal-overlay');
  const body = document.getElementById('modal-body');

  overlay.classList.add('open');
  body.innerHTML = `<p style="color:var(--text-muted);text-align:center;padding:40px 0">Loading…</p>`;

  try {
    const res = await fetch(`/api/whales/${encodeURIComponent(address)}`);
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();
    renderModalContent(data);
  } catch (err) {
    body.innerHTML = `<p class="error-msg" style="text-align:center;padding:40px 0">Failed to load: ${escHtml(err.message)}</p>`;
  }
}

function renderModalContent(data) {
  // Update modal header
  document.getElementById('modal-title').textContent = data.label;
  document.getElementById('modal-address').textContent = data.address;

  // Stats
  document.getElementById('modal-btc').textContent = fmtBTC(data.btc_balance) + ' BTC';
  document.getElementById('modal-usd').textContent = fmtUSD(data.usd_value);
  document.getElementById('modal-txcount').textContent = data.tx_count?.toLocaleString() ?? '—';

  // Transactions
  const body = document.getElementById('modal-body');
  const txs = data.recent_transactions;

  if (!txs || txs.length === 0) {
    body.innerHTML = '<p class="no-txs">No recent transactions found.</p>';
    return;
  }

  body.innerHTML = `
    <p class="section-title">Recent Transactions</p>
    <table class="tx-table">
      <thead>
        <tr>
          <th>TX Hash</th>
          <th>Date</th>
          <th>Amount (BTC)</th>
          <th>Fee (BTC)</th>
          <th>Status</th>
        </tr>
      </thead>
      <tbody>
        ${txs.map(tx => {
          const dir = tx.direction === 'in';
          const amtClass = dir ? 'amount-in' : 'amount-out';
          const sign = dir ? '+' : '−';
          const abs = Math.abs(tx.btc_amount);
          return `
            <tr>
              <td class="tx-hash">
                <a href="https://blockstream.info/tx/${escHtml(tx.txid)}" target="_blank" rel="noopener"
                   title="${escHtml(tx.txid)}">${shortAddr(tx.txid)}</a>
              </td>
              <td style="color:var(--text-secondary);white-space:nowrap">${formatDate(tx.timestamp)}</td>
              <td class="${amtClass}">${sign}${fmtBTC(abs)}</td>
              <td style="color:var(--text-muted)">${tx.fee_btc ? fmtBTC(tx.fee_btc) : '—'}</td>
              <td>
                <span class="confirmed-badge ${tx.confirmed ? 'yes' : 'no'}">
                  ${tx.confirmed
                    ? '<svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"></polyline></svg> Confirmed'
                    : '<svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><circle cx="12" cy="12" r="10"></circle></svg> Pending'}
                </span>
              </td>
            </tr>
          `;
        }).join('')}
      </tbody>
    </table>
  `;
}

function closeModal() {
  document.getElementById('modal-overlay').classList.remove('open');
}

// ── XSS protection ────────────────────────────

function escHtml(str) {
  return String(str ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

// ── Main refresh cycle ────────────────────────

async function refresh(showSkeletons = false) {
  const btn = document.getElementById('refresh-btn');
  btn.classList.add('spinning');

  if (showSkeletons) renderSkeletons();

  try {
    const [whales, priceData] = await Promise.all([fetchWhales(), fetchPrice()]);
    renderWhaleCards(whales);
    renderPrice(priceData);
    renderStats(whales, priceData?.btc_usd ?? 0);

    document.getElementById('last-updated').textContent =
      'Updated ' + new Date().toLocaleTimeString();
  } catch (err) {
    console.error('refresh error:', err);
    if (showSkeletons) {
      document.getElementById('whale-grid').innerHTML =
        `<p style="color:var(--red);grid-column:1/-1">Failed to load whale data. Retrying…</p>`;
    }
  } finally {
    btn.classList.remove('spinning');
  }
}

// ── Init ──────────────────────────────────────

document.addEventListener('DOMContentLoaded', () => {
  // Initial load
  refresh(true);

  // Auto-refresh
  refreshTimer = setInterval(() => refresh(false), REFRESH_INTERVAL);

  // Manual refresh button
  document.getElementById('refresh-btn').addEventListener('click', () => {
    clearInterval(refreshTimer);
    refresh(false);
    refreshTimer = setInterval(() => refresh(false), REFRESH_INTERVAL);
  });

  // Modal close
  document.getElementById('modal-overlay').addEventListener('click', e => {
    if (e.target === e.currentTarget) closeModal();
  });
  document.getElementById('modal-close').addEventListener('click', closeModal);
  document.addEventListener('keydown', e => {
    if (e.key === 'Escape') closeModal();
  });
});
