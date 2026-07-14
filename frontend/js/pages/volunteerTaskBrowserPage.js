import { taskBrowserActions, taskBrowserStore } from '../state/volunteerStore.js';
import { TaskCard } from '../components/TaskCard.js';
import { TaskFilter } from '../components/TaskFilter.js';
import { APIErrorDisplay } from '../components/APIErrorDisplay.js';
import { Loader } from '../components/Loader.js';
import { navigate } from '../utils/utils.js';

export async function renderVolunteerTaskBrowserPage(container) {
  container.innerHTML = `
    <section class="task-browser container">
      <header class="page-header">
        <h1>Browse Tasks</h1>
        <p class="page-header__sub">Find volunteer opportunities that match your skills and location.</p>
      </header>
      <div id="browser-error"></div>
      <div id="filter-container"></div>
      <div id="browser-loading"></div>
      <div id="task-grid" class="task-grid" role="list" aria-label="Available tasks"></div>
    </section>
  `;

  const errDisplay = new APIErrorDisplay(container.querySelector('#browser-error'), {
    onRetry: () => renderVolunteerTaskBrowserPage(container),
  });

  // Mount filter
  new TaskFilter(container.querySelector('#filter-container'), {
    onChange: (filters) => {
      taskBrowserActions.setFilter(filters);
      renderCards();
    },
  });

  const spinner = Loader.inline(container.querySelector('#browser-loading'), 'Loading tasks…');

  try {
    await taskBrowserActions.fetchTasks();
    spinner.remove();
  } catch (err) {
    spinner.remove();
    errDisplay.show(err);
    return;
  }

  function renderCards() {
    const grid  = container.querySelector('#task-grid');
    const tasks = taskBrowserActions.getFilteredTasks();
    grid.innerHTML = '';

    if (!tasks.length) {
      grid.innerHTML = `<p class="task-grid__empty">No tasks match the current filters.</p>`;
      return;
    }

    tasks.forEach(task => {
      const card = new TaskCard(task, {
        onClick: () => navigate(`/volunteer/tasks/${task.slug}`),
      });
      grid.appendChild(card.element);
    });
  }

  // Re-render on store changes (filter updates)
  taskBrowserStore.subscribe(() => renderCards());
  renderCards();
}
