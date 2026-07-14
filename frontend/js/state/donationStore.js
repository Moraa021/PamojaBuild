import { Store } from './store.js';
import lightningApi from '../api/lightningApi.js';

const initialState = {
  invoice: null,   // DonationInvoiceResponse | null
  loading: false,
  error:   null,
};

const donationStore = new Store(initialState);

export const donationActions = {
  async requestInvoice(slug, amountSats) {
    donationStore.setState({ loading: true, error: null, invoice: null });
    try {
      const invoice = await lightningApi.requestInvoice(slug, { amount_sats: amountSats });
      donationStore.setState({ invoice, loading: false });
      return invoice;
    } catch (err) {
      donationStore.setState({ error: err.message, loading: false });
      throw err;
    }
  },

  clearInvoice() {
    donationStore.setState({ invoice: null, error: null });
  },
};

export default donationStore;
