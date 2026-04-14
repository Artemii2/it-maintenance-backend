package repository

import (
	"errors"
	"fmt"
	"math"
	"strings"
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

// CreateUser регистрирует нового пользователя.
func (r *Repository) CreateUser(username, passwordHash string, isModerator bool) (*User, error) {
	u := &User{
		Username:     username,
		PasswordHash: passwordHash,
		IsModerator:  isModerator,
	}
	if err := r.DB.Create(u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

// GetUserByUsername — поиск пользователя по логину (для входа).
func (r *Repository) GetUserByUsername(username string) (*User, error) {
	var u User
	if err := r.DB.Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

type Service struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"size:120;not null"`
	Description string `gorm:"not null"`

	Status   string  `gorm:"size:16;not null"` // active/deleted
	ImageURL string `gorm:"column:image_url"`
	VideoURL string `gorm:"column:video_url"`

	PadType          string `gorm:"size:40;not null"`
	BaseResource     int    `gorm:"not null"`
	Price            int    `gorm:"not null"`
	DrivingStyleHint string `gorm:"column:driving_style_hint;size:120"`
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

	// вычисляемые поля только для отдачи в API (не хранятся в БД)
	ItemsCount            int     `gorm:"-"`
	DrivingStyleCoeff     float64 `gorm:"-"` // k для формулы износа по стилю вождения
	MinRemainingKm        float64 `gorm:"-"` // минимальный остаток по строкам заявки (итог «по теме»)
	MinRemainingPercent   int     `gorm:"-"`
}

type ApplicationService struct {
	ApplicationID uint `gorm:"primaryKey;column:application_id"`
	ServiceID     uint `gorm:"primaryKey;column:service_id"`

	Quantity  int    `gorm:"not null;default:1"`
	Position  int    `gorm:"not null;default:1"`
	IsPrimary bool   `gorm:"not null;default:false"`
	Comment   string

	Service Service `gorm:"foreignKey:ServiceID"`

	// расчётные поля для прогноза износа (не хранятся в БД)
	RemainingKM           float64 `gorm:"-"`
	RemainingPercent      int     `gorm:"-"`
	EffectiveResourceKm float64 `gorm:"-"` // base_resource * k
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

// FilterServices — список услуг для REST API с простыми фильтрами по полям.
// Фильтры опциональны; записи со статусом "deleted" не возвращаются.
func (r *Repository) FilterServices(title, padType, status string) ([]Service, error) {
	var services []Service

	q := r.DB.Model(&Service{})

	// по умолчанию исключаем удалённые
	if status == "" {
		q = q.Where("status <> ?", "deleted")
	} else {
		q = q.Where("status = ?", status)
	}

	if title != "" {
		like := fmt.Sprintf("%%%s%%", title)
		q = q.Where("title ILIKE ? OR description ILIKE ?", like, like)
	}
	if padType != "" {
		q = q.Where("pad_type = ?", padType)
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

// CreateService — создание новой услуги.
func (r *Repository) CreateService(s *Service) error {
	if s.Status == "" {
		s.Status = "active"
	}
	return r.DB.Create(s).Error
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

// GetCartInfo — черновая заявка пользователя + количество услуг в ней.
func (r *Repository) GetCartInfo(userID uint) (*Application, int, error) {
	app, err := r.GetDraftApplication(userID)
	if err != nil {
		return nil, 0, err
	}
	if app == nil {
		return nil, 0, nil
	}
	count := 0
	for _, it := range app.Items {
		count += it.Quantity
	}
	return app, count, nil
}

// FilterApplicationsForUser — список заявок пользователя (кроме deleted и draft)
// с фильтрацией по статусу и диапазону даты формирования.
func (r *Repository) FilterApplicationsForUser(
	userID uint,
	status string,
	formedFrom, formedTo *time.Time,
) ([]Application, error) {
	var apps []Application

	q := r.DB.Model(&Application{}).
		Preload("Items.Service").
		Where("created_by_id = ? AND status NOT IN ('deleted', 'draft')", userID)

	if status != "" {
		q = q.Where("status = ?", status)
	}
	if formedFrom != nil {
		q = q.Where("formed_at >= ?", *formedFrom)
	}
	if formedTo != nil {
		q = q.Where("formed_at <= ?", *formedTo)
	}

	if err := q.Order("formed_at DESC, id DESC").Find(&apps).Error; err != nil {
		return nil, err
	}

	for i := range apps {
		r.applyWearAndTotalPrice(&apps[i])
		apps[i].ItemsCount = len(apps[i].Items)
	}
	return apps, nil
}

// CountVisibleApplicationsByStatusForUser — количество заявок пользователя по статусам
// (исключая deleted и draft), без учёта фильтров списка.
func (r *Repository) CountVisibleApplicationsByStatusForUser(userID uint) (map[string]int, int, error) {
	type row struct {
		Status string
		Count  int
	}

	var rows []row
	if err := r.DB.Model(&Application{}).
		Select("status, COUNT(*) AS count").
		Where("created_by_id = ? AND status NOT IN ('deleted', 'draft')", userID).
		Group("status").
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	out := map[string]int{
		"formed":    0,
		"completed": 0,
		"rejected":  0,
	}
	total := 0
	for _, r := range rows {
		out[r.Status] = r.Count
		total += r.Count
	}
	return out, total, nil
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

	r.applyWearAndTotalPrice(&app)
	return &app, nil
}

// applyWearAndTotalPrice — прогноз износа по теме «brake pad wear» и сумма по услугам.
// Формула на строку: effective = base_resource * k(стиль); остаток км = max(0, effective − пробег);
// остаток % = round(остаток / effective * 100).
func (r *Repository) applyWearAndTotalPrice(app *Application) {
	coef := drivingStyleCoef(app.DrivingStyle)
	app.DrivingStyleCoeff = coef
	mileage := float64(app.Mileage)

	total := 0
	for i := range app.Items {
		base := float64(app.Items[i].Service.BaseResource)
		if base > 0 {
			initial := base * coef
			app.Items[i].EffectiveResourceKm = initial
			remaining := initial - mileage
			if remaining < 0 {
				remaining = 0
			}
			app.Items[i].RemainingKM = remaining

			if initial > 0 {
				percent := (remaining / initial) * 100
				if percent < 0 {
					percent = 0
				}
				app.Items[i].RemainingPercent = int(math.Round(percent))
			}
		}

		total += app.Items[i].Service.Price * app.Items[i].Quantity
	}
	app.TotalPrice = total

	minKm := math.MaxFloat64
	minPct := 101
	for _, it := range app.Items {
		if it.RemainingKM < minKm {
			minKm = it.RemainingKM
		}
		if it.RemainingPercent < minPct {
			minPct = it.RemainingPercent
		}
	}
	if minKm == math.MaxFloat64 {
		minKm = 0
	}
	if minPct == 101 {
		minPct = 0
	}
	app.MinRemainingKm = minKm
	app.MinRemainingPercent = minPct

	if app.Status == "completed" {
		_ = r.DB.Model(&Application{}).Where("id = ?", app.ID).Update("total_price", total).Error
	}
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

// UpdateApplicationWearParams обновляет параметры расчёта износа у заявки пользователя.
// Вызывается при автоматической отправке формы на странице заявки (без кнопки «Рассчитать»).
func (r *Repository) UpdateApplicationWearParams(appID, userID uint, drivingStyle string, mileage int) error {
	updates := map[string]any{
		"driving_style": drivingStyle,
		"mileage":       mileage,
	}

	return r.DB.Model(&Application{}).
		Where("id = ? AND created_by_id = ? AND status <> 'deleted'", appID, userID).
		Updates(updates).Error
}

// DeleteApplicationItem удаляет одну услугу из заявки (м-м).
func (r *Repository) DeleteApplicationItem(appID, serviceID, userID uint) error {
	// проверяем, что заявка принадлежит пользователю
	var app Application
	if err := r.DB.Where("id = ? AND created_by_id = ?", appID, userID).First(&app).Error; err != nil {
		return err
	}
	return r.DB.Where("application_id = ? AND service_id = ?", appID, serviceID).
		Delete(&ApplicationService{}).Error
}

// UpdateApplicationItem обновляет количество/позицию/флаг is_primary в м-м.
// Поля обновляются только если указаны (qty/pos/isPrimary не nil).
func (r *Repository) UpdateApplicationItem(
	appID, serviceID, userID uint,
	qty, pos *int,
	isPrimary *bool,
) error {
	// проверяем, что заявка принадлежит пользователю
	var app Application
	if err := r.DB.Where("id = ? AND created_by_id = ?", appID, userID).First(&app).Error; err != nil {
		return err
	}

	updates := map[string]any{}
	if qty != nil {
		updates["quantity"] = *qty
	}
	if pos != nil {
		updates["position"] = *pos
	}
	if isPrimary != nil {
		updates["is_primary"] = *isPrimary
	}
	if len(updates) == 0 {
		return nil
	}

	return r.DB.Model(&ApplicationService{}).
		Where("application_id = ? AND service_id = ?", appID, serviceID).
		Updates(updates).Error
}

// UpdateApplicationForCreator обновляет поля заявки, которую меняет создатель.
// Используется, в том числе, для формирования заявки (смена статуса на formed).
func (r *Repository) UpdateApplicationForCreator(
	appID, userID uint,
	fields map[string]any,
) (*Application, error) {
	var app Application
	if err := r.DB.Where("id = ? AND created_by_id = ?", appID, userID).First(&app).Error; err != nil {
		return nil, err
	}

	// обрабатываем статус отдельно, чтобы выставить FormedAt
	if statusRaw, ok := fields["status"]; ok {
		if status, ok2 := statusRaw.(string); ok2 && status == "formed" && app.FormedAt == nil {
			now := time.Now()
			fields["formed_at"] = now
		}
	}

	if err := r.DB.Model(&Application{}).Where("id = ?", appID).Updates(fields).Error; err != nil {
		return nil, err
	}

	// перечитываем с позициями и пересчитываем total_price
	if err := r.DB.Preload("Items.Service").First(&app, appID).Error; err != nil {
		return nil, err
	}

	total := 0
	for _, it := range app.Items {
		total += it.Service.Price * it.Quantity
	}
	app.TotalPrice = total
	_ = r.DB.Model(&Application{}).Where("id = ?", app.ID).Update("total_price", total).Error

	return &app, nil
}

// SetApplicationStatus модифицирует статус заявки модератором (approve / reject).
func (r *Repository) SetApplicationStatus(
	appID uint,
	newStatus string,
	moderatorID uint,
) (*Application, error) {
	var app Application
	if err := r.DB.First(&app, appID).Error; err != nil {
		return nil, err
	}

	// разрешаем менять только сформированные заявки
	if app.Status != "formed" {
		return nil, fmt.Errorf("status change allowed only from 'formed'")
	}

	fields := map[string]any{
		"status":       newStatus,
		"moderator_id": moderatorID,
		"completed_at": time.Now(),
	}

	if err := r.DB.Model(&Application{}).Where("id = ?", appID).Updates(fields).Error; err != nil {
		return nil, err
	}

	if err := r.DB.Preload("Items.Service").First(&app, appID).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

// SoftDeleteApplication — логическое удаление заявки (смена статуса на deleted).
func (r *Repository) SoftDeleteApplication(appID, userID uint) error {
	return r.DB.Model(&Application{}).
		Where("id = ? AND created_by_id = ? AND status <> 'deleted'", appID, userID).
		Update("status", "deleted").Error
}

// drivingStyleCoef возвращает коэффициент ресурса колодок
// в зависимости от стиля вождения.
// Спокойный  -> ресурс больше, коэффициент > 1
// Спортивный -> базовый ресурс
// Агрессивный-> ресурс меньше, коэффициент < 1
func drivingStyleCoef(style string) float64 {
	s := strings.ToLower(style)

	switch {
	case strings.Contains(s, "спокой"):
		return 1.2
	case strings.Contains(s, "спорт"):
		return 1.0
	case strings.Contains(s, "агресс"):
		return 0.8
	default:
		return 1.0
	}
}