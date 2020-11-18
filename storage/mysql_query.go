package storage

import (
	"fmt"
	"github.com/SilverCory/CovidSim"
	"time"
)

func (m *MySQL) GetUser(id string) (u CovidSim.CovidUser, err error) {
	u, _, err = m.LoadUser(id)
	return
}

func (m *MySQL) GetUserInfections(id string) (fromID string, ids []string, err error) {
	fromID = id // TODO add a default to patient zero.
	rows, err := m.db.Model(&CovidSim.CovidUser{}).
		Select("id").
		Where(
			"contraction_time <> ? AND contraction_time IS NOT NULL AND contracted_by = ?",
			time.Unix(0, 0),
			id,
		).
		Order("contraction_time ASC").
		Rows()
	if err != nil {
		return "", nil, fmt.Errorf("failed GetUserInfections: rows: %w", err)
	}

	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return "", nil, fmt.Errorf("failed GetUserInfections: scan: %w", err)
		}
		ids = append(ids, id)
	}

	return
}

func (m *MySQL) GetInfections(args CovidSim.InfectionQueryArguments) (response CovidSim.InfectionQueryResponse, err error) {
	panic("implement me")
}
