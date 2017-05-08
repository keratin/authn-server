package services

var MISSING = "MISSING"

type Error struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Account struct {
	Id       int
	Username string
}

func AccountCreator(username string, password string) (*Account, []Error) {
	errors := make([]Error, 0, 1)

	if username == "" {
		errors = append(errors, Error{Field: "username", Message: MISSING})
	}
	if password == "" {
		errors = append(errors, Error{Field: "password", Message: MISSING})
	}

	if len(errors) > 0 {
		return nil, errors
	}

	account := Account{
		Id:       0,
		Username: username,
	}

	return &account, nil
}
