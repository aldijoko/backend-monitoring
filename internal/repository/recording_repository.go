package repository

import (
	"gorm.io/gorm"

	"monitoring-cctv-be/internal/model"
)

type RecordingRepository struct {
	db *gorm.DB
}

func NewRecordingRepository(db *gorm.DB) *RecordingRepository {
	return &RecordingRepository{db: db}
}

type RecordingListParams struct {
	Search   string
	EdgeID   *uint
	CameraID *uint
	DateFrom *string
	DateTo   *string
	Limit    int
	Offset   int
	Sort     string
	Order    string
}

func (r *RecordingRepository) List(p RecordingListParams) ([]model.Recording, int64, error) {
	q := r.db.Model(&model.Recording{})

	if p.Search != "" {
		like := "%" + p.Search + "%"
		q = q.Where("filename ILIKE ? OR edge_code ILIKE ? OR camera_name ILIKE ?", like, like, like)
	}
	if p.EdgeID != nil {
		q = q.Where("edge_id = ?", *p.EdgeID)
	}
	if p.CameraID != nil {
		q = q.Where("camera_id = ?", *p.CameraID)
	}
	if p.DateFrom != nil {
		q = q.Where("started_at >= ?", *p.DateFrom)
	}
	if p.DateTo != nil {
		q = q.Where("started_at <= ?", *p.DateTo)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sort := p.Sort
	if sort == "" {
		sort = "started_at"
	}
	order := p.Order
	if order != "asc" {
		order = "desc"
	}
	limit := p.Limit
	if limit <= 0 {
		limit = 10
	}

	items := []model.Recording{}
	if err := q.Order(sort + " " + order).Limit(limit).Offset(p.Offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *RecordingRepository) FindByID(id uint) (*model.Recording, error) {
	var rec model.Recording
	if err := r.db.First(&rec, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *RecordingRepository) FindByIDs(ids []uint) ([]model.Recording, error) {
	var items []model.Recording
	if err := r.db.Where("id IN ?", ids).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *RecordingRepository) Create(rec *model.Recording) error {
	return r.db.Create(rec).Error
}

type CameraOption struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	EdgeID   uint   `json:"edge_id"`
	EdgeCode string `json:"edge_code"`
}

func (r *RecordingRepository) DistinctCameras() ([]CameraOption, error) {
	opts := []CameraOption{}
	err := r.db.Model(&model.Recording{}).
		Distinct("camera_id as id, camera_name as name, edge_id, edge_code").
		Order("camera_name ASC").
		Scan(&opts).Error
	return opts, err
}
