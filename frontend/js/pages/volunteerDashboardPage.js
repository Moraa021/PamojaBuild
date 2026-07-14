import { profileActions, applicationActions, paymentActions, profileStore, applicationStore, paymentStore } from '../state/volunteerStore.js';
import { APIErrorDisplay } from '../components/APIErrorDisplay.js';
import { Loader } from '../components/Loader.js';
import { ReputationBadge } from '../components/ReputationBadge.js';
import { ApplicationStatusBadge } from '../components/ApplicationStatusBadge.js';
import { PaymentSummaryCard } from '../components/PaymentSummaryCard.js';
import { formatSats, formatDate, navigate } from '../utils/utils.js';

export async function renderVolunteerDashboardPage(container) {
  container.innerHTML = `
    <section class="vol-dashboard container">
      <div id="dash-error"></div>
      <div id="dash-loading"></div>
    </section>
  `;

  const section    = container.querySelector('.vol-dashboard');
  const errDisplay = new APIErrorDisplay(container.querySelector('#dash-error'), {
    onRetry: () => renderVolunteerDashboardPage(container),
  });
  const spinner = Loader.inline(container.querySelector('#dash-loading'), 'Loading your dashboard…');

  try {
    await Promise.all([
      profileActions.fetchProfile(),
      applicationActions.fetchApplications(),
      paymentActions.fetchPayments(),
    ]);
    spinner.remove();
  } catch (err) {
    spinner.remove();
    errDisplay.show(err);
    return;
  }

  const { profile } = profileStore.state;
  const { applications } = applicationStore.state;
  const { payments } = paymentStore.state;

  section.innerHTML = `
    <header class="page-header">
      <div class="dash-header">
        <div>
          <h1>Welcome back, ${esc(profile.display_name)}</h1>
          <p class="page-header__sub">Here is a summary of your volunteer activity.</p>
        </div>
        <div id="rep-badge-container"></div>
      </div>
    </header>

    <div class="vol-dashboard__grid">

      <div class="dash-section">
        <h2 class="dash-section__title">Payment summary</h2>
        <div id="payment-summary"></div>
      </div>

      <div class="dash-section">
        <h2 class="dash-section__title">Recent applications</h2>
        <div id="recent-applications">
          ${applications.length === 0
            ? '<p class="text-muted">No applications yet. <a href="/volunteer/tasks">Browse tasks.</a></p>'
            : applications.slice(0, 5).map(a => `
                <div class="dash-app-row">
                  <div>
                    <code class="mono">${esc(a.task_slug)}</code>
                    <time class="dash-app-row__date">${formatDate(a.applied_at)}</time>
                  </div>
                  <div id="app-badge-${a.id}"></div>
                </div>
              `).join('')
          }
        </div>
        ${applications.length > 5
          ? `<a href="/volunteer/applications" class="btn btn--ghost btn--sm">View all applications</a>`
          : ''
        }
      </div>

      <div class="dash-section">
        <h2 class="dash-section__title">Quick actions</h2>
        <div class="dash-actions">
          <a class="btn btn--primary" href="/volunteer/tasks">Browse tasks</a>
          <a class="btn btn--ghost"   href="/volunteer/profile">Edit profile</a>
          <a class="btn btn--ghost"   href="/volunteer/payments">Payment history</a>
        </div>
      </div>
    </div>
  `;

  // Mount sub-components
  const repBadge = new ReputationBadge({ score: profile.reputation_score });
  section.querySelector('#rep-badge-container').appendChild(repBadge.element);

  new PaymentSummaryCard(section.querySelector('#payment-summary'), payments);

  applications.slice(0, 5).forEach(a => {
    const container = section.querySelector(`#app-badge-${a.id}`);
    if (container) {
      const badge = new ApplicationStatusBadge(a.status);
      container.appendChild(badge.element);
    }
  });
}

function esc(str) {
  const d = document.createElement('div');
  d.textContent = str || '';
  return d.innerHTML;
}
