import { taskBrowserActions, taskBrowserStore } from '../state/volunteerStore.js';
import { APIErrorDisplay } from '../components/APIErrorDisplay.js';
import { Loader }          from '../components/Loader.js';
import { formatDate }      from '../utils/utils.js';

const STATUS_ORDER = ['open', 'in_progress', 'pending_verification', 'completed'];
const FIN_ORDER    = ['ACTIVE', 'LIQUIDATING', 'READY_FOR_PAYOUT', 'SYSTEM_LOCKDOWN', 'ARCHIVED'];

export async function renderAdminPage(container) {
  container.innerHTML = `
    <section class="admin-page container">
      <header class="page-header">
        <h1>Admin Overview</h1>
        <p class="page-header__sub">System-wide task and financial state summary.</p>
      </header>
      <div id="admin-error"></div>
      <div id="admin-loading"></div>
    </section>
  `;

  const section    = container.querySelector('.admin-page');
  const errDisplay = new APIErrorDisplay(container.querySelector('#admin-error'), {
    onRetry: () => renderAdminPage(container),
  });
  const spinner = Loader.inline(container.querySelector('#admin-loading'), 'Loading system data…');

  try {
    await taskBrowserActions.fetchTasks();
    spinner.remove();
  } catch (err) {
    spinner.remove();
    errDisplay.show(err);
    return;
  }

  const tasks = taskBrowserStore.state.tasks;

  // Aggregate counts
  const byStatus = STATUS_ORDER.reduce((acc, s) => {
    acc[s] = tasks.filter(t => t.status === s).length;
    return acc;
  }, {});

  const byFinancial = FIN_ORDER.reduce((acc, s) => {
    acc[s] = tasks.filter(t => t.financial_state === s).length;
    return acc;
  }, {});

  const flaggedTasks = tasks.filter(t =>
    t.financial_state === 'SYSTEM_LOCKDOWN' || t.status === 'pending_verification'
  );

  section.innerHTML += `
    <div class="admin-grid">

      <!-- Task status breakdown -->
      <div class="admin-card">
        <h2 class="admin-card__title">Tasks by status</h2>
        <dl class="admin-stat-list">
          ${STATUS_ORDER.map(s => `
            <div class="admin-stat-row">
              <dt><span class="badge badge--status badge--status-${s}">${s.replace('_',' ')}</span></dt>
              <dd class="admin-stat-num">${byStatus[s]}</dd>
            </div>
          `).join('')}
        </dl>
      </div>

      <!-- Financial state breakdown -->
      <div class="admin-card">
        <h2 class="admin-card__title">Tasks by financial state</h2>
        <dl class="admin-stat-list">
          ${FIN_ORDER.map(s => `
            <div class="admin-stat-row">
              <dt><span class="badge badge--financial badge--financial-${s}">${s.replace('_',' ')}</span></dt>
              <dd class="admin-stat-num">${byFinancial[s]}</dd>
            </div>
          `).join('')}
        </dl>
      </div>

      <!-- Total -->
      <div class="admin-card admin-card--highlight">
        <h2 class="admin-card__title">Total tasks</h2>
        <p class="admin-big-num">${tasks.length}</p>
      </div>

      <!-- Quick links -->
      <div class="admin-card">
        <h2 class="admin-card__title">Quick actions</h2>
        <div class="admin-links">
          <a href="/volunteer/tasks"  class="btn btn--ghost">All tasks</a>
          <a href="/trustee/payout"   class="btn btn--ghost">Payout review</a>
          <a href="/campaigns/new"    class="btn btn--ghost">Create campaign</a>
          <a href="/trustee/register" class="btn btn--ghost">Register trustee</a>
        </div>
      </div>

    </div>

    <!-- Flagged tasks -->
    ${flaggedTasks.length ? `
      <div class="admin-flagged">
        <h2 class="admin-flagged__title">⚠ Requires attention (${flaggedTasks.length})</h2>
        <div class="admin-flagged__list">
          ${flaggedTasks.map(t => `
            <a href="/tasks/${esc(t.slug)}" class="admin-flagged__row">
              <code class="mono">${esc(t.slug)}</code>
              <span class="badge badge--status badge--status-${t.status}">${t.status.replace('_',' ')}</span>
              <span class="badge badge--financial badge--financial-${t.financial_state}">${t.financial_state.replace('_',' ')}</span>
              <time>${formatDate(t.created_at)}</time>
            </a>
          `).join('')}
        </div>
      </div>
    ` : '<p class="admin-clear">✓ No tasks require immediate attention.</p>'}
  `;
}

function esc(str) {
  const d = document.createElement('div');
  d.textContent = str || '';
  return d.innerHTML;
}
