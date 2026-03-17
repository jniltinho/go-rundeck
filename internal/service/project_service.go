package service

import (
	"errors"
	"go-rundeck/internal/model"
	"go-rundeck/internal/repository"
)

// ProjectService provides business logic for projects.
type ProjectService struct {
	repo *repository.ProjectRepository
}

// NewProjectService creates a new ProjectService.
func NewProjectService(repo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

// Create validates and persists a new project.
func (s *ProjectService) Create(name, description, tags string, createdBy uint) (*model.Project, error) {
	if name == "" {
		return nil, errors.New("project name is required")
	}
	p := &model.Project{
		Name:        name,
		Description: description,
		Tags:        tags,
		Active:      true,
		CreatedBy:   createdBy,
	}
	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

// GetByID returns a project by ID.
func (s *ProjectService) GetByID(id uint) (*model.Project, error) {
	return s.repo.GetByID(id)
}

// List returns all active projects.
func (s *ProjectService) List() ([]model.Project, error) {
	return s.repo.List()
}

// Update applies changes to a project.
func (s *ProjectService) Update(id uint, name, description, tags string) (*model.Project, error) {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		p.Name = name
	}
	p.Description = description
	p.Tags = tags
	if err := s.repo.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

// Delete soft-deletes a project.
func (s *ProjectService) Delete(id uint) error {
	return s.repo.Delete(id)
}

// Count returns the number of active projects.
func (s *ProjectService) Count() (int64, error) {
	return s.repo.Count()
}
