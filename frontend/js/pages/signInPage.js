import { ENV }           from '../config/env.js';
import { APIErrorDisplay } from '../components/APIErrorDisplay.js';
import { authActions }   from '../state/authStore.js';
import { Toast }         from '../components/Toast.js';
import { navigate }      from '../utils/utils.js';

const AUTH_SIGNIN_URL = `${ENV.API_BASE_URL}/api/${ENV.API_VERSION}/auth/signin`;

export function renderSignInPage(container) {
  container.innerHTML = `
    <section class="auth-page container">
      <div class="auth-card">
        <div class="auth-card__logo" aria-hidden="true">⚡</div>
        <h1 class="auth-card__title">Sign in</h1>
        <p class="auth-card__sub">Welcome back. Sign in to continue.</p>

        <div id="signin-error"></div>

        <form id="signin-form" novalidate aria-label="Sign in form">
          <div class="form-field">
            <label for="field-email">Email <span aria-hidden="true">*</span></label>
            <input id="field-email" name="email" type="email" required
                   autocomplete="email" placeholder="you@example.com" />
            <div class="form-field__error" role="alert" aria-live="polite"></div>
          </div>

          <div class="form-field">
            <label for="field-password">Password <span aria-hidden="true">*</span></label>
            <input id="field-password" name="password" type="password" required
                   autocomplete="current-password" placeholder="••••••••" />
            <div class="form-field__error" role="alert" aria-live="polite"></div>
          </div>

          <button type="submit" class="btn btn--primary btn--full" id="signin-btn">
            <span class="btn__label">Sign in</span>
            <span class="btn__loading" hidden>Signing in…</span>
          </button>
        </form>

        <p class="auth-card__footer">
          Don't have an account? <a href="/register">Register</a>
        </p>
      </div>
    </section>
  `;

  const form       = container.querySelector('#signin-form');
  const submitBtn  = container.querySelector('#signin-btn');
  const errDisplay = new APIErrorDisplay(container.querySelector('#signin-error'));

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    errDisplay.clear();

    const data     = new FormData(form);
    const email    = data.get('email').trim();
    const password = data.get('password');

    let valid = true;
    if (!email) {
      setFieldError(form, 'email', 'Email is required.');
      valid = false;
    }
    if (!password) {
      setFieldError(form, 'password', 'Password is required.');
      valid = false;
    }
    if (!valid) return;

    setLoading(submitBtn, true);

    try {
      const res = await fetch(AUTH_SIGNIN_URL, {
        method:  'POST',
        headers: { 'Content-Type': 'application/json' },
        body:    JSON.stringify({ email, password }),
      });

      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        throw new Error(body.message || `Sign in failed (${res.status})`);
      }

      const { token, user_id, role } = await res.json();
      authActions.setSession({ token, userId: user_id, role });

      Toast.show({ message: 'Signed in successfully.', type: 'success' });

      // Redirect based on role
      const redirectMap = {
        volunteer: '/volunteer/dashboard',
        creator:   '/campaigns',
        trustee:   '/trustee/payout',
        admin:     '/admin',
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
