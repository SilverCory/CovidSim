package CovidSim

type Storage interface {
	LoadUser(id string) (u CovidUser, ok bool, err error)
	SaveUser(user CovidUser) error
}
