import taskApi             from '../api/taskApi.js';
import { TaskCard }        from '../components/TaskCard.js';
import { APIErrorDisplay } from '../components/APIErrorDisplay.js';
import { Loader }          from '../components/Loader.js';
import { navigate }        from '../utils/utils.js';
import authStore           from '../state/authStore.js';

export async function renderCampaignsPage(container) {
  container.innerHTML = `
    <section class="campaigns-page container">
      <header class="page-header">
        <div class="page-header__row">
          <div>
            <h1>My Campaigns</h1>
            <p class="page-header__sub">Manage your volunteer task campaigns.</p>
          </div>
          <a href="/campaigns/new" class="btn btn--primary">+ New campaign</a>
        </div>
      </header>
      <div id="campaigns-error"></div>
      <div id="campaigns-loading"></div>
      <div id="campaigns-grid" class="task-grid" role="list" aria-label="Your campaigns"></div>
    </section>
  `;

  const errDisplay = new APIErrorDisplay(container.querySelector('#campaigns-error'), {
    onRetry: () => renderCampaignsPage(container),
  });
  const spinner = Loader.inline(container.querySelector('#campaigns-loading'), 'Loading your campaigns…');

  let tasks = [];
  try {
    tasks = await taskApi.list?.() || await fetch(
      `${(await import('../config/env.js')).ENV.API_BASE_URL}/api/v1/tasks`,
      { headers: { Authorization: `Bearer ${authStore.state.token}` } }
    ).then(r => r.json());
    spinner.remove();
  } catch (err) {
    spinner.remove();
    errDisplay.show(err);
    return;
  }

  const userId = authStore.state.userId;
  const mine   = userId ? tasks.filter(t => t.creator_id === Number(userId)) : tasks;
  const grid   = container.querySelector('#campaigns-grid');

  if (!mine.length) {
    grid.innerHTML = `
      <div class="empty-state">
        <p>You have not created any campaigns yet.</p>
        <a href="/campaigns/new" class="btn btn--primary">Create your first campaign</a>
      </div>
    `;
    return;
  }

  mine.forEach(task => {
    const card = new TaskCard(task, {
      onClick: () => navigate(`/tasks/${task.slug}`),
    });
    card.element.setAttribute('role', 'listitem');
    grid.appendChild(card.element);
  });
}
