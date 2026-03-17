package repository

import (
	"go-rundeck/internal/model"

	"gorm.io/gorm"
)

// NodeRepository handles persistence for Node entities.
type NodeRepository struct {
	db *gorm.DB
}

// NewNodeRepository creates a new NodeRepository.
func NewNodeRepository(db *gorm.DB) *NodeRepository {
	return &NodeRepository{db: db}
}

// Create inserts a new node.
func (r *NodeRepository) Create(n *model.Node) error {
	return r.db.Create(n).Error
}

// GetByID retrieves a node by its primary key.
func (r *NodeRepository) GetByID(id uint) (*model.Node, error) {
	var n model.Node
	err := r.db.Preload("Project").First(&n, id).Error
	return &n, err
}

// ListByProject returns all active nodes belonging to a project.
func (r *NodeRepository) ListByProject(projectID uint) ([]model.Node, error) {
	var nodes []model.Node
	err := r.db.Where("project_id = ? AND active = ?", projectID, true).
		Order("name asc").Find(&nodes).Error
	return nodes, err
}

// Update saves changes to an existing node.
func (r *NodeRepository) Update(n *model.Node) error {
	return r.db.Save(n).Error
}

// Delete soft-deletes a node.
func (r *NodeRepository) Delete(id uint) error {
	return r.db.Delete(&model.Node{}, id).Error
}

// CountByProject returns the total number of active nodes for a project.
func (r *NodeRepository) CountByProject(projectID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Node{}).
		Where("project_id = ? AND active = ?", projectID, true).Count(&count).Error
	return count, err
}

// FindByTags returns nodes whose Tags field contains all provided tags.
func (r *NodeRepository) FindByTags(projectID uint, tags []string) ([]model.Node, error) {
	query := r.db.Where("project_id = ? AND active = ?", projectID, true)
	for _, tag := range tags {
		query = query.Where("tags LIKE ?", "%"+tag+"%")
	}
	var nodes []model.Node
	err := query.Find(&nodes).Error
	return nodes, err
}
