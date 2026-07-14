import { Store } from './store.js';
import escrowApi from '../api/escrowApi.js';

const initialState = {
  manifest:          null,    // PayoutReviewResponse | null
  thresholdReached:  false,
  submittedSigs:     [],      // array of submitted trustee public key hexes
  loading:           false,
  error:             null,
};

const payoutStore = new Store(initialState);

export const payoutActions = {
  async fetchManifest(slug) {
    payoutStore.setState({ loading: true, error: null });
    try {
      const manifest = await escrowApi.getPayoutManifest(slug);
      payoutStore.setState({ manifest, loading: false });
      return manifest;
    } catch (err) {
      payoutStore.setState({ error: err.message, loading: false });
      throw err;
    }
  },

  async submitSignature(slug, payload) {
    payoutStore.setState({ loading: true, error: null });
    try {
      const result = await escrowApi.submitCoSignatures(slug, payload);
      const submittedSigs = [
        ...payoutStore.state.submittedSigs,
        payload.trustee_public_key_hex,
      ];
      payoutStore.setState({
        thresholdReached: result.threshold_reached || false,
        submittedSigs,
        loading: false,
      });
      return result;
    } catch (err) {
      payoutStore.setState({ error: err.message, loading: false });
      throw err;
    }
  },

  clearPayout() {
    payoutStore.setState({
      manifest:         null,
      thresholdReached: false,
      submittedSigs:    [],
      error:            null,
    });
  },
};

export default payoutStore;
