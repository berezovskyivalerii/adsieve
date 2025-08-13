package entity

type Ad struct {
	AdID      int64  `json:"ad_id"        db:"ad_id"`
	AccountID int64  `json:"account_id"   db:"account_id"`
	Name      string `json:"platfnameorm" db:"name"`
	Status    string `json:"status"       db:"status"`
	Platform  string `json:"platform"     db:"platform"`
}
