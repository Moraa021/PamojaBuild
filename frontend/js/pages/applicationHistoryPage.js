import { applicationActions, applicationStore } from '../state/volunteerStore.js';
import { ApplicationStatusBadge } from '../components/ApplicationStatusBadge.js';
import { APIErrorDisplay } from '../components/APIErrorDisplay.js';
import { Loader } from '../components/Loader.js';
import { formatDate, navigate } from '../utils/utils.js';

export async function renderApplicationHistoryPage(container) {
  container.innerHTML = `
    <section class="app-history container">
      <header class="page-header">
        <h1>My Applications</h1>
        <p class="page-header__sub">Track the status of every task you have applied for.</p>
      </header>
      <div id="app-error"></div>
      <div id="app-loading"></div>
      <div id="app-list"></div>
    </section>
  `;

  const errDisplay = new APIErrorDisplay(container.querySelector('#app-error'), {
    onRetry: () => renderApplicationHistoryPage(container),
  });
  const spinner = Loader.inline(container.querySelector('#app-loading'), 'Loading applications…');

  try {
    await applicationActions.fetchApplications();
    spinner.remove();
  } catch (err) {
    spinner.remove();
    errDisplay.show(err);
    return;
  }

  const { applications } = applicationStore.state;
  const listEl = container.querySelector('#app-list');

  if (!applications.length) {
    listEl.innerHTML = `
      <div class="empty-state">
        <p>You have not applied for any tasks yet.</p>
        <a href="/volunteer/tasks" class="btn btn--primary">Browse tasks</a>
      </div>
    `;
    return;
  }

  applications.forEach(app => {
    const row = document.createElement('div');
    row.className = 'app-row';
    row.innerHTML = `
      <div class="app-row__main">
        <button class="app-row__slug btn btn--ghost btn--sm" data-slug="${esc(app.task_slug)}">
          ${esc(app.task_slug)}
        </button>
        <p class="app-row__message text-muted">${esc(app.message || '—')}</p>
        <time class="app-row__date">${formatDate(app.applied_at)}</time>
      </div>
      <div class="app-row__badge" id="badge-${app.id}"></div>
    `;

    row.querySelector(`#badge-${app.id}`).appendChild(new ApplicationStatusBadge(app.status).element);
    row.querySelector('.app-row__slug').addEventListener('click', () => navigate(`/volunteer/tasks/${app.task_slug}`));
    listEl.appendChild(row);
  });
}

function esc(str) {
  const d = document.createElement('div');
  d.textContent = str || '';
  return d.innerHTML;
}
