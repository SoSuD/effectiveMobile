package apiserver

// addHumanRequest represents the payload for adding a human
// swagger:model
type addHumanRequest struct {
	// имя
	// required: true
	Name string `json:"name"`
	// фамилия
	// required: true
	Surname string `json:"surname"`
	// отчество
	// required: false
	Patronymic string `json:"patronymic"`
}

// deleteHumanRequest represents the payload for deleting a human
// swagger:model
type deleteHumanRequest struct {
	// ID человека
	// required: true
	ID int `json:"id"`
}

// updateHumanRequest represents the payload for updating a human
// swagger:model
type updateHumanRequest struct {
	// ID человека
	// required: true
	ID int `json:"id"`
	// имя
	// required: false
	Name string `json:"name"`
	// фамилия
	// required: false
	Surname string `json:"surname"`
	// отчество
	// required: false
	Patronymic string `json:"patronymic"`
	// возраст
	// required: false
	Age int `json:"age"`
	// пол
	// required: false
	Gender string `json:"gender"`
	// национальность
	// required: false
	Nationality string `json:"nationality"`
}
