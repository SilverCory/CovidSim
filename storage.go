package CovidSim

import "time"

type Storage interface {
	LoadUser(id string) (u CovidUser, ok bool, err error)
	SaveUser(user CovidUser) error
}

type QueryStorage interface {
	GetUser(id string) (u CovidUser, err error)
	GetUserInfections(id string) (fromID string, ids []string, err error)
	GetInfections(args InfectionQueryArguments) (response InfectionQueryResponse, err error)
}

type InfectionQueryArguments struct {
	Before  time.Time
	After   time.Time
	Divisor time.Duration
}

type InfectionQueryResponse struct {
	At    time.Time
	Count int
}
