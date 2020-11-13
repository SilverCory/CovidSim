package storage

import (
	"errors"
	"fmt"
	"github.com/SilverCory/CovidSim"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// TODO not use gorm, it's not required and bleeds onto CovidSim.CovidUser.

var _ CovidSim.Storage = &MySQL{}

type MySQL struct {
	db *gorm.DB
}

func NewMySQL(dsn string) (*MySQL, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("unable to open gorm: %w", err)
	}

	if err := db.AutoMigrate(&CovidSim.CovidUser{}); err != nil {
		return nil, fmt.Errorf("unable to migrate coviduser: %w", err)
	}

	return &MySQL{db: db}, nil
}

func (m *MySQL) LoadUser(id string) (u CovidSim.CovidUser, ok bool, err error) {
	var ret = new(CovidSim.CovidUser)
	ret.ID = id
	res := m.db.First(&ret)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return CovidSim.CovidUser{}, false, nil
	} else if res.Error != nil {
		return CovidSim.CovidUser{}, false, err
	}

	return *ret, true, nil
}

func (m *MySQL) SaveUser(u CovidSim.CovidUser) error {
	err := m.db.Save(&u).Error
	if err != nil {
		return fmt.Errorf("saveuser: %w", err)
	}

	return nil
}
