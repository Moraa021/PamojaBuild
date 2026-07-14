import { Store } from './store.js';
import trusteeApi from '../api/trusteeApi.js';

const initialState = {
  registeredKeys: [], // array of TrusteeKey responses
  loading:        false,
  error:          null,
};

const trusteeStore = new Store(initialState);

export const trusteeActions = {
  async registerKeys(slug, payload) {
    trusteeStore.setState({ loading: true, error: null });
    try {
      const result = await trusteeApi.registerKeys(slug, payload);
      trusteeStore.setState({
        registeredKeys: [...trusteeStore.state.registeredKeys, result],
        loading: false,
      });
      return result;
    } catch (err) {
      trusteeStore.setState({ error: err.message, loading: false });
      throw err;
    }
  },

  clearKeys() {
    trusteeStore.setState({ registeredKeys: [], error: null });
  },
};

export default trusteeStore;
