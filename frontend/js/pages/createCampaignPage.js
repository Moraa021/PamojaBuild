import { taskActions } from '../state/taskStore.js';
import authStore from '../state/authStore.js';
import { APIErrorDisplay } from '../components/APIErrorDisplay.js';
import { Toast } from '../components/Toast.js';
import { navigate } from '../utils/utils.js';

export function renderCreateCampaignPage(container) {
  container.innerHTML = `
    <section class="create-campaign container">
      <header class="page-header">
        <h1>Create a Campaign</h1>
        <p class="page-header__sub">Define your volunteer task and fundraising goal.</p>
      </header>

      <div id="create-error"></div>

      <form id="create-campaign-form" novalidate aria-label="Create campaign form">

        <fieldset class="form-section">
          <legend>Campaign details</legend>

          <div class="form-field">
            <label for="field-title">Title <span aria-hidden="true">*</span></label>
            <input id="field-title" name="title" type="text" required
                   autocomplete="off" placeholder="e.g. Clean-up Nairobi River Banks" />
            <div class="form-field__error" role="alert" aria-live="polite"></div>
          </div>

          <div class="form-field">
            <label for="field-description">Description <span aria-hidden="true">*</span></label>
            <textarea id="field-description" name="description" rows="5" required
                      placeholder="Describe the task, goals, and how volunteers can help."></textarea>
            <div class="form-field__error" role="alert" aria-live="polite"></div>
          </div>

          <div class="form-row">
            <div class="form-field">
              <label for="field-category">Category <span aria-hidden="true">*</span></label>
              <select id="field-category" name="category" required>
                <option value="">Select a category</option>
                <option value="environment">Environment</option>
                <option value="education">Education</option>
                <option value="health">Health</option>
                <option value="infrastructure">Infrastructure</option>
                <option value="community">Community</option>
                <option value="other">Other</option>
              </select>
              <div class="form-field__error" role="alert" aria-live="polite"></div>
            </div>

            <div class="form-field">
              <label for="field-region">Region <span aria-hidden="true">*</span></label>
              <input id="field-region" name="region" type="text" required
                     placeholder="e.g. Nairobi, Kenya" />
              <div class="form-field__error" role="alert" aria-live="polite"></div>
            </div>
          </div>

          <div class="form-field">
            <label for="field-location-detail">Location detail <span class="label--optional">(optional)</span></label>
            <input id="field-location-detail" name="location_detail" type="text"
                   placeholder="e.g. Uhuru Park main entrance" />
          </div>
        </fieldset>

        <fieldset class="form-section">
          <legend>Funding & volunteers</legend>

          <div class="form-row">
            <div class="form-field">
              <label for="field-goal-sats">Fundraising goal (sats) <span class="label--optional">(optional)</span></label>
              <input id="field-goal-sats" name="goal_sats" type="number" min="1" placeholder="e.g. 500000" />
            </div>

            <div class="form-field">
              <label for="field-max-volunteers">Max volunteers <span aria-hidden="true">*</span></label>
              <input id="field-max-volunteers" name="max_volunteers" type="number" min="0" required placeholder="0 for unlimited" />
              <div class="form-field__error" role="alert" aria-live="polite"></div>
            </div>
          </div>

          <div class="form-field">
            <label for="field-volunteer-mode">Volunteer mode <span aria-hidden="true">*</span></label>
            <select id="field-volunteer-mode" name="volunteer_mode" required>
              <option value="">Select mode</option>
              <option value="open">Open — anyone can join</option>
              <option value="approval_required">Approval required</option>
            </select>
            <div class="form-field__error" role="alert" aria-live="polite"></div>
          </div>
        </fieldset>

        <div class="form-actions">
          <button type="button" class="btn btn--ghost" id="cancel-btn">Cancel</button>
          <button type="submit" class="btn btn--primary" id="submit-btn">
            <span class="btn__label">Create Campaign</span>
            <span class="btn__loading" aria-hidden="true" hidden>Creating…</span>
          </button>
        </div>
      </form>
    </section>
  `;

  const form      = container.querySelector('#create-campaign-form');
  const submitBtn = container.querySelector('#submit-btn');
  const errDisplay = new APIErrorDisplay(container.querySelector('#create-error'));

  container.querySelector('#cancel-btn').addEventListener('click', () => navigate('/'));

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    errDisplay.clear();

    if (!validateForm(form)) return;

    const data    = new FormData(form);
    const userId  = authStore.state.userId;
    if (!userId) {
      Toast.show({ message: 'You must be signed in to create a campaign.', type: 'error' });
      return;
    }

    const payload = {
      creator_id:      Number(userId),
      title:           data.get('title').trim(),
      description:     data.get('description').trim(),
      category:        data.get('category'),
      region:          data.get('region').trim(),
      location_detail: data.get('location_detail').trim() || undefined,
      goal_sats:       data.get('goal_sats') ? Number(data.get('goal_sats')) : undefined,
      max_volunteers:  Number(data.get('max_volunteers')),
      volunteer_mode:  data.get('volunteer_mode'),
    };

    setLoading(submitBtn, true);
    try {
      const task = await taskActions.createTask(payload);
      Toast.show({ message: 'Campaign created successfully!', type: 'success' });
      navigate(`/tasks/${task.slug}`);
    } catch (err) {
      errDisplay.show(err);
      setLoading(submitBtn, false);
    }
  });
}

function validateForm(form) {
  let valid = true;
  form.querySelectorAll('[required]').forEach(field => {
    const errEl = field.closest('.form-field')?.querySelector('.form-field__error');
    if (!field.value.trim()) {
      if (errEl) errEl.textContent = 'This field is required.';
      field.setAttribute('aria-invalid', 'true');
      valid = false;
    } else {
      if (errEl) errEl.textContent = '';
      field.removeAttribute('aria-invalid');
    }
  });
  return valid;
}

function setLoading(btn, loading) {
  btn.disabled = loading;
  btn.querySelector('.btn__label').hidden  = loading;
  btn.querySelector('.btn__loading').hidden = !loading;
}
