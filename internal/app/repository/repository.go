package repository

import (
	"fmt"
	"strings"
)

// Repository — хранилище данных (Lab 1: без БД).
type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

// =====================
// УСЛУГА — ПРОГНОЗ ИЗНОСА ТОРМОЗНЫХ КОЛОДОК
// =====================

// BrakePadService описывает тип колодок и коэффициенты износа для стилей вождения.
type BrakePadService struct {
	ID           int
	Title        string
	Description  string
	Image        string // ключ изображения (Minio)
	Video        string // ключ видеоролика (Minio)
	PadType      string // тип колодок
	BaseResource int    // базовый ресурс (км)

	// краткое описание рекомендованного стиля вождения
	DrivingStyleHint string

	// коэффициенты износа по стилю вождения
	CalmCoef       float64 // спокойный
	SportCoef      float64 // спортивный
	AggressiveCoef float64 // агрессивный
}

// =====================
// ЗАЯВКА brake pad wear
// =====================

// BrakePadWear — заявка на расчёт износа колодок.
type BrakePadWear struct {
	ID          int
	CarInfo     string  // описание автомобиля текстом
	DrivingStyle string  // стиль вождения
	Mileage      int     // пробег в км
	Status       string

	Items []BrakePadWearItem
}

// BrakePadWearItem — связь заявки и услуги (м-м) с пробегом и результатом.
type BrakePadWearItem struct {
	Service       BrakePadService
	Mileage       int     // пробег для этой услуги (км)
	RemainingKM   float64 // результат: остаточный ресурс колодок в км
}

// =====================
// ДАННЫЕ УСЛУГ
// =====================

func (r *Repository) GetPads() ([]BrakePadService, error) {
	pads := []BrakePadService{
		{
			ID:           1,
			Title:        "Керамические тормозные колодки",
			Description:  "Керамические колодки обеспечивают стабильное торможение, низкий уровень шума и минимальное пылеобразование. Отличаются увеличенным сроком службы.",
			Image:        "ceramic.jpg",
			Video:        "ceramic.MP4",
			PadType:      "Керамические",
			BaseResource: 60000,
			DrivingStyleHint: "Спокойный и размеренный.",
			CalmCoef:     1.0,
			SportCoef:    0.85,
			AggressiveCoef: 0.7,
		},
		{
			ID:           2,
			Title:        "Органические тормозные колодки",
			Description:  "Органические колодки отличаются мягкой работой и доступной стоимостью. Подходят для спокойного городского режима.",
			Image:        "organic.jpg",
			Video:        "organic.MP4",
			PadType:      "Органические",
			BaseResource: 35000,
			DrivingStyleHint: "Спокойный.",
			CalmCoef:     1.0,
			SportCoef:    0.8,
			AggressiveCoef: 0.6,
		},
		{
			ID:           3,
			Title:        "Полуметаллические тормозные колодки",
			Description:  "Полуметаллические колодки обеспечивают эффективное торможение при высоких нагрузках и подходят для активной езды.",
			Image:        "semi_metallic.jpg",
			Video:        "semi_metallic.MP4",
			PadType:      "Полуметаллические",
			BaseResource: 45000,
			DrivingStyleHint: "Спортивный, размеренный.",
			CalmCoef:     1.0,
			SportCoef:    0.9,
			AggressiveCoef: 0.75,
		},
	}

	return pads, nil
}

// =====================
// ФИЛЬТР ПО НАЗВАНИЮ
// =====================

func (r *Repository) GetPadsByTitle(query string) ([]BrakePadService, error) {
	pads, _ := r.GetPads()
	q := strings.ToLower(query)

	var result []BrakePadService
	for _, p := range pads {
		if strings.Contains(strings.ToLower(p.Title), q) ||
			strings.Contains(strings.ToLower(p.Description), q) ||
			strings.Contains(strings.ToLower(p.PadType), q) {
			result = append(result, p)
		}
	}
	return result, nil
}

// =====================
// ПОЛУЧИТЬ КОЛОДКУ ПО ID
// =====================

func (r *Repository) GetPad(id int) (BrakePadService, error) {
	pads, _ := r.GetPads()
	for _, p := range pads {
		if p.ID == id {
			return p, nil
		}
	}
	return BrakePadService{}, fmt.Errorf("колодки не найдены")
}

// =====================
// КОЭФФИЦИЕНТ СТИЛЯ
// =====================

func getDrivingCoefficient(p BrakePadService, style string) float64 {
	switch style {
	case "Спокойный":
		return p.CalmCoef
	case "Спортивный":
		return p.SportCoef
	case "Агрессивный":
		return p.AggressiveCoef
	default:
		return p.CalmCoef
	}
}

// =====================
// ПОСТРОЕНИЕ ЗАЯВКИ
// =====================

func (r *Repository) buildCalculation(id int, carInfo string, drivingStyle string, mileage int, padID int) (BrakePadWear, error) {
	pad, err := r.GetPad(padID)
	if err != nil {
		return BrakePadWear{}, err
	}

	coef := getDrivingCoefficient(pad, drivingStyle)
	effectiveResource := float64(pad.BaseResource) * coef
	remaining := effectiveResource - float64(mileage)

	if remaining < 0 {
		remaining = 0
	}

	item := BrakePadWearItem{
		Service:     pad,
		Mileage:     mileage,
		RemainingKM: remaining,
	}

	return BrakePadWear{
		ID:          id,
		CarInfo:     carInfo,
		DrivingStyle: drivingStyle,
		Mileage:      mileage,
		Status:       "Рассчитано",
		Items:        []BrakePadWearItem{item},
	}, nil
}

// =====================
// СПИСОК ЗАЯВОК
// =====================

func (r *Repository) GetCalculations() ([]BrakePadWear, error) {
	calc, err := r.buildCalculation(
		1,
		"Kia Rio X-Line, 2020 г.",
		"Спортивный",
		20000,
		1,
	)
	if err != nil {
		return nil, err
	}

	return []BrakePadWear{calc}, nil
}

func (r *Repository) GetCalculation(id int) (BrakePadWear, error) {
	calcs, _ := r.GetCalculations()
	for _, c := range calcs {
		if c.ID == id {
			return c, nil
		}
	}
	return BrakePadWear{}, fmt.Errorf("заявка не найдена")
}