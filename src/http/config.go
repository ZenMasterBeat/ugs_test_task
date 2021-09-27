package http

import (
	"fmt"
	config2 "ugc_test_task/src/config"
	"ugc_test_task/src/managers/buildings"
	"ugc_test_task/src/managers/categories"
	"ugc_test_task/src/managers/companies"
	models2 "ugc_test_task/src/models"
)

type CompanyManager interface {
	GetCompanies(query companies.GetQuery, clb func(firm models2.Company) error) error
	AddCompany(query companies.AddQuery) (models2.Company, error)
}

type BuildingManager interface {
	GetBuildings(query buildings.GetQuery, callback func(models2.Building) error) error
	AddBuilding(query buildings.AddQuery) (models2.Building, error)
}

type CategoryManager interface {
	AddCategory(query categories.AddQuery) (models2.Category, error)
	GetCategories(query categories.GetQuery, callback func(models2.Category) error) error
}

type Config struct {
	Host              string
	Port              string
	MetricsPort       string
	DebugPort         string
	ReadTimeout       config2.Duration
	ReadHeaderTimeout config2.Duration
	WriteTimeout      config2.Duration
	IdleTimeout       config2.Duration
	MaxHeaderBytes    config2.Bytes
	CompanyManager    CompanyManager
	BuildingManager   BuildingManager
	CategoryManager   CategoryManager
}

func (conf Config) Address() string {
	return conf.Host + ":" + conf.Port
}

func (conf Config) MetricsAddress() string {
	return conf.Host + ":" + conf.MetricsPort
}

func (conf Config) DebugAddress() string {
	return conf.Host + ":" + conf.DebugPort
}

func (conf Config) Validate() error {
	if len(conf.Host) == 0 {
		return fmt.Errorf("'host' is empty")
	}
	if len(conf.Port) == 0 {
		return fmt.Errorf("'port' is empty")
	}

	return nil
}