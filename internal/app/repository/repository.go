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
// УСЛУГА — ТОРМОЗНЫЕ КОЛОДКИ
// =====================

type BrakePad struct {
	ID            int
	Title         string
	Description   string
	Image         string // ключ изображения (Minio)
	Video         string // ключ видеоролика (Minio)
	PadType       string // тип колодок
	BaseResource  int    // базовый ресурс (км)
	Price         int    // цена (руб)
}

// =====================
// ЗАЯВКА — РАСЧЁТ ИЗНОСА
// =====================

type WearCalculation struct {
	ID                int
	Title             string
	Description       string
	DrivingStyle      string  // стиль вождения
	Mileage           int     // текущий пробег
	ResultRemainingKM float64 // остаток в км
	ResultPercent     float64 // остаток в %
	Status            string
	Pads              []CalculationPad
	PadCount          int
}

// связь м-м
type CalculationPad struct {
	Pad     BrakePad
	Comment string
}

// =====================
// ДАННЫЕ УСЛУГ
// =====================

func (r *Repository) GetPads() ([]BrakePad, error) {
	pads := []BrakePad{
		{
			ID:           1,
			Title:        "Керамические тормозные колодки",
			Description:  "Керамические колодки обеспечивают стабильное торможение, низкий уровень шума и минимальное пылеобразование. Отличаются увеличенным сроком службы.",
			Image:        "ceramic.jpg",
			Video:        "ceramic.MP4",
			PadType:      "Керамические",
			BaseResource: 60000,
			Price:        8500,
		},
		{
			ID:           2,
			Title:        "Органические тормозные колодки",
			Description:  "Органические колодки отличаются мягкой работой и доступной стоимостью. Подходят для спокойного городского режима.",
			Image:        "organic.jpg",
			Video:        "organic.MP4",
			PadType:      "Органические",
			BaseResource: 35000,
			Price:        4500,
		},
		{
			ID:           3,
			Title:        "Полуметаллические тормозные колодки",
			Description:  "Полуметаллические колодки обеспечивают эффективное торможение при высоких нагрузках и подходят для активной езды.",
			Image:        "semi_metallic.jpg",
			Video:        "semi_metallic.MP4",
			PadType:      "Полуметаллические",
			BaseResource: 45000,
			Price:        6500,
		},
	}

	return pads, nil
}

// =====================
// ФИЛЬТР ПО НАЗВАНИЮ
// =====================

func (r *Repository) GetPadsByTitle(query string) ([]BrakePad, error) {
	pads, _ := r.GetPads()
	q := strings.ToLower(query)

	var result []BrakePad
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

func (r *Repository) GetPad(id int) (BrakePad, error) {
	pads, _ := r.GetPads()
	for _, p := range pads {
		if p.ID == id {
			return p, nil
		}
	}
	return BrakePad{}, fmt.Errorf("колодки не найдены")
}

// =====================
// КОЭФФИЦИЕНТ СТИЛЯ
// =====================

func getDrivingCoefficient(style string) float64 {
	switch style {
	case "Спокойный":
		return 1.0
	case "Спортивный":
		return 0.8
	case "Агрессивный":
		return 0.6
	default:
		return 1.0
	}
}

// =====================
// ПОСТРОЕНИЕ ЗАЯВКИ
// =====================

func (r *Repository) buildCalculation(id int, drivingStyle string, mileage int, padID int) (WearCalculation, error) {
	pad, err := r.GetPad(padID)
	if err != nil {
		return WearCalculation{}, err
	}

	coef := getDrivingCoefficient(drivingStyle)
	effectiveResource := float64(pad.BaseResource) * coef
	remaining := effectiveResource - float64(mileage)

	if remaining < 0 {
		remaining = 0
	}

	percent := (remaining / effectiveResource) * 100

	return WearCalculation{
		ID:                id,
		Title:             "Расчёт остаточного ресурса тормозных колодок",
		Description:       "Расчёт выполнен на основе типа колодок и стиля вождения.",
		DrivingStyle:      drivingStyle,
		Mileage:           mileage,
		ResultRemainingKM: remaining,
		ResultPercent:     percent,
		Status:            "Рассчитано",
		Pads: []CalculationPad{
			{
				Pad:     pad,
				Comment: "Основной комплект колодок",
			},
		},
		PadCount: 1,
	}, nil
}

// =====================
// СПИСОК ЗАЯВОК
// =====================

func (r *Repository) GetCalculations() ([]WearCalculation, error) {
	calc, err := r.buildCalculation(
		1,
		"Спортивный",
		20000,
		1,
	)
	if err != nil {
		return nil, err
	}

	return []WearCalculation{calc}, nil
}

func (r *Repository) GetCalculation(id int) (WearCalculation, error) {
	calcs, _ := r.GetCalculations()
	for _, c := range calcs {
		if c.ID == id {
			return c, nil
		}
	}
	return WearCalculation{}, fmt.Errorf("заявка не найдена")
}

func (r *Repository) GetCalculationForPad(padID int) (*CalculationPad, error) {
	calcs, _ := r.GetCalculations()
	for _, c := range calcs {
		for _, p := range c.Pads {
			if p.Pad.ID == padID {
				return &p, nil
			}
		}
	}
	return nil, fmt.Errorf("не найдено")
}