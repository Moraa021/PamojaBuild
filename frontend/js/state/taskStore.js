import { Store } from './store.js';
import taskApi from '../api/taskApi.js';

const initialState = {
  currentTask: null,   // TaskResponse | null
  loading:     false,
  error:       null,   // string | null
};

const taskStore = new Store(initialState);

export const taskActions = {
  async fetchTask(slug) {
    taskStore.setState({ loading: true, error: null });
    try {
      const task = await taskApi.getBySlug(slug);
      taskStore.setState({ currentTask: task, loading: false });
      return task;
    } catch (err) {
      taskStore.setState({ error: err.message, loading: false });
      throw err;
    }
  },

  async createTask(payload) {
    taskStore.setState({ loading: true, error: null });
    try {
      const task = await taskApi.create(payload);
      taskStore.setState({ currentTask: task, loading: false });
      return task;
    } catch (err) {
      taskStore.setState({ error: err.message, loading: false });
      throw err;
    }
  },

  clearTask() {
    taskStore.setState({ currentTask: null, error: null });
  },
};

export default taskStore;
