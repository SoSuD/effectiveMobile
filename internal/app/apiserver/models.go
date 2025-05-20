package apiserver

// addHumanRequest represents the payload for adding a human
// swagger:model
type addHumanRequest struct {
	// имя
	// required: true
	Name string `json:"name" example:"John"`
	// фамилия
	// required: true
	Surname string `json:"surname" example:"Doe"`
	// отчество
	// required: false
	Patronymic string `json:"patronymic" example:"Johnny"`
}

// deleteHumanRequest represents the payload for deleting a human
// swagger:model
type deleteHumanRequest struct {
	// ID человека
	// required: true
	ID int `json:"id" example:"1"`
}

// updateHumanRequest represents the payload for updating a human
// swagger:model
type updateHumanRequest struct {
	// ID человека
	// required: true
	ID int `json:"id" example:"1"`
	// имя
	// required: false
	Name string `json:"name" example:"John"`
	// фамилия
	// required: false
	Surname string `json:"surname" example:"Doe"`
	// отчество
	// required: false
	Patronymic string `json:"patronymic" example:"Johnny"`
	// возраст
	// required: false
	Age int `json:"age" example:"30"`
	// пол
	// required: false
	Gender string `json:"gender" example:"male"`
	// национальность
	// required: false
	Nationality string `json:"nationality" example:"RU"`
}
