package mock

type actives struct {
}

func NewActives() *actives {
	return &actives{}
}

func (a *actives) Track(int) error {
	return nil
}

func (a *actives) ActivesByDay() (map[string]int, error) {
	return make(map[string]int, 0), nil
}

func (a *actives) ActivesByWeek() (map[string]int, error) {
	return make(map[string]int, 0), nil
}

func (a *actives) ActivesByMonth() (map[string]int, error) {
	return make(map[string]int, 0), nil
}
