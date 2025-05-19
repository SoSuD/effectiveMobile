package model

type Human struct {
	Id          int    `json:"id" db:"omitempty"`
	Name        string `json:"name" db:"name"`
	Surname     string `json:"surname" db:"surname"`
	Patronymic  string `json:"patronymic" db:"patronymic"`
	Age         int    `json:"age" db:"age"`
	Gender      string `json:"gender" db:"gender"`
	Nationality string `json:"nationality" db:"nationality"`
}

type HumanFilter struct {
	ID          int
	Name        string
	Surname     string
	Patronymic  string
	MinAge      int
	MaxAge      int
	Gender      string
	Nationality string

	Page     int
	PageSize int
}
