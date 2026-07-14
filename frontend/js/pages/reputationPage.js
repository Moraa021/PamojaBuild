import { profileActions, paymentActions, applicationActions, submissionActions, profileStore, paymentStore, applicationStore, submissionStore } from '../state/volunteerStore.js';
import { ReputationBadge } from '../components/ReputationBadge.js';
import { APIErrorDisplay } from '../components/APIErrorDisplay.js';
import { Loader } from '../components/Loader.js';
import { formatSats } from '../utils/utils.js';

export async function renderReputationPage(container) {
  container.innerHTML = `
    <section class="reputation-page container">
      <header class="page-header">
        <h1>Reputation</h1>
        <p class="page-header__sub">Your volunteer track record, computed from verified backend data.</p>
      </header>
      <div id="rep-error"></div>
      <div id="rep-loading"></div>
    </section>
  `;

  const section    = container.querySelector('.reputation-page');
  const errDisplay = new APIErrorDisplay(container.querySelector('#rep-error'), {
    onRetry: () => renderReputationPage(container),
  });
  const spinner = Loader.inline(container.querySelector('#rep-loading'), 'Computing reputation…');

  try {
    await Promise.all([
      profileActions.fetchProfile(),
      paymentActions.fetchPayments(),
      applicationActions.fetchApplications(),
    ]);
    spinner.remove();
  } catch (err) {
    spinner.remove();
    errDisplay.show(err);
    return;
  }

  const { profile }      = profileStore.state;
  const { payments }     = paymentStore.state;
  const { applications } = applicationStore.state;

  // Derived metrics (all from real backend data)
  const completedTasks    = payments.filter(p => p.status === 'completed').length;
  const totalEarned       = payments.filter(p => p.status === 'completed').reduce((s, p) => s + p.amount_sats, 0);
  const approvedApps      = applications.filter(a => a.status === 'approved').length;
  const approvalRate      = applications.length ? Math.round((approvedApps / applications.length) * 100) : 0;

  section.innerHTML += `
    <div class="reputation-page__layout">
      <div class="rep-hero">
        <div id="rep-badge-hero"></div>
        <h2 class="rep-hero__name">${esc(profile.display_name)}</h2>
        <p class="rep-hero__score-label">Overall reputation score</p>
        <p class="rep-hero__score">${profile.reputation_score.toLocaleString()}</p>
      </div>

      <div class="rep-metrics">
        <h2 class="dash-section__title">Metrics</h2>
        <dl class="rep-metrics__grid">
          <div class="rep-metric">
            <dt>Completed tasks</dt>
            <dd>${completedTasks}</dd>
          </div>
          <div class="rep-metric">
            <dt>Total earned</dt>
            <dd class="value-sats">${formatSats(totalEarned)}</dd>
          </div>
          <div class="rep-metric">
            <dt>Applications submitted</dt>
            <dd>${applications.length}</dd>
          </div>
          <div class="rep-metric">
            <dt>Approval rate</dt>
            <dd>
              <div class="rep-bar" role="progressbar" aria-valuenow="${approvalRate}" aria-valuemin="0" aria-valuemax="100">
                <div class="rep-bar__fill" style="width: ${approvalRate}%"></div>
              </div>
              <span>${approvalRate}%</span>
            </dd>
          </div>
        </dl>
      </div>
    </div>
  `;

  const repBadge = new ReputationBadge({ score: profile.reputation_score });
  section.querySelector('#rep-badge-hero').appendChild(repBadge.element);
}

function esc(str) {
  const d = document.createElement('div');
  d.textContent = str || '';
  return d.innerHTML;
}
