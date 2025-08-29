package entity

// Бэкендовая сущность объявления (как в таблице ads)
type Ad struct {
	AdID      int64  `json:"ad_id"      db:"ad_id"`
	AccountID int64  `json:"account_id" db:"account_id"`
	Name      string `json:"name"       db:"name"`     // ← исправлен json-тег
	Status    string `json:"status"     db:"status"`   // active | paused
	Platform  string `json:"platform"   db:"platform"` // facebook | google
}

// DTO для ответа GET /api/ads (без внутренних полей)
type AdDTO struct {
	AdID     int64  `json:"ad_id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Platform string `json:"platform"`
}

// Фильтр списка объявлений
type AdsFilter struct {
	Status   *string 
	Platform *string 
	Query    *string 
	AdIDs    []int64
	Limit    int
	Offset   int
	Sort     string
}
