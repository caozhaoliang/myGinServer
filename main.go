package main

import (
	"myGinServer/config"
	"myGinServer/controller"
	"myGinServer/internal/store"
	"myGinServer/router"
)

const (
	prometheusMonitor = false
)

func main() {
	configYaml := config.InitConfig("./config.yaml")

	db, err := store.NewDatabase(&configYaml.DBConfig)
	if err != nil {
		panic(err)
	}

	userController := controller.NewUserController(db, configYaml)

	chatController := controller.NewChatController(configYaml)

	r := router.NewRouter(userController, chatController, db)

	if prometheusMonitor {
		go func() {
			router.NewPromethusMetricsEngine().Run(":18081")
		}()
	}

	err = r.Start(":8081")
	if err != nil {
		panic(err)
	}

}
