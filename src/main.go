package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"ugc_test_task/src/config"
	"ugc_test_task/src/http"
	"ugc_test_task/src/logger"
	buildmng "ugc_test_task/src/managers/buildings"
	categmng "ugc_test_task/src/managers/categories"
	companmng "ugc_test_task/src/managers/companies"
	"ugc_test_task/src/pg"
	buildrepos "ugc_test_task/src/repositories/buildings"
	categrepos "ugc_test_task/src/repositories/categories"
	companrepos "ugc_test_task/src/repositories/companies"
)

var (
	conf config.Config

	categoryRepos categrepos.Repository
	companyRepos  companrepos.Repository
	buildingRepos buildrepos.Repository

	companyMng  companmng.Manager
	buildingMng buildmng.Manager
	categoryMng categmng.Manager

	httpApi http.Api
)

func main() {
	var err error
	conf, err = config.New()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error while creating config: %v\n", err)
		os.Exit(1)
	}
	if err := initLogger(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error while init logger: %v\n", err)
		os.Exit(1)
	}
	if err := initRepositories(); err != nil {
		logger.Msg("error while init repositories").Error(err.Error())
		os.Exit(1)
	}
	if err := initManagers(); err != nil {
		logger.Msg("error while init managers").Error(err.Error())
		os.Exit(1)
	}
	httpApi = http.NewApi(http.Config{
		Host:              conf.HttpServer.Host,
		Port:              conf.HttpServer.Port,
		MetricsPort:       conf.HttpServer.MetricsPort,
		DebugPort:         conf.HttpServer.DebugPort,
		ReadTimeout:       conf.HttpServer.ReadTimeout,
		ReadHeaderTimeout: conf.HttpServer.ReadHeaderTimeout,
		WriteTimeout:      conf.HttpServer.WriteTimeout,
		IdleTimeout:       conf.HttpServer.IdleTimeout,
		MaxHeaderBytes:    conf.HttpServer.MaxHeaderBytes,
		CompanyManager:    companyMng,
		BuildingManager:   buildingMng,
		CategoryManager:   categoryMng,
	})
	if err != nil {
		logger.Msg("error while creating http api").Error(err.Error())
		os.Exit(1)
	}
	httpApi.Start(func(err error) {
		logger.Msg("error while start http api").Error(err.Error())
		os.Exit(1)
	})
	handleOsSignals()
}

func initLogger() (err error) {
	return logger.Init(logger.Config{
		Path:   conf.Logger.Path,
		Stdout: conf.Logger.Stdout,
		Stderr: conf.Logger.Stderr,
		Level:  logger.LevelFromString(conf.Logger.Level),
	})
}

func initRepositories() (err error) {
	pgConfig := pg.Config{
		Host:     conf.Pg.Host,
		Port:     conf.Pg.Port,
		Database: conf.Pg.DbName,
		User:     conf.Pg.User,
		Password: conf.Pg.Password,
	}

	buildingRepos, err = buildrepos.New(buildrepos.NewConfig(pgConfig))
	if err != nil {
		return fmt.Errorf("init 'building' repository: %v", err)
	}

	categoryRepos, err = categrepos.New(categrepos.NewConfig(pgConfig))
	if err != nil {
		return fmt.Errorf("init 'category' repository: %v", err)
	}

	companyConf := companrepos.NewConfig(pgConfig)
	companyConf.CategoryRepos = categoryRepos
	companyRepos, err = companrepos.New(companyConf)
	if err != nil {
		return fmt.Errorf("init 'company' repository: %v", err)
	}

	return nil
}

func initManagers() (err error) {
	companyMng, err = companmng.New(companmng.Config{
		CompanyRepos: companyRepos,
	})
	if err != nil {
		return fmt.Errorf("error while creating company manager: %v", err)
	}

	buildingMng, err = buildmng.New(buildmng.Config{
		BuildingRepos: buildingRepos,
	})
	if err != nil {
		return fmt.Errorf("error while creating building manager: %v", err)
	}

	categoryMng, err = categmng.New(categmng.Config{
		CategoryRepos: categoryRepos,
	})
	if err != nil {
		return fmt.Errorf("error while creating category manager: %v", err)
	}
	return nil
}

func handleOsSignals() {
	osSignals := make(chan os.Signal)
	defer close(osSignals)
	signal.Notify(osSignals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	for {
		<-osSignals
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		httpApi.Shutdown(ctx)
		if err := buildingRepos.Stop(ctx); err != nil {
			logger.Msg("shutdown 'building' repository").Error(err.Error())
		}
		if err := categoryRepos.Stop(ctx); err != nil {
			logger.Msg("shutdown 'category' repository").Error(err.Error())
		}
		if err := companyRepos.Stop(ctx); err != nil {
			logger.Msg("shutdown 'company' repository").Error(err.Error())
		}
		cancel()
	}
}
