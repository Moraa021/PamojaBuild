package volunteers

import (
	"errors"

	"PamojaBuild/internal/models"
	tasksRepo "PamojaBuild/internal/tasks"
)

type Service struct {
	Repo      *Repository
	TasksRepo *tasksRepo.Repository
}

func NewService(repo *Repository, tasksRepo *tasksRepo.Repository) *Service {
	return &Service{Repo: repo, TasksRepo: tasksRepo}
}

func (s *Service) Apply(taskID, userID int64) (*models.Volunteer, error) {
	task, err := s.TasksRepo.GetByID(taskID)
	if err != nil {
		return nil, errors.New("task not found")
	}
	if task.Status != "open" {
		return nil, errors.New("task is not open for volunteers")
	}
	if task.CreatorID == userID {
		return nil, errors.New("task poster cannot volunteer for their own task")
	}
	already, err := s.Repo.AlreadyApplied(taskID, userID)
	if err != nil {
		return nil, err
	}
	if already {
		return nil, errors.New("you have already applied to this task")
	}
	count, err := s.Repo.CountForTask(taskID)
	if err != nil {
		return nil, err
	}
	if count >= task.MaxVolunteers {
		return nil, errors.New("volunteer cap reached")
	}
	status := "pending"
	if task.VolunteerMode == "open" {
		status = "approved"
	}
	v := &models.Volunteer{TaskID: taskID, UserID: userID, Status: status}
	if err := s.Repo.Create(v); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *Service) Approve(volunteerID, requesterID int64) error {
	v, err := s.Repo.GetByID(volunteerID)
	if err != nil {
		return errors.New("volunteer not found")
	}
	task, err := s.TasksRepo.GetByID(v.TaskID)
	if err != nil {
		return err
	}
	if task.CreatorID != requesterID {
		return errors.New("only the task poster can approve volunteers")
	}
	if v.Status != "pending" {
		return errors.New("volunteer is not in pending status")
	}
	count, err := s.Repo.CountForTask(v.TaskID)
	if err != nil {
		return err
	}
	if count >= task.MaxVolunteers {
		return errors.New("volunteer cap reached")
	}
	return s.Repo.UpdateStatus(volunteerID, "approved")
}

func (s *Service) Reject(volunteerID, requesterID int64) error {
	v, err := s.Repo.GetByID(volunteerID)
	if err != nil {
		return errors.New("volunteer not found")
	}
	task, err := s.TasksRepo.GetByID(v.TaskID)
	if err != nil {
		return err
	}
	if task.CreatorID != requesterID {
		return errors.New("only the task poster can reject volunteers")
	}
	return s.Repo.UpdateStatus(volunteerID, "rejected")
}

func (s *Service) ListForTask(taskID int64) ([]models.Volunteer, error) {
	return s.Repo.ListForTask(taskID)
}
