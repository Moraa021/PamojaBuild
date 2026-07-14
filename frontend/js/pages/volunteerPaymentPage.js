import { paymentActions, paymentStore } from '../state/volunteerStore.js';
import { PaymentHistoryTable } from '../components/PaymentHistoryTable.js';
import { PaymentSummaryCard } from '../components/PaymentSummaryCard.js';
import { APIErrorDisplay } from '../components/APIErrorDisplay.js';
import { Loader } from '../components/Loader.js';

const PAYMENT_STATUSES = ['pending', 'approved', 'signed', 'broadcast', 'completed'];

export async function renderVolunteerPaymentsPage(container) {
  container.innerHTML = `
    <section class="vol-payments container">
      <header class="page-header">
        <h1>Payment History</h1>
        <p class="page-header__sub">Track every payout from submission to your Lightning wallet.</p>
      </header>
      <div id="payments-error"></div>
      <div id="payments-loading"></div>
    </section>
  `;

  const section    = container.querySelector('.vol-payments');
  const errDisplay = new APIErrorDisplay(container.querySelector('#payments-error'), {
    onRetry: () => renderVolunteerPaymentsPage(container),
  });
  const spinner = Loader.inline(container.querySelector('#payments-loading'), 'Loading payment history…');

  try {
    await paymentActions.fetchPayments();
    spinner.remove();
  } catch (err) {
    spinner.remove();
    errDisplay.show(err);
    return;
  }

  const { payments } = paymentStore.state;

  // Status filter tabs
  const tabs = document.createElement('div');
  tabs.className = 'payment-tabs';
  tabs.setAttribute('role', 'tablist');
  tabs.setAttribute('aria-label', 'Filter payments by status');

  let activeFilter = '';

  const renderTable = (filter) => {
    activeFilter = filter;
    const filtered = filter ? payments.filter(p => p.status === filter) : payments;
    const tableContainer = section.querySelector('#payment-table-container');
    tableContainer.innerHTML = '';
    new PaymentHistoryTable(tableContainer, filtered);
  };

  const allTab = buildTab('All', '', true);
  tabs.appendChild(allTab);
  PAYMENT_STATUSES.forEach(s => tabs.appendChild(buildTab(cap(s), s, false)));

  tabs.addEventListener('click', (e) => {
    const btn = e.target.closest('[role="tab"]');
    if (!btn) return;
    tabs.querySelectorAll('[role="tab"]').forEach(t => {
      t.setAttribute('aria-selected', 'false');
      t.classList.remove('payment-tab--active');
    });
    btn.setAttribute('aria-selected', 'true');
    btn.classList.add('payment-tab--active');
    renderTable(btn.dataset.filter);
  });

  section.innerHTML += `<div id="summary-container"></div>`;
  section.appendChild(tabs);
  section.innerHTML += `<div id="payment-table-container"></div>`;

  new PaymentSummaryCard(section.querySelector('#summary-container'), payments);
  renderTable('');
}

function buildTab(label, filter, active) {
  const btn = document.createElement('button');
  btn.className = `payment-tab ${active ? 'payment-tab--active' : ''}`;
  btn.setAttribute('role', 'tab');
  btn.setAttribute('aria-selected', String(active));
  btn.dataset.filter = filter;
  btn.textContent = label;
  return btn;
}

function cap(str) {
  return str.charAt(0).toUpperCase() + str.slice(1);
}
