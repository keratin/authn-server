package data

type Account struct {
	Id       int
	Username string
}

type AccountStore interface {
	Create(u string, p string) (*Account, error)
}
