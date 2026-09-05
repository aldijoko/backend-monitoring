package repository

import (
	"gorm.io/gorm"

	"monitoring-cctv-be/internal/model"
)

type CameraRepository struct {
	db *gorm.DB
}

func NewCameraRepository(db *gorm.DB) *CameraRepository {
	return &CameraRepository{db: db}
}

type CameraFilters struct {
	EdgeID *uint
	Status string
	Q      string
	// Limit <= 0 means unlimited — /live and other internal callers rely on
	// getting every matching camera back, so pagination stays opt-in.
	Limit  int
	Offset int
}

func (r *CameraRepository) List(f CameraFilters) ([]model.Camera, int64, error) {
	q := r.db.Model(&model.Camera{})

	if f.EdgeID != nil {
		q = q.Where("edge_id = ?", *f.EdgeID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Q != "" {
		q = q.Where("name ILIKE ?", "%"+f.Q+"%")
	}

	var total int64
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if f.Limit > 0 {
		q = q.Limit(f.Limit).Offset(f.Offset)
	}

	var cameras []model.Camera
	if err := q.Order("id ASC").Find(&cameras).Error; err != nil {
		return nil, 0, err
	}
	return cameras, total, nil
}

func (r *CameraRepository) FindByID(id uint) (*model.Camera, error) {
	var c model.Camera
	if err := r.db.First(&c, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *CameraRepository) Create(c *model.Camera) error {
	return r.db.Create(c).Error
}

func (r *CameraRepository) Update(c *model.Camera) error {
	return r.db.Save(c).Error
}

func (r *CameraRepository) Delete(id uint) error {
	return r.db.Delete(&model.Camera{}, id).Error
}

func (r *CameraRepository) EdgeNamesByIDs(ids []uint) (map[uint]model.Edge, error) {
	if len(ids) == 0 {
		return map[uint]model.Edge{}, nil
	}
	var edges []model.Edge
	if err := r.db.Where("id IN ?", ids).Find(&edges).Error; err != nil {
		return nil, err
	}
	m := make(map[uint]model.Edge, len(edges))
	for _, e := range edges {
		m[e.ID] = e
	}
	return m, nil
}

func (r *CameraRepository) UpdateStatus(id uint, status model.CameraStatus, streamURL string, lastSeenAt *string) error {
	updates := map[string]any{"status": status}
	if streamURL != "" {
		updates["stream_url"] = streamURL
	}
	if lastSeenAt != nil {
		updates["last_seen_at"] = *lastSeenAt
	}
	return r.db.Model(&model.Camera{}).Where("id = ?", id).Updates(updates).Error
}
