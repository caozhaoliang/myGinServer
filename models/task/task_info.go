package task

type Task struct {
	Id        string `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	Des       string `json:"des" db:"des"`
	Completed bool   `json:"completed" db:"completed"`
}
