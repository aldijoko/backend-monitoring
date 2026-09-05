package repository

import (
	"gorm.io/gorm"

	"monitoring-cctv-be/internal/model"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

type UserListParams struct {
	Search   string
	Role     string
	IsActive *bool
	Limit    int
	Offset   int
}

func (r *UserRepository) List(p UserListParams) ([]model.User, int64, error) {
	q := r.db.Model(&model.User{})

	if p.Search != "" {
		q = q.Where("username ILIKE ? OR email ILIKE ?", "%"+p.Search+"%", "%"+p.Search+"%")
	}
	if p.Role != "" {
		q = q.Where("role = ?", p.Role)
	}
	if p.IsActive != nil {
		q = q.Where("is_active = ?", *p.IsActive)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := p.Limit
	if limit <= 0 {
		limit = 10
	}

	users := []model.User{}
	if err := q.Order("created_at DESC").Limit(limit).Offset(p.Offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var u model.User
	if err := r.db.First(&u, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var u model.User
	if err := r.db.Where("username = ?", username).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Create(u *model.User) error {
	return r.db.Create(u).Error
}

func (r *UserRepository) Update(u *model.User) error {
	return r.db.Save(u).Error
}

func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

func (r *UserRepository) CountAll() (int64, error) {
	var total int64
	err := r.db.Model(&model.User{}).Count(&total).Error
	return total, err
}
