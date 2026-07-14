import { taskActions } from '../state/taskStore.js';
import { donationActions } from '../state/donationStore.js';
import { QRCodeViewer } from '../components/QRCodeviewer.js';
import { APIErrorDisplay } from '../components/APIErrorDisplay.js';
import { Loader } from '../components/Loader.js';
import { formatSats, getPathParam, navigate } from '../utils/utils.js';

export async function renderDonationPage(container) {
  const slug = getPathParam(1);
  if (!slug) {
    container.innerHTML = '<p>Invalid campaign URL.</p>';
    return;
  }

  container.innerHTML = `
    <section class="donation-page container">
      <div id="donate-error"></div>
      <div id="task-loading"></div>
    </section>
  `;

  const section    = container.querySelector('.donation-page');
  const errDisplay = new APIErrorDisplay(container.querySelector('#donate-error'), {
    onRetry: () => renderDonationPage(container),
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
      <button class="btn btn--ghost btn--sm" id="back-btn">← Back to campaign</button>
      <h1>Donate via Lightning</h1>
      <p class="page-header__sub">${esc(task.title)}</p>
    </header>

    <div id="donate-api-error"></div>

    <div class="donation-page__layout">
      <div class="donation-page__form-side">
        <form id="donate-form" novalidate aria-label="Donation form">
          <div class="form-field">
            <label for="field-amount">Amount (satoshis) <span aria-hidden="true">*</span></label>
            <input id="field-amount" name="amount_sats" type="number"
                   min="1" required placeholder="e.g. 10000"
                   aria-describedby="amount-hint" />
            ${task.goal_sats ? `<p id="amount-hint" class="form-field__hint">Campaign goal: ${formatSats(task.goal_sats)}</p>` : ''}
            <div class="form-field__error" role="alert" aria-live="polite"></div>
          </div>
          <button type="submit" class="btn btn--primary btn--full" id="invoice-btn">
            <span class="btn__label">⚡ Generate Invoice</span>
            <span class="btn__loading" hidden>Generating…</span>
          </button>
        </form>
      </div>

      <div class="donation-page__qr-side" id="qr-container">
        <p class="donation-page__qr-placeholder" aria-hidden="true">
          Enter an amount and generate an invoice to see the QR code here.
        </p>
      </div>
    </div>
  `;

  container.querySelector('#back-btn').addEventListener('click', () => navigate(`/tasks/${slug}`));

  const form      = section.querySelector('#donate-form');
  const invoiceBtn = section.querySelector('#invoice-btn');
  const qrContainer = section.querySelector('#qr-container');
  const formErrDisplay = new APIErrorDisplay(section.querySelector('#donate-api-error'), {
    onRetry: () => form.dispatchEvent(new Event('submit')),
  });
  const qrViewer = new QRCodeViewer(qrContainer);

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    formErrDisplay.clear();

    const amountInput = form.querySelector('[name="amount_sats"]');
    const errEl = amountInput.closest('.form-field').querySelector('.form-field__error');
    const amount = Number(amountInput.value);

    if (!amount || amount < 1) {
      errEl.textContent = 'Enter a valid amount greater than 0.';
      amountInput.setAttribute('aria-invalid', 'true');
      return;
    }
    errEl.textContent = '';
    amountInput.removeAttribute('aria-invalid');

    setLoading(invoiceBtn, true);
    try {
      const invoice = await donationActions.requestInvoice(slug, amount);
      await qrViewer.render(invoice);
    } catch (err) {
      formErrDisplay.show(err);
    } finally {
      setLoading(invoiceBtn, false);
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
