import { taskActions } from '../state/taskStore.js';
import { applicationActions } from '../state/volunteerStore.js';
import { APIErrorDisplay } from '../components/APIErrorDisplay.js';
import { Loader } from '../components/Loader.js';
import { Modal } from '../components/Modal.js';
import { Toast } from '../components/Toast.js';
import { formatSats, formatDate, getPathParam, navigate } from '../utils/utils.js';

export async function renderVolunteerTaskDetailsPage(container) {
  const slug = getPathParam(2); // /volunteer/tasks/:slug
  if (!slug) {
    container.innerHTML = '<section class="container"><p>Invalid URL.</p></section>';
    return;
  }

  container.innerHTML = `
    <section class="task-detail container">
      <div id="vtd-error"></div>
      <div id="vtd-loading"></div>
    </section>
  `;

  const section    = container.querySelector('.task-detail');
  const errDisplay = new APIErrorDisplay(container.querySelector('#vtd-error'), {
    onRetry: () => renderVolunteerTaskDetailsPage(container),
  });
  const spinner = Loader.inline(container.querySelector('#vtd-loading'), 'Loading task…');

  let task;
  try {
    task = await taskActions.fetchTask(slug);
    spinner.remove();
  } catch (err) {
    spinner.remove();
    errDisplay.show(err);
    return;
  }

  section.innerHTML = `
    <header class="page-header">
      <button class="btn btn--ghost btn--sm" id="back-btn">← Browse tasks</button>
    </header>

    <div class="task-detail__header">
      ${task.image_path ? `<img class="task-detail__image" src="${esc(task.image_path)}" alt="${esc(task.title)} image" />` : ''}
      <div class="task-detail__header-body">
        <div class="task-detail__badges">
          <span class="badge badge--category">${esc(task.category)}</span>
          <span class="badge badge--status badge--status-${task.status}">${esc(task.status)}</span>
        </div>
        <h1 class="task-detail__title">${esc(task.title)}</h1>
        <p class="task-detail__region">${esc(task.region)}${task.location_detail ? ` · ${esc(task.location_detail)}` : ''}</p>
      </div>
    </div>

    <div class="task-detail__body">
      <article class="task-detail__description">
        <h2>About this task</h2>
        <p>${esc(task.description)}</p>
      </article>

      <aside class="task-detail__sidebar">
        <div class="info-card">
          <dl class="info-card__dl">
            <div><dt>Goal</dt><dd>${task.goal_sats ? formatSats(task.goal_sats) : 'Not specified'}</dd></div>
            <div><dt>Max volunteers</dt><dd>${task.max_volunteers > 0 ? task.max_volunteers : 'Unlimited'}</dd></div>
            <div><dt>Mode</dt><dd>${task.volunteer_mode === 'approval_required' ? 'Approval required' : 'Open'}</dd></div>
            <div><dt>Created</dt><dd>${formatDate(task.created_at)}</dd></div>
          </dl>

          <div class="info-card__actions" id="task-actions"></div>
        </div>
      </aside>
    </div>
  `;

  section.querySelector('#back-btn').addEventListener('click', () => navigate('/volunteer/tasks'));

  const actionsEl = section.querySelector('#task-actions');

  if (task.status === 'open') {
    const applyBtn = document.createElement('button');
    applyBtn.className   = 'btn btn--primary btn--full';
    applyBtn.textContent = 'Apply to volunteer';
    applyBtn.addEventListener('click', () => openApplyModal(task.slug));
    actionsEl.appendChild(applyBtn);
  } else if (task.status === 'in_progress') {
    const submitBtn = document.createElement('a');
    submitBtn.className = 'btn btn--primary btn--full';
    submitBtn.href      = `/volunteer/tasks/${task.slug}/submit`;
    submitBtn.textContent = 'Submit work evidence';
    actionsEl.appendChild(submitBtn);
  } else {
    const note = document.createElement('p');
    note.className = 'info-card__inactive';
    note.textContent = `This task is currently ${task.status.replace('_', ' ')} and not accepting new volunteers.`;
    actionsEl.appendChild(note);
  }
}

function openApplyModal(slug) {
  const form = document.createElement('div');
  form.innerHTML = `
    <div class="form-field">
      <label for="apply-message">Why do you want to volunteer? <span class="label--optional">(optional)</span></label>
      <textarea id="apply-message" rows="4" placeholder="Describe your motivation and relevant experience."></textarea>
    </div>
  `;

  new Modal({
    title:        'Apply to volunteer',
    content:      form,
    confirmLabel: 'Submit application',
    onConfirm:    async () => {
      const message = form.querySelector('#apply-message').value.trim();
      await applicationActions.applyToTask(slug, message);
      Toast.show({ message: 'Application submitted!', type: 'success' });
    },
  }).open();
}

function esc(str) {
  const d = document.createElement('div');
  d.textContent = str || '';
  return d.innerHTML;
}
