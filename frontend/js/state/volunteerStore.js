import { Store } from './store.js';
import volunteerApi from '../api/volunteerApi.js';

const profileStore = new Store({
  profile:        null,   // Volunteer | null
  paymentProfile: null,   // VolunteerPaymentProfile | null
  loading:        false,
  error:          null,
});

export const profileActions = {
  async fetchProfile() {
    profileStore.setState({ loading: true, error: null });
    try {
      const profile = await volunteerApi.getProfile();
      profileStore.setState({ profile, loading: false });
      return profile;
    } catch (err) {
      profileStore.setState({ error: err.message, loading: false });
      throw err;
    }
  },

  async updateProfile(payload) {
    profileStore.setState({ loading: true, error: null });
    try {
      const profile = await volunteerApi.updateProfile(payload);
      profileStore.setState({ profile, loading: false });
      return profile;
    } catch (err) {
      profileStore.setState({ error: err.message, loading: false });
      throw err;
    }
  },

  async fetchPaymentProfile() {
    profileStore.setState({ loading: true, error: null });
    try {
      const paymentProfile = await volunteerApi.getPaymentProfile();
      profileStore.setState({ paymentProfile, loading: false });
      return paymentProfile;
    } catch (err) {
      profileStore.setState({ error: err.message, loading: false });
      throw err;
    }
  },

  async savePaymentProfile(payload) {
    profileStore.setState({ loading: true, error: null });
    try {
      const paymentProfile = await volunteerApi.savePaymentProfile(payload);
      profileStore.setState({ paymentProfile, loading: false });
      return paymentProfile;
    } catch (err) {
      profileStore.setState({ error: err.message, loading: false });
      throw err;
    }
  },
};

const applicationStore = new Store({
  applications: [],   // VolunteerApplication[]
  loading:      false,
  error:        null,
});

export const applicationActions = {
  async fetchApplications() {
    applicationStore.setState({ loading: true, error: null });
    try {
      const applications = await volunteerApi.getApplications();
      applicationStore.setState({ applications, loading: false });
      return applications;
    } catch (err) {
      applicationStore.setState({ error: err.message, loading: false });
      throw err;
    }
  },

  async applyToTask(slug, message) {
    applicationStore.setState({ loading: true, error: null });
    try {
      const application = await volunteerApi.applyToTask(slug, { message });
      applicationStore.setState({
        applications: [...applicationStore.state.applications, application],
        loading: false,
      });
      return application;
    } catch (err) {
      applicationStore.setState({ error: err.message, loading: false });
      throw err;
    }
  },
};

const submissionStore = new Store({
  submissions: [],   // WorkSubmission[]
  loading:     false,
  error:       null,
});

export const submissionActions = {
  async fetchSubmissions(slug) {
    submissionStore.setState({ loading: true, error: null });
    try {
      const submissions = await volunteerApi.getSubmissions(slug);
      submissionStore.setState({ submissions, loading: false });
      return submissions;
    } catch (err) {
      submissionStore.setState({ error: err.message, loading: false });
      throw err;
    }
  },

  async createSubmission(slug, payload) {
    submissionStore.setState({ loading: true, error: null });
    try {
      const submission = await volunteerApi.createSubmission(slug, payload);
      submissionStore.setState({
        submissions: [...submissionStore.state.submissions, submission],
        loading: false,
      });
      return submission;
    } catch (err) {
      submissionStore.setState({ error: err.message, loading: false });
      throw err;
    }
  },

  async completeTask(slug) {
    submissionStore.setState({ loading: true, error: null });
    try {
      const result = await volunteerApi.completeTask(slug);
      submissionStore.setState({ loading: false });
      return result;
    } catch (err) {
      submissionStore.setState({ error: err.message, loading: false });
      throw err;
    }
  },
};

const paymentStore = new Store({
  payments: [],   // VolunteerPayment[]
  loading:  false,
  error:    null,
});

export const paymentActions = {
  async fetchPayments() {
    paymentStore.setState({ loading: true, error: null });
    try {
      const payments = await volunteerApi.getPayments();
      paymentStore.setState({ payments, loading: false });
      return payments;
    } catch (err) {
      paymentStore.setState({ error: err.message, loading: false });
      throw err;
    }
  },
};

const taskBrowserStore = new Store({
  tasks:   [],    // TaskResponse[]
  filters: { category: '', region: '', status: 'open' },
  loading: false,
  error:   null,
});

export const taskBrowserActions = {
  async fetchTasks() {
    taskBrowserStore.setState({ loading: true, error: null });
    try {
      const tasks = await volunteerApi.listTasks();
      taskBrowserStore.setState({ tasks, loading: false });
      return tasks;
    } catch (err) {
      taskBrowserStore.setState({ error: err.message, loading: false });
      throw err;
    }
  },

  setFilter(partial) {
    taskBrowserStore.setState({
      filters: { ...taskBrowserStore.state.filters, ...partial },
    });
  },

  getFilteredTasks() {
    const { tasks, filters } = taskBrowserStore.state;
    return tasks.filter(t => {
      if (filters.category && t.category !== filters.category) return false;
      if (filters.region   && !t.region.toLowerCase().includes(filters.region.toLowerCase())) return false;
      if (filters.status   && t.status !== filters.status) return false;
      return true;
    });
  },
};

export { profileStore, applicationStore, submissionStore, paymentStore, taskBrowserStore };
