package repository

import (
	"gorm.io/gorm"

	"monitoring-cctv-be/internal/model"
)

type EdgeRepository struct {
	db *gorm.DB
}

func NewEdgeRepository(db *gorm.DB) *EdgeRepository {
	return &EdgeRepository{db: db}
}

type EdgeListParams struct {
	Search string
	Status string
	Limit  int
	Offset int
	Sort   string
	Order  string
}

// EdgeWithStats decorates an Edge with the camera_count the frontend expects
// (Edge.camera_count?), computed from the cameras table rather than stored.
type EdgeWithStats struct {
	model.Edge
	CameraCount int64 `json:"camera_count"`
}

func (r *EdgeRepository) List(p EdgeListParams) ([]EdgeWithStats, int64, error) {
	q := r.db.Model(&model.Edge{})

	if p.Search != "" {
		like := "%" + p.Search + "%"
		q = q.Where("code ILIKE ? OR name ILIKE ? OR hostname ILIKE ? OR ip_address ILIKE ?", like, like, like, like)
	}
	if p.Status != "" {
		q = q.Where("status = ?", p.Status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sort := p.Sort
	if sort == "" {
		sort = "created_at"
	}
	order := p.Order
	if order != "asc" {
		order = "desc"
	}
	limit := p.Limit
	if limit <= 0 {
		limit = 10
	}

	var edges []model.Edge
	if err := q.Order(sort + " " + order).Limit(limit).Offset(p.Offset).Find(&edges).Error; err != nil {
		return nil, 0, err
	}

	result := make([]EdgeWithStats, 0, len(edges))
	for _, e := range edges {
		var count int64
		r.db.Model(&model.Camera{}).Where("edge_id = ?", e.ID).Count(&count)
		result = append(result, EdgeWithStats{Edge: e, CameraCount: count})
	}

	return result, total, nil
}

func (r *EdgeRepository) FindByID(id uint) (*model.Edge, error) {
	var e model.Edge
	if err := r.db.First(&e, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *EdgeRepository) FindByCode(code string) (*model.Edge, error) {
	var e model.Edge
	if err := r.db.Where("code = ?", code).First(&e).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *EdgeRepository) Create(e *model.Edge) error {
	return r.db.Create(e).Error
}

func (r *EdgeRepository) Update(e *model.Edge) error {
	return r.db.Save(e).Error
}

func (r *EdgeRepository) Delete(id uint) error {
	return r.db.Delete(&model.Edge{}, id).Error
}

// RecalculateStatus derives the edge's status/last_heartbeat from the
// aggregate status of its cameras — there is no physical device at the edge
// sending its own heartbeat in this single-site/centralized deployment.
func (r *EdgeRepository) RecalculateStatus(edgeID uint) error {
	var onlineCount int64
	r.db.Model(&model.Camera{}).
		Where("edge_id = ? AND status IN ?", edgeID, []string{"online", "recording"}).
		Count(&onlineCount)

	var totalCount int64
	r.db.Model(&model.Camera{}).Where("edge_id = ?", edgeID).Count(&totalCount)

	status := model.EdgeOffline
	if totalCount == 0 {
		status = model.EdgePending
	} else if onlineCount > 0 {
		status = model.EdgeOnline
	}

	var lastSeen *string
	row := r.db.Model(&model.Camera{}).
		Where("edge_id = ?", edgeID).
		Select("MAX(last_seen_at)").Row()
	_ = row.Scan(&lastSeen)

	updates := map[string]any{"status": status}
	if lastSeen != nil {
		updates["last_heartbeat"] = *lastSeen
	}

	return r.db.Model(&model.Edge{}).Where("id = ?", edgeID).Updates(updates).Error
}
