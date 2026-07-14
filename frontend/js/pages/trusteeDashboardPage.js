import { trusteeActions } from '../state/trusteeStore.js';
import authStore from '../state/authStore.js';
import { taskActions } from '../state/taskStore.js';
import { APIErrorDisplay } from '../components/APIErrorDisplay.js';
import { Loader } from '../components/Loader.js';
import { Toast } from '../components/Toast.js';
import { generateTrusteeKeyPair } from '../utils/crypto.js';
import { navigate, isValidXpub, isHex } from '../utils/utils.js';

export async function renderTrusteeDashboardPage(container) {
  const params = new URLSearchParams(window.location.search);
  const slug   = params.get('slug');

  if (!slug) {
    container.innerHTML = '<section class="container"><p>No campaign specified. Add <code>?slug=your-task-slug</code> to the URL.</p></section>';
    return;
  }

  container.innerHTML = `
    <section class="trustee-dashboard container">
      <div id="trustee-error"></div>
      <div id="task-loading"></div>
    </section>
  `;

  const section = container.querySelector('.trustee-dashboard');
  const errDisplay = new APIErrorDisplay(container.querySelector('#trustee-error'), {
    onRetry: () => renderTrusteeDashboardPage(container),
  });
  const spinner = Loader.inline(container.querySelector('#task-loading'), 'Loading campaign…');

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
      <h1>Trustee Registration</h1>
      <p class="page-header__sub">Campaign: <strong>${esc(task.title)}</strong></p>
    </header>

    <div class="trustee-dashboard__layout">
      <div class="trustee-dashboard__info">
        <div class="info-card">
          <h2>How trustee keys work</h2>
          <p>As a trustee, you provide two keys:</p>
          <ol class="trustee-info__list">
            <li><strong>xpub</strong> — Your BIP32 HD master public key. Used to derive the 3-of-5 multi-sig on-chain address for Layer 1 payouts.</li>
            <li><strong>WebCrypto key</strong> — A browser-generated ECDSA P-256 key. Used to authorise Layer 2 (Lightning) tail payouts without hardware wallets.</li>
          </ol>
          <p class="notice">Your private keys never leave your browser or device.</p>
        </div>
      </div>

      <div class="trustee-dashboard__form-side">
        <div id="reg-error"></div>
        <form id="trustee-reg-form" novalidate aria-label="Trustee key registration form">

          <div class="form-field">
            <label for="field-trustee-index">Your trustee slot <span aria-hidden="true">*</span></label>
            <select id="field-trustee-index" name="trustee_index" required>
              <option value="">Select your assigned slot</option>
              <option value="0">Slot 0</option>
              <option value="1">Slot 1</option>
              <option value="2">Slot 2</option>
              <option value="3">Slot 3</option>
              <option value="4">Slot 4</option>
            </select>
            <div class="form-field__error" role="alert" aria-live="polite"></div>
          </div>

          <div class="form-field">
            <label for="field-xpub">xpub (BIP32 HD Master Public Key) <span aria-hidden="true">*</span></label>
            <textarea id="field-xpub" name="xpub" required rows="3"
                      placeholder="xpub6CUGRUonZSQ4TWtTMmzXdrXDtypWKiKrhko4egpiMZbpiaQL2jkwSB1icqYh2cfDfVxdx4df189oLKnC5fSwqPfgyP3hooxujYzAu3fDVmz"
                      class="mono"></textarea>
            <div class="form-field__error" role="alert" aria-live="polite"></div>
          </div>

          <div class="form-field">
            <label>WebCrypto public key</label>
            <div class="key-gen-panel">
              <code id="pubkey-display" class="mono key-gen-panel__key" aria-live="polite">
                Not yet generated
              </code>
              <button type="button" class="btn btn--ghost" id="gen-key-btn">
                Generate browser key pair
              </button>
            </div>
            <input type="hidden" id="field-pubkey-hex" name="web_crypto_pubkey_hex" />
            <p class="form-field__hint">Click the button to generate a key pair in your browser. Only the public key is sent to the server.</p>
          </div>

          <div class="form-actions">
            <button type="submit" class="btn btn--primary" id="reg-submit-btn" disabled>
              <span class="btn__label">Register Keys</span>
              <span class="btn__loading" hidden>Registering…</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  `;

  const form       = section.querySelector('#trustee-reg-form');
  const submitBtn  = section.querySelector('#reg-submit-btn');
  const genKeyBtn  = section.querySelector('#gen-key-btn');
  const pubkeyDisp = section.querySelector('#pubkey-display');
  const pubkeyHidden = section.querySelector('#field-pubkey-hex');
  const regErrDisplay = new APIErrorDisplay(section.querySelector('#reg-error'));

  let storedKeyPair = null;

  genKeyBtn.addEventListener('click', async () => {
    genKeyBtn.disabled = true;
    genKeyBtn.textContent = 'Generating…';
    try {
      const { keyPair, pubKeyHex } = await generateTrusteeKeyPair();
      storedKeyPair = keyPair;
      pubkeyHidden.value  = pubKeyHex;
      pubkeyDisp.textContent = pubKeyHex;
      pubkeyDisp.setAttribute('title', pubKeyHex);
      submitBtn.disabled  = false;
      genKeyBtn.textContent = 'Regenerate key pair';
      Toast.show({ message: 'Key pair generated. Do not navigate away.', type: 'info' });
    } catch {
      Toast.show({ message: 'Key generation failed. Try again.', type: 'error' });
    } finally {
      genKeyBtn.disabled = false;
    }
  });

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    regErrDisplay.clear();

    if (!validateTrusteeForm(form)) return;

    const userId = authStore.state.userId;
    if (!userId) {
      Toast.show({ message: 'Sign in to register as a trustee.', type: 'error' });
      return;
    }

    const data = new FormData(form);
    const payload = {
      user_id:              Number(userId),
      trustee_index:        Number(data.get('trustee_index')),
      xpub:                 data.get('xpub').trim(),
      web_crypto_pubkey_hex: data.get('web_crypto_pubkey_hex').trim(),
    };

    setLoading(submitBtn, true);
    try {
      await trusteeActions.registerKeys(slug, payload);
      Toast.show({ message: 'Keys registered successfully.', type: 'success' });
      navigate(`/tasks/${slug}`);
    } catch (err) {
      regErrDisplay.show(err);
      setLoading(submitBtn, false);
    }
  });
}

function validateTrusteeForm(form) {
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

  const xpubField = form.querySelector('[name="xpub"]');
  if (xpubField && xpubField.value.trim() && !isValidXpub(xpubField.value.trim())) {
    const errEl = xpubField.closest('.form-field')?.querySelector('.form-field__error');
    if (errEl) errEl.textContent = 'Enter a valid BIP32 xpub key (starts with "xpub").';
    xpubField.setAttribute('aria-invalid', 'true');
    valid = false;
  }

  return valid;
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
