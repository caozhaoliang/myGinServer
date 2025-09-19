package article

import "time"

type Channel struct {
	Id        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Status    string    `json:"status" db:"status"`
	Des       string    `json:"des" db:"des"`
	CreatedOn time.Time `json:"created_on" db:"created_on"`
}
