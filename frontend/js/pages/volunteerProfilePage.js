import { profileActions, profileStore } from '../state/volunteerStore.js';
import { ReputationBadge } from '../components/ReputationBadge.js';
import { APIErrorDisplay } from '../components/APIErrorDisplay.js';
import { Loader } from '../components/Loader.js';
import { Toast } from '../components/Toast.js';

export async function renderVolunteerProfilePage(container) {
  container.innerHTML = `
    <section class="vol-profile container">
      <div id="profile-error"></div>
      <div id="profile-loading"></div>
    </section>
  `;

  const section    = container.querySelector('.vol-profile');
  const errDisplay = new APIErrorDisplay(container.querySelector('#profile-error'), {
    onRetry: () => renderVolunteerProfilePage(container),
  });
  const spinner = Loader.inline(container.querySelector('#profile-loading'), 'Loading profile…');

  try {
    await Promise.all([
      profileActions.fetchProfile(),
      profileActions.fetchPaymentProfile(),
    ]);
    spinner.remove();
  } catch (err) {
    spinner.remove();
    errDisplay.show(err);
    return;
  }

  const { profile, paymentProfile } = profileStore.state;

  section.innerHTML = `
    <header class="page-header">
      <h1>My Profile</h1>
    </header>

    <div class="vol-profile__layout">
      <!-- Profile card -->
      <div class="vol-profile__side">
        <div class="info-card">
          <div class="profile-avatar" aria-hidden="true">${esc(profile.display_name?.[0]?.toUpperCase())}</div>
          <h2 class="profile-name">${esc(profile.display_name)}</h2>
          <p class="profile-bio text-muted">${esc(profile.bio || 'No bio yet.')}</p>
          <div id="rep-container"></div>
        </div>
      </div>

      <!-- Edit forms -->
      <div class="vol-profile__main">
        <div id="edit-error"></div>

        <form id="profile-form" novalidate aria-label="Edit volunteer profile">
          <fieldset class="form-section">
            <legend>Public profile</legend>
            <div class="form-field">
              <label for="field-display-name">Display name <span aria-hidden="true">*</span></label>
              <input id="field-display-name" name="display_name" type="text" required
                     value="${esc(profile.display_name)}" />
              <div class="form-field__error" role="alert" aria-live="polite"></div>
            </div>
            <div class="form-field">
              <label for="field-bio">Bio</label>
              <textarea id="field-bio" name="bio" rows="4">${esc(profile.bio || '')}</textarea>
            </div>
          </fieldset>
          <div class="form-actions">
            <button type="submit" class="btn btn--primary" id="profile-save-btn">
              <span class="btn__label">Save profile</span>
              <span class="btn__loading" hidden>Saving…</span>
            </button>
          </div>
        </form>

        <form id="payment-form" novalidate aria-label="Payment profile form" style="margin-top: var(--space-8);">
          <fieldset class="form-section">
            <legend>Payment profile</legend>
            <div class="form-field">
              <label for="field-ln-address">Lightning address</label>
              <input id="field-ln-address" name="lightning_address" type="text"
                     placeholder="you@wallet.example"
                     value="${esc(paymentProfile?.lightning_address || '')}" />
              <p class="form-field__hint">Used to receive Lightning payments for completed tasks.</p>
            </div>
            <div class="form-field">
              <label for="field-default-invoice">Default BOLT11 invoice (fallback)</label>
              <textarea id="field-default-invoice" name="default_invoice" rows="3"
                        class="mono" placeholder="lnbc…">${esc(paymentProfile?.default_invoice || '')}</textarea>
            </div>
          </fieldset>
          <div class="form-actions">
            <button type="submit" class="btn btn--primary" id="payment-save-btn">
              <span class="btn__label">Save payment profile</span>
              <span class="btn__loading" hidden>Saving…</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  `;

  // Rep badge
  const repBadge = new ReputationBadge({ score: profile.reputation_score });
  section.querySelector('#rep-container').appendChild(repBadge.element);

  const editErrDisplay = new APIErrorDisplay(section.querySelector('#edit-error'));

  // Profile form
  section.querySelector('#profile-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    editErrDisplay.clear();
    const data    = new FormData(e.target);
    const nameVal = data.get('display_name').trim();
    if (!nameVal) {
      e.target.querySelector('.form-field__error').textContent = 'Display name is required.';
      return;
    }
    const btn = section.querySelector('#profile-save-btn');
    setLoading(btn, true);
    try {
      await profileActions.updateProfile({ display_name: nameVal, bio: data.get('bio').trim() });
      Toast.show({ message: 'Profile updated.', type: 'success' });
    } catch (err) {
      editErrDisplay.show(err);
    } finally {
      setLoading(btn, false);
    }
  });

  // Payment form
  section.querySelector('#payment-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    editErrDisplay.clear();
    const data = new FormData(e.target);
    const btn  = section.querySelector('#payment-save-btn');
    setLoading(btn, true);
    try {
      await profileActions.savePaymentProfile({
        lightning_address: data.get('lightning_address').trim(),
        default_invoice:   data.get('default_invoice').trim(),
      });
      Toast.show({ message: 'Payment profile saved.', type: 'success' });
    } catch (err) {
      editErrDisplay.show(err);
    } finally {
      setLoading(btn, false);
    }
  });
}

function setLoading(btn, loading) {
  btn.disabled = loading;
  btn.querySelector('.btn__label').hidden  = loading;
  btn.querySelector('.btn__loading').hidden = !loading;
}

function esc(str) {
  const d = document.createElement('div');
  d.textContent = str || '';
  return d.innerHTML;
}
