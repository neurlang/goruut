// Package controllers handles HTTP request routing and API layer logic.
package controllers

import (
	"net/http"
)
import . "github.com/martinarisk/di/dependency_injection"

type ControllerBackendType byte

const (
	MainController ControllerBackendType = 1 + iota
	AdminController
	AllController
)

// Controller is the interface a controller must satisfy
type Controller interface {
	Init(di *DependencyInjection)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	BackendType() ControllerBackendType
}

// Set defines the controllers set for storing all controllers
type Set map[string]Controller

// AllControllers contains all the controllers that are registered by the application
var AllControllers = make(Set)
