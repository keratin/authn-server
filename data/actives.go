package data

type Actives interface {
	Track(int) error
	ActivesByDay() (map[string]int, error)
	ActivesByWeek() (map[string]int, error)
	ActivesByMonth() (map[string]int, error)
}
