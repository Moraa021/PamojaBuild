import taskApi      from '../api/taskApi.js';
import trusteeApi   from '../api/trusteeApi.js';
import lightningApi from '../api/lightningApi.js';

const taskService = {
  canAcceptDonations(task) {
    return ['ACTIVE', 'LIQUIDATING'].includes(task.financial_state);
  },

  isReadyForPayout(task) {
    return task.financial_state === 'READY_FOR_PAYOUT';
  },

  async fetchTaskAndRequestInvoice(slug, amountSats) {
    const task = await taskApi.getBySlug(slug);
    if (!this.canAcceptDonations(task)) {
      throw new Error(`Campaign "${task.title}" is not currently accepting donations (state: ${task.financial_state}).`);
    }
    const invoice = await lightningApi.requestInvoice(slug, { amount_sats: amountSats });
    return { task, invoice };
  },

  describeState(task) {
    const map = {
      open:                 'This campaign is open and accepting volunteer applications.',
      in_progress:          'Volunteers are currently working on this campaign.',
      pending_verification: 'Work has been submitted and is awaiting trustee verification.',
      completed:            'This campaign has been completed and funds have been disbursed.',
    };
    return map[task.status] || task.status;
  },
};

export default taskService;
