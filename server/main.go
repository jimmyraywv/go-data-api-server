package main

import "C"
import (
	"flag"
	"fmt"
	"jimmyray.io/data-api/pkg/model"
	"net/http"
	"os"

	Log "github.com/sirupsen/logrus"
	"jimmyray.io/data-api/pkg/utils"

	"github.com/gorilla/mux"
)

func initService() {
	flagAppName := "Application name"
	flagLogLevel := "Application log-level"
	flagMock := "Enable data mocking"

	var appName string
	var logLevel string
	var mock bool

	flag.StringVar(&appName, "name", "apis", flagAppName)
	flag.StringVar(&appName, "n", "apis", flagAppName)
	flag.StringVar(&logLevel, "level", "info", flagLogLevel)
	flag.StringVar(&logLevel, "l", "info", flagLogLevel)
	flag.BoolVar(&mock, "mock", false, flagMock)
	flag.BoolVar(&mock, "m", false, flagMock)
	flag.Parse()

	model.ServiceInfo.NAME = appName
	model.ServiceInfo.ID = model.GetServiceId()

	hostName, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	fields := Log.Fields{
		"hostname": hostName,
		"service":  model.ServiceInfo.NAME,
		"id":       model.ServiceInfo.ID,
	}

	var level Log.Level
	switch logLevel {
	case "debug":
		level = Log.DebugLevel
	case "error":
		level = Log.ErrorLevel
	case "fatal":
		level = Log.FatalLevel
	case "warn":
		level = Log.WarnLevel
	default:
		level = Log.InfoLevel
	}

	utils.InitLogs(fields, level)

	utils.Logger.WithFields(utils.StandardFields).WithFields(Log.Fields{"args": os.Args, "mode": "init", "logLevel": level}).Info("Service started successfully.")

	model.InitValidator()

	model.C = model.Controller{
		L:        &model.L,
		Validate: model.Validate,
	}

	model.IC = model.InfoController{
		ServiceInfo: model.ServiceInfo,
	}

	model.L.ServiceData = make(map[string]model.Employee)

	if mock {
		err := model.LoadMockData()

		if err == nil {
			utils.Logger.Info("Mock data loaded successfully.")
		} else {
			errorData := utils.ErrorLog{Skip: 1, Event: model.MockDataErr, Message: err.Error()}
			utils.LogErrors(errorData)
		}
	}
}

func Router() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	return r
}

func main() {
	initService()

	utils.Logger.WithFields(utils.StandardFields).WithFields(Log.Fields{"mode": "run"}).Info("Listening on port 8080")

	//router := mux.NewRouter().StrictSlash(true)
	router := Router()
	router.HandleFunc("/healthz", model.IC.HealthCheck).Methods(http.MethodGet)
	router.HandleFunc("/info", model.IC.GetServiceInfo).Methods(http.MethodGet)
	router.HandleFunc("/data", model.C.GetAllData).Methods(http.MethodGet)
	router.HandleFunc("/data", model.C.CreateData).Methods(http.MethodPut)
	router.HandleFunc("/data/{id}", model.C.GetData).Methods(http.MethodGet)
	router.HandleFunc("/data", model.C.PatchData).Methods(http.MethodPatch)
	router.HandleFunc("/data/{id}", model.C.DeleteData).Methods(http.MethodDelete)

	fmt.Println(http.ListenAndServe(":8080", router))
}
