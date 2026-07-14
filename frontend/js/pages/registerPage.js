import { ENV }             from '../config/env.js';
import { APIErrorDisplay } from '../components/APIErrorDisplay.js';
import { authActions }     from '../state/authStore.js';
import { Toast }           from '../components/Toast.js';
import { navigate, isValidEmail } from '../utils/utils.js';
import { ROLES }           from '../config/roles.js';

const AUTH_REGISTER_URL = `${ENV.API_BASE_URL}/api/${ENV.API_VERSION}/auth/register`;

export function renderRegisterPage(container) {
  container.innerHTML = `
    <section class="auth-page container">
      <div class="auth-card">
        <div class="auth-card__logo" aria-hidden="true">⚡</div>
        <h1 class="auth-card__title">Create account</h1>
        <p class="auth-card__sub">Join the platform as a volunteer, donor, campaign creator, or trustee.</p>

        <div id="register-error"></div>

        <form id="register-form" novalidate aria-label="Registration form">

          <div class="form-field">
            <label for="field-display-name">Display name <span aria-hidden="true">*</span></label>
            <input id="field-display-name" name="display_name" type="text" required
                   autocomplete="name" placeholder="Your public name" />
            <div class="form-field__error" role="alert" aria-live="polite"></div>
          </div>

          <div class="form-field">
            <label for="field-email">Email <span aria-hidden="true">*</span></label>
            <input id="field-email" name="email" type="email" required
                   autocomplete="email" placeholder="you@example.com" />
            <div class="form-field__error" role="alert" aria-live="polite"></div>
          </div>

          <div class="form-field">
            <label for="field-password">Password <span aria-hidden="true">*</span></label>
            <input id="field-password" name="password" type="password" required
                   autocomplete="new-password" placeholder="At least 8 characters" minlength="8" />
            <div class="form-field__error" role="alert" aria-live="polite"></div>
          </div>

          <div class="form-field">
            <label for="field-confirm-password">Confirm password <span aria-hidden="true">*</span></label>
            <input id="field-confirm-password" name="confirm_password" type="password" required
                   autocomplete="new-password" placeholder="Repeat your password" />
            <div class="form-field__error" role="alert" aria-live="polite"></div>
          </div>

          <div class="form-field">
            <label for="field-role">I am joining as <span aria-hidden="true">*</span></label>
            <select id="field-role" name="role" required>
              <option value="">Select your role</option>
              <option value="volunteer">Volunteer — I want to do work and earn Bitcoin</option>
              <option value="donor">Donor — I want to fund campaigns</option>
              <option value="creator">Campaign Creator — I want to create volunteer tasks</option>
              <option value="trustee">Trustee — I want to co-sign payouts</option>
            </select>
            <div class="form-field__error" role="alert" aria-live="polite"></div>
          </div>

          <button type="submit" class="btn btn--primary btn--full" id="register-btn">
            <span class="btn__label">Create account</span>
            <span class="btn__loading" hidden>Creating account…</span>
          </button>
        </form>

        <p class="auth-card__footer">
          Already have an account? <a href="/signin">Sign in</a>
        </p>
      </div>
    </section>
  `;

  const form       = container.querySelector('#register-form');
  const submitBtn  = container.querySelector('#register-btn');
  const errDisplay = new APIErrorDisplay(container.querySelector('#register-error'));

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    errDisplay.clear();

    const data            = new FormData(form);
    const display_name    = data.get('display_name').trim();
    const email           = data.get('email').trim();
    const password        = data.get('password');
    const confirmPassword = data.get('confirm_password');
    const role            = data.get('role');

    let valid = true;

    if (!display_name) {
      setFieldError(form, 'display_name', 'Display name is required.');
      valid = false;
    }
    if (!email || !isValidEmail(email)) {
      setFieldError(form, 'email', 'A valid email address is required.');
      valid = false;
    }
    if (!password || password.length < 8) {
      setFieldError(form, 'password', 'Password must be at least 8 characters.');
      valid = false;
    }
    if (password !== confirmPassword) {
      setFieldError(form, 'confirm_password', 'Passwords do not match.');
      valid = false;
    }
    if (!role) {
      setFieldError(form, 'role', 'Please select a role.');
      valid = false;
    }
    if (!valid) return;

    setLoading(submitBtn, true);

    try {
      const res = await fetch(AUTH_REGISTER_URL, {
        method:  'POST',
        headers: { 'Content-Type': 'application/json' },
        body:    JSON.stringify({ display_name, email, password, role }),
      });

      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || `Registration failed (${res.status})`);
      }

      const { token, user_id } = await res.json();
      authActions.setSession({ token, userId: user_id, role });

      Toast.show({ message: 'Account created! Welcome aboard.', type: 'success' });

      const redirectMap = {
        volunteer: '/volunteer/dashboard',
        creator:   '/campaigns/new',
        trustee:   '/trustee/register',
        donor:     '/',
      };
      navigate(redirectMap[role] || '/');

    } catch (err) {
      errDisplay.show(err);
      setLoading(submitBtn, false);
    }
  });
}

function setFieldError(form, name, msg) {
  const field = form.querySelector(`[name="${name}"]`);
  const errEl = field?.closest('.form-field')?.querySelector('.form-field__error');
  if (errEl) errEl.textContent = msg;
  field?.setAttribute('aria-invalid', 'true');
}

function setLoading(btn, loading) {
  btn.disabled = loading;
  btn.querySelector('.btn__label').hidden  = loading;
  btn.querySelector('.btn__loading').hidden = !loading;
}