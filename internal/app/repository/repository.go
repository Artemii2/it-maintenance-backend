package repository

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository — доступ к данным через ORM (GORM).
type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

// =====================
// МОДЕЛИ (ORM)
// =====================

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"size:50;not null;unique"`
	PasswordHash string `gorm:"not null"`
	IsModerator  bool   `gorm:"not null;default:false"`
	CreatedAt    time.Time
}

type Service struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"size:120;not null"`
	Description string `gorm:"not null"`

	Status   string  `gorm:"size:16;not null"` // active/deleted
	ImageURL string `gorm:"column:image_url"`
	VideoURL string `gorm:"column:video_url"`

	PadType      string `gorm:"size:40;not null"`
	BaseResource int    `gorm:"not null"`
	Price        int    `gorm:"not null"`
}

type Application struct {
	ID          uint      `gorm:"primaryKey"`
	Status      string    `gorm:"size:16;not null"` // draft/deleted/formed/completed/rejected
	CreatedAt   time.Time `gorm:"not null"`
	CreatedByID uint      `gorm:"not null"`
	CreatedBy   User      `gorm:"foreignKey:CreatedByID"`

	FormedAt     *time.Time
	CompletedAt  *time.Time
	ModeratorID  *uint
	Moderator    *User `gorm:"foreignKey:ModeratorID"`
	DrivingStyle string `gorm:"column:driving_style"`
	Mileage      int    `gorm:"column:mileage"`

	// расчётное поле (например, при завершении заявки)
	TotalPrice int `gorm:"column:total_price"`

	Items []ApplicationService `gorm:"foreignKey:ApplicationID"`
}

type ApplicationService struct {
	ApplicationID uint `gorm:"primaryKey;column:application_id"`
	ServiceID     uint `gorm:"primaryKey;column:service_id"`

	Quantity  int    `gorm:"not null;default:1"`
	Position  int    `gorm:"not null;default:1"`
	IsPrimary bool   `gorm:"not null;default:false"`
	Comment   string

	Service Service `gorm:"foreignKey:ServiceID"`
}

// =====================
// ORM-ОПЕРАЦИИ (4 контроллера через ORM используют эти методы)
// =====================

func (r *Repository) SearchServices(query string) ([]Service, error) {
	var services []Service
	q := r.DB.Model(&Service{}).Where("status = 'active'")
	if query != "" {
		like := fmt.Sprintf("%%%s%%", query)
		q = q.Where("title ILIKE ? OR description ILIKE ? OR pad_type ILIKE ?", like, like, like)
	}
	if err := q.Order("id ASC").Find(&services).Error; err != nil {
		return nil, err
	}
	return services, nil
}

func (r *Repository) GetService(id uint) (Service, error) {
	var s Service
	if err := r.DB.Where("id = ? AND status = 'active'", id).First(&s).Error; err != nil {
		return Service{}, err
	}
	return s, nil
}

func (r *Repository) GetDraftApplication(userID uint) (*Application, error) {
	var app Application
	err := r.DB.
		Preload("Items.Service").
		Where("created_by_id = ? AND status = 'draft'", userID).
		First(&app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *Repository) GetApplicationByID(appID uint, userID uint) (*Application, error) {
	var app Application
	err := r.DB.
		Preload("Items.Service").
		Where("id = ? AND created_by_id = ? AND status <> 'deleted'", appID, userID).
		First(&app).Error
	if err != nil {
		return nil, err
	}

	// расчётное поле: total_price = Σ (price * quantity) при завершении
	if app.Status == "completed" && app.TotalPrice == 0 && len(app.Items) > 0 {
		total := 0
		for _, it := range app.Items {
			total += it.Service.Price * it.Quantity
		}
		if err := r.DB.Model(&Application{}).Where("id = ?", app.ID).Update("total_price", total).Error; err == nil {
			app.TotalPrice = total
		}
	}
	return &app, nil
}

func (r *Repository) AddServiceToDraft(userID, serviceID uint, qty int) (*Application, error) {
	returnApp := &Application{}

	return returnApp, r.DB.Transaction(func(tx *gorm.DB) error {
		// проверяем услугу
		var svc Service
		if err := tx.Where("id = ? AND status = 'active'", serviceID).First(&svc).Error; err != nil {
			return err
		}

		// ищем черновик
		var draft Application
		err := tx.Where("created_by_id = ? AND status = 'draft'", userID).First(&draft).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			draft = Application{
				Status:      "draft",
				CreatedByID: userID,
				CreatedAt:   time.Now(),
			}
			if err := tx.Create(&draft).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		item := ApplicationService{
			ApplicationID: draft.ID,
			ServiceID:     serviceID,
			Quantity:      qty,
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "application_id"}, {Name: "service_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"quantity": gorm.Expr("application_services.quantity + EXCLUDED.quantity"),
			}),
		}).Create(&item).Error; err != nil {
			return err
		}

		// перезагружаем с items
		if err := tx.Preload("Items.Service").First(returnApp, draft.ID).Error; err != nil {
			return err
		}
		return nil
	})
}