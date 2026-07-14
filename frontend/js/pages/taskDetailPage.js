import taskStore, { taskActions } from '../state/taskStore.js';
import { APIErrorDisplay } from '../components/APIErrorDisplay.js';
import { Loader } from '../components/Loader.js';
import { Toast } from '../components/Toast.js';
import { navigate, formatSats, formatDate, getPathParam } from '../utils/utils.js';

const STATUS_LABELS = {
  open: 'Open for Volunteers',
  in_progress: 'In Progress',
  pending_verification: 'Pending Verification',
  completed: 'Completed',
};

export async function renderTaskDetailPage(container) {
  const slug = getPathParam(1); // /tasks/:slug
  if (!slug) {
    container.innerHTML = '<p>Invalid campaign URL.</p>';
    return;
  }

  container.innerHTML = `
    <section class="task-detail container" aria-busy="true" aria-label="Loading campaign">
      <div id="task-detail-error"></div>
      <div id="task-detail-loader"></div>
    </section>
  `;

  const loaderEl = container.querySelector('#task-detail-loader');
  const errorEl  = container.querySelector('#task-detail-error');
  const spinner  = Loader.inline(loaderEl, 'Loading campaign…');
  const errDisplay = new APIErrorDisplay(errorEl, {
    onRetry: () => renderTaskDetailPage(container),
  });

  let task;
  try {
    task = await taskActions.fetchTask(slug);
    spinner.remove();
  } catch (err) {
    spinner.remove();
    errDisplay.show(err);
    return;
  }

  container.querySelector('.task-detail').setAttribute('aria-busy', 'false');
  container.querySelector('.task-detail').innerHTML = buildTaskHTML(task);

  // Wire donate button
  const donateBtn = container.querySelector('[data-action="donate"]');
  if (donateBtn) {
    donateBtn.addEventListener('click', () => navigate(`/tasks/${task.slug}/donate`));
  }
}

function buildTaskHTML(t) {
  const imagePart = t.image_path
    ? `<img class="task-detail__image" src="${t.image_path}" alt="${esc(t.title)} campaign image" />`
    : '';

  return `
    <header class="task-detail__header">
      ${imagePart}
      <div class="task-detail__header-body">
        <div class="task-detail__badges">
          <span class="badge badge--category">${esc(t.category)}</span>
          <span class="badge badge--status badge--status-${t.status}">${STATUS_LABELS[t.status] || t.status}</span>
          <span class="badge badge--financial badge--financial-${t.financial_state}">${esc(t.financial_state)}</span>
        </div>
        <h1 class="task-detail__title">${esc(t.title)}</h1>
        <p class="task-detail__region">${esc(t.region)}${t.location_detail ? ` · ${esc(t.location_detail)}` : ''}</p>
      </div>
    </header>

    <div class="task-detail__body">
      <article class="task-detail__description">
        <h2>About this campaign</h2>
        <p>${esc(t.description)}</p>
      </article>

      <aside class="task-detail__sidebar">
        <div class="info-card">
          <dl class="info-card__dl">
            <div><dt>Goal</dt><dd>${t.goal_sats ? formatSats(t.goal_sats) : 'No goal set'}</dd></div>
            <div><dt>Max volunteers</dt><dd>${t.max_volunteers > 0 ? t.max_volunteers : 'Unlimited'}</dd></div>
            <div><dt>Volunteer mode</dt><dd>${t.volunteer_mode === 'approval_required' ? 'Approval required' : 'Open'}</dd></div>
            <div><dt>Created</dt><dd>${formatDate(t.created_at)}</dd></div>
            <div><dt>Task ID</dt><dd class="mono">${esc(t.slug)}</dd></div>
          </dl>

          ${t.financial_state === 'ACTIVE' || t.financial_state === 'LIQUIDATING'
            ? `<button class="btn btn--primary btn--full" data-action="donate">
                 ⚡ Donate via Lightning
               </button>`
            : `<p class="info-card__inactive">Donations are not currently accepted for this campaign.</p>`
          }
        </div>
      </aside>
    </div>
  `;
}

function esc(str) {
  const d = document.createElement('div');
  d.textContent = str || '';
  return d.innerHTML;
}
