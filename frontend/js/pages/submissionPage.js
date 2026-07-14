import { submissionActions, submissionStore } from '../state/volunteerStore.js';
import { taskActions } from '../state/taskStore.js';
import { SubmissionUploader } from '../components/SubmissionUploader.js';
import { EvidenceGallery } from '../components/EvidenceGallery.js';
import { VerificationTracker } from '../components/VerificationTracker.js';
import { APIErrorDisplay } from '../components/APIErrorDisplay.js';
import { Loader } from '../components/Loader.js';
import { Toast } from '../components/Toast.js';
import { Modal } from '../components/Modal.js';
import { getPathParam, navigate } from '../utils/utils.js';

export async function renderSubmissionPage(container) {
  const slug = getPathParam(2); // /volunteer/tasks/:slug/submit
  if (!slug) {
    container.innerHTML = '<section class="container"><p>Invalid task URL.</p></section>';
    return;
  }

  container.innerHTML = `
    <section class="submission-page container">
      <div id="sub-page-error"></div>
      <div id="sub-page-loading"></div>
    </section>
  `;

  const section    = container.querySelector('.submission-page');
  const errDisplay = new APIErrorDisplay(container.querySelector('#sub-page-error'), {
    onRetry: () => renderSubmissionPage(container),
  });
  const spinner = Loader.inline(container.querySelector('#sub-page-loading'), 'Loading submission portal…');

  let task;
  try {
    [task] = await Promise.all([
      taskActions.fetchTask(slug),
      submissionActions.fetchSubmissions(slug),
    ]);
    spinner.remove();
  } catch (err) {
    spinner.remove();
    errDisplay.show(err);
    return;
  }

  const { submissions } = submissionStore.state;

  section.innerHTML = `
    <header class="page-header">
      <button class="btn btn--ghost btn--sm" id="back-btn">← Back to task</button>
      <h1>Submit Work</h1>
      <p class="page-header__sub">${esc(task.title)}</p>
    </header>

    <div class="submission-page__layout">
      <div class="submission-page__form">
        <h2>Upload evidence</h2>
        <div id="uploader-container"></div>

        ${task.status === 'in_progress'
          ? `<div class="complete-section">
               <h2>Ready to finish?</h2>
               <p class="text-muted">Once all work is submitted and confirmed, mark the task as complete to begin the verification and payout process.</p>
               <button class="btn btn--primary" id="complete-btn">Mark task as complete</button>
             </div>`
          : ''
        }
      </div>

      <div class="submission-page__sidebar">
        <div id="tracker-container"></div>
        <h2>Previous submissions</h2>
        <div id="gallery-container"></div>
      </div>
    </div>
  `;

  section.querySelector('#back-btn').addEventListener('click', () => navigate(`/volunteer/tasks/${slug}`));

  // Verification tracker
  const tracker = new VerificationTracker(section.querySelector('#tracker-container'));
  tracker.update({
    taskStatus:          task.status,
    financialState:      task.financial_state,
    hasApplication:      true,
    applicationApproved: true,
    hasSubmission:       submissions.length > 0,
  });

  // Submission uploader
  new SubmissionUploader(section.querySelector('#uploader-container'), {
    onSubmit: async (payload) => {
      try {
        await submissionActions.createSubmission(slug, payload);
        Toast.show({ message: 'Submission uploaded successfully.', type: 'success' });
        // Refresh gallery
        await submissionActions.fetchSubmissions(slug);
        gallery.update(submissionStore.state.submissions);
        tracker.update({
          taskStatus: task.status,
          financialState: task.financial_state,
          hasApplication: true,
          applicationApproved: true,
          hasSubmission: submissionStore.state.submissions.length > 0,
        });
      } catch (err) {
        Toast.show({ message: `Upload failed: ${err.message}`, type: 'error' });
        throw err;
      }
    },
  });

  // Evidence gallery
  const galleryContainer = section.querySelector('#gallery-container');
  const gallery = new EvidenceGallery(galleryContainer, submissions);

  // Complete task button
  const completeBtn = section.querySelector('#complete-btn');
  if (completeBtn) {
    completeBtn.addEventListener('click', () => {
      const confirmContent = document.createElement('p');
      confirmContent.textContent = 'Marking the task as complete will trigger the verification stage. Trustees will be notified to review and sign the payout. This cannot be undone.';

      new Modal({
        title:        'Confirm task completion',
        content:      confirmContent,
        confirmLabel: 'Mark complete',
        onConfirm:    async () => {
          try {
            await submissionActions.completeTask(slug);
            Toast.show({ message: 'Task marked as complete. Verification has begun.', type: 'success' });
            navigate(`/volunteer/tasks/${slug}`);
          } catch (err) {
            Toast.show({ message: `Error: ${err.message}`, type: 'error' });
            throw err;
          }
        },
      }).open();
    });
  }
}

function esc(str) {
  const d = document.createElement('div');
  d.textContent = str || '';
  return d.innerHTML;
}
