import { Store } from './store.js';
import { setTokenAccessor } from '../api/client.js';

const SESSION_KEY = 'vt_session';

function loadPersistedSession() {
  try {
    const raw = sessionStorage.getItem(SESSION_KEY);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

const persisted = loadPersistedSession();

const initialState = {
  token:   persisted?.token   || null,
  userId:  persisted?.userId  || null,
  loading: false,
  error:   null,
};

const authStore = new Store(initialState);

// Wire the token accessor so the API client always gets the current token
setTokenAccessor(() => authStore.state.token);

export const authActions = {
  setSession({ token, userId }) {
    const session = { token, userId };
    sessionStorage.setItem(SESSION_KEY, JSON.stringify(session));
    authStore.setState({ token, userId, error: null });
  },

  clearSession() {
    sessionStorage.removeItem(SESSION_KEY);
    authStore.setState({ token: null, userId: null, error: null });
  },
};

export default authStore;
