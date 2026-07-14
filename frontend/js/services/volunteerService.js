import volunteerApi from '../api/volunteerApi.js';

const volunteerService = {
  computeReputationStats({ profile, payments, applications }) {
    const completedPayments = payments.filter(p => p.status === 'completed');
    const approvedApps      = applications.filter(a => a.status === 'approved');

    return {
      reputationScore:   profile.reputation_score,
      completedTasks:    completedPayments.length,
      totalEarnedSats:   completedPayments.reduce((s, p) => s + (p.amount_sats || 0), 0),
      totalApplications: applications.length,
      approvedApps:      approvedApps.length,
      approvalRate:      applications.length
        ? Math.round((approvedApps.length / applications.length) * 100)
        : 0,
      pendingPayments:   payments.filter(p => p.status === 'pending').length,
    };
  },

  findApplication(applications, taskSlug) {
    return applications.find(a => a.task_slug === taskSlug) || null;
  },

  filterPayments(payments, status) {
    if (!status) return payments;
    return payments.filter(p => p.status === status);
  },
};

export default volunteerService;
