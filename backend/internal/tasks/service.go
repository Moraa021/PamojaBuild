package tasks

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"PamojaBuild/internal/models"
)

type Service struct {
	Repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{Repo: repo}
}

var validCounties = map[string]bool{
	"Mombasa": true, "Kwale": true, "Kilifi": true, "Tana River": true, "Lamu": true,
	"Taita-Taveta": true, "Garissa": true, "Wajir": true, "Mandera": true, "Marsabit": true,
	"Isiolo": true, "Meru": true, "Tharaka-Nithi": true, "Embu": true, "Kitui": true,
	"Machakos": true, "Makueni": true, "Nyandarua": true, "Nyeri": true, "Kirinyaga": true,
	"Murang'a": true, "Kiambu": true, "Turkana": true, "West Pokot": true, "Samburu": true,
	"Trans Nzoia": true, "Uasin Gishu": true, "Elgeyo-Marakwet": true, "Nandi": true,
	"Baringo": true, "Laikipia": true, "Nakuru": true, "Narok": true, "Kajiado": true,
	"Kericho": true, "Bomet": true, "Kakamega": true, "Vihiga": true, "Bungoma": true,
	"Busia": true, "Siaya": true, "Kisumu": true, "Homa Bay": true, "Migori": true,
	"Kisii": true, "Nyamira": true, "Nairobi": true,
}

func (s *Service) CreateTask(t *models.Task, file multipart.File, header *multipart.FileHeader) error {
	if !validCounties[t.Region] {
		return errors.New("invalid region")
	}
	if t.VolunteerMode != "open" && t.VolunteerMode != "approval_required" {
		return errors.New("volunteer_mode must be open or approval_required")
	}
	if t.MaxVolunteers <= 0 {
		return errors.New("max_volunteers is required")
	}
	if file != nil {
		buf := make([]byte, 512)
		n, _ := file.Read(buf)
		ct := http.DetectContentType(buf[:n])
		file.Seek(0, io.SeekStart)
		if ct != "image/jpeg" && ct != "image/png" {
			return errors.New("only JPEG and PNG images are allowed")
		}
		uploadDir := "./static/uploads/tasks"
		os.MkdirAll(uploadDir, 0755)
		filename := fmt.Sprintf("%d_%s", t.CreatorID, filepath.Base(header.Filename))
		savePath := filepath.Join(uploadDir, filename)
		dst, err := os.Create(savePath)
		if err != nil {
			return err
		}
		defer dst.Close()
		if _, err := io.Copy(dst, file); err != nil {
			return err
		}
		relPath := "static/uploads/tasks/" + filename
		t.ImagePath = sql.NullString{String: relPath, Valid: true}
	}
	return s.Repo.Create(t)
}

func (s *Service) ListTasks(region, status, category string) ([]models.Task, error) {
	return s.Repo.List(region, status, category)
}

func (s *Service) GetTask(id int64) (*models.Task, error) {
	return s.Repo.GetByID(id)
}

func (s *Service) RaiseCap(taskID, newCap, requesterID int64) error {
	task, err := s.Repo.GetByID(taskID)
	if err != nil {
		return err
	}
	if task.CreatorID != requesterID {
		return errors.New("only the task poster can raise the cap")
	}
	if newCap < task.MaxVolunteers {
		return errors.New("cannot lower the volunteer cap")
	}
	return s.Repo.RaiseCap(taskID, newCap)
}
