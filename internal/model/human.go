package model

type Human struct {
	Id          int    `json:"id" db:"omitempty" example:"1"`
	Name        string `json:"name" db:"name" example:"John"`
	Surname     string `json:"surname" db:"surname" example:"Doe"`
	Patronymic  string `json:"patronymic" db:"patronymic" example:"Ivanovich"`
	Age         int    `json:"age" db:"age" example:"25"`
	Gender      string `json:"gender" db:"gender" example:"male"`
	Nationality string `json:"nationality" db:"nationality" example:"RU"`
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
