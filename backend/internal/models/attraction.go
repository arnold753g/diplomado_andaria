package models

import "time"

type AttractionCategory struct {
	ID            uint64                  `gorm:"column:id;primaryKey" json:"id"`
	Name          string                  `gorm:"column:nombre" json:"name"`
	Position      int                     `gorm:"column:posicion" json:"-"`
	Subcategories []AttractionSubcategory `gorm:"foreignKey:CategoryID" json:"subcategories"`
}
type AttractionSubcategory struct {
	ID         uint64 `gorm:"column:id;primaryKey" json:"id"`
	CategoryID uint64 `gorm:"column:categoria_id" json:"category_id"`
	Name       string `gorm:"column:nombre" json:"name"`
	Position   int    `gorm:"column:posicion" json:"-"`
}
type AttractionClassification struct {
	AttractionID  uint64                `gorm:"column:atraccion_id;primaryKey" json:"-"`
	SubcategoryID uint64                `gorm:"column:subcategoria_id;primaryKey" json:"subcategory_id"`
	Position      int                   `gorm:"column:posicion" json:"position"`
	Subcategory   AttractionSubcategory `gorm:"foreignKey:SubcategoryID" json:"subcategory"`
}

type Attraction struct {
	Subcategories    []AttractionClassification `gorm:"foreignKey:AttractionID" json:"subcategories"`
	ScheduleMode     string                     `gorm:"column:modo_horario" json:"schedule_mode"`
	OpeningTime      string                     `gorm:"column:hora_apertura" json:"opening_time"`
	ClosingTime      string                     `gorm:"column:hora_cierre" json:"closing_time"`
	OpeningDays      []int                      `gorm:"column:dias_apertura;serializer:json" json:"opening_days"`
	SeasonMode       string                     `gorm:"column:modo_temporada" json:"season_mode"`
	SeasonStartMonth *int                       `gorm:"column:mes_inicio_temporada" json:"season_start_month"`
	SeasonEndMonth   *int                       `gorm:"column:mes_fin_temporada" json:"season_end_month"`
	ID               uint64                     `gorm:"column:id;primaryKey" json:"id"`
	Name             string                     `gorm:"column:nombre" json:"name"`
	Description      string                     `gorm:"column:descripcion" json:"description"`
	Category         string                     `gorm:"column:categoria" json:"category"`
	Department       string                     `gorm:"column:departamento" json:"department"`
	City             string                     `gorm:"column:ciudad" json:"city"`
	Address          string                     `gorm:"column:direccion" json:"address"`
	Latitude         *float64                   `gorm:"column:latitud" json:"latitude"`
	Longitude        *float64                   `gorm:"column:longitud" json:"longitude"`
	OpeningHours     string                     `gorm:"column:horario_atencion" json:"opening_hours"`
	AdmissionCents   int                        `gorm:"column:entrada_centavos" json:"admission_cents"`
	Recommendations  string                     `gorm:"column:recomendaciones" json:"recommendations"`
	Phone            string                     `gorm:"column:telefono" json:"phone"`
	ManagerID        uint64                     `gorm:"column:encargado_id" json:"-"`
	Manager          *User                      `gorm:"foreignKey:ManagerID" json:"-"`
	Status           string                     `gorm:"column:estado" json:"-"`
	Published        bool                       `gorm:"column:publicado" json:"-"`
	Version          int                        `gorm:"column:version" json:"-"`
	Photos           []AttractionPhoto          `gorm:"foreignKey:AttractionID" json:"photos"`
	CreatedAt        time.Time                  `gorm:"column:creado_en" json:"-"`
	UpdatedAt        time.Time                  `gorm:"column:actualizado_en" json:"updated_at"`
}

type AttractionPhoto struct {
	ID           uint64 `gorm:"column:id;primaryKey" json:"id"`
	AttractionID uint64 `gorm:"column:atraccion_id" json:"-"`
	Position     int    `gorm:"column:posicion" json:"position"`
	Image        []byte `gorm:"column:imagen" json:"-"`
}

type AttractionFavorite struct {
	UserID       uint64    `gorm:"column:usuario_id;primaryKey" json:"-"`
	AttractionID uint64    `gorm:"column:atraccion_id;primaryKey" json:"attraction_id"`
	CreatedAt    time.Time `gorm:"column:creado_en" json:"created_at"`
}

func (Attraction) TableName() string { return "atracciones" }

func (AttractionPhoto) TableName() string { return "fotos_atracciones" }

func (AttractionFavorite) TableName() string { return "favoritos_atracciones" }

func (AttractionCategory) TableName() string { return "categorias_atracciones" }

func (AttractionSubcategory) TableName() string { return "subcategorias_atracciones" }

func (AttractionClassification) TableName() string { return "clasificaciones_atracciones" }
