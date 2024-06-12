package app

import (
	"github.com/neurlang/goruut/helpers/log"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)
import . "github.com/martinarisk/di/dependency_injection"
import _ "github.com/neurlang/goruut/controllers/v0"
import . "github.com/neurlang/goruut/controllers"
import "github.com/gorilla/mux"
import "net"
import "net/http"

// Initialize initializes the server with the specified port.
func (s *Server) Initialize(port, adminPort string) {

	log.Now().Infof("Binding port: %s", port)

	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Now().Fatalf("Server couldn't listen: %e", err)
	}
	log.Now().Infof("Binding admin port: %s", adminPort)

	l2, err := net.Listen("tcp", ":"+adminPort)
	if err != nil {
		log.Now().Fatalf("Server couldn't listen: %e", err)
	}

	s.httpServerListener = l
	s.adminServerListener = l2
}

// RunForever starts the server and makes it run indefinitely.
func (s *Server) RunForever() {

	log.Now().Infof("Serving...")

	s.httpServer = &http.Server{
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      s.httpServerHandler,
	}
	s.adminServer = &http.Server{
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      s.adminServerHandler,
	}
	go func() { log.Fatal0(s.adminServer.Serve(s.adminServerListener)) }()
	log.Fatal0(s.httpServer.Serve(s.httpServerListener))
}

type router interface {
	Queries(buffer ...string)
}
type myRouter struct {
	data []*mux.Route
}

func (m *myRouter) Queries(buffer ...string) {
	for _, r := range m.data {
		r.Queries(buffer...)
	}
}

// NewAppControllers creates new instances of application controllers.
func (app *App) NewAppControllers(di *DependencyInjection) *Controllers {

	log.Now().Infof("Initializing %d controllers...", len(AllControllers))

	for _, handle := range AllControllers {

		log.Now().Debugf("Initializing controller: %T", handle)

		handle.Init(di)
	}

	log.Now().Infof("Initialized %d controllers", len(AllControllers))

	s := MustAny[*Server](di)
	v := MustAny[*Views](di)
	conf := MustAny[*Configs](di)

	r := mux.NewRouter()
	r2 := mux.NewRouter()

	const apiPrefix = "/api"
	const getPrefix = "/tts"

	setRouter := r.PathPrefix(getPrefix).Subrouter()
	apiRouter := r2.PathPrefix(apiPrefix).Subrouter()

	handleIt := func(path string, handler Controller) router {
		switch handler.BackendType() {
		case MainController:
			log.Now().Debugf("Connecting controller: %s%s: %T", getPrefix, path, handler)
			return &myRouter{data: []*mux.Route{setRouter.Handle(path, handler)}}
		case AdminController:
			log.Now().Debugf("Connecting controller: %s%s: %T", apiPrefix, path, handler)
			return &myRouter{data: []*mux.Route{apiRouter.Handle(path, handler)}}
		case AllController:
			log.Now().Debugf("Connecting controller: %s%s: %T", getPrefix, path, handler)
			log.Now().Debugf("Connecting controller: %s%s: %T", apiPrefix, path, handler)
			qa := apiRouter.Handle(path, handler)
			qb := setRouter.Handle(path, handler)
			return &myRouter{data: []*mux.Route{qa, qb}}
		}
		return nil
	}

	// Backend handlers
	for path, handle := range AllControllers {

		if strings.Contains(path, "?") {
			path2 := strings.SplitN(path, "?", 2)
			path = path2[0]
			qs := strings.SplitN(path2[1], "&", -1)
			var buffer []string
			for _, q := range qs {
				qss := strings.SplitN(q, "=", 2)
				buffer = append(buffer, qss[0], qss[1])
			}
			handleIt(path+"/", handle).Queries(buffer...)
			handleIt(path+"/", handle)

			handleIt(path, handle).Queries(buffer...)
			handleIt(path, handle)
		} else {
			handleIt(path, handle)
		}
	}

	// Frontend icon handler

	remote, err := url.Parse(conf.GetFavIconSite())
	if err == nil {
		handler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
			return func(w http.ResponseWriter, r *http.Request) {
				r.RequestURI = ""
				r.Host = remote.Host
				w.Header().Set("X-GopherIcon", "HiFromProxy")
				p.ServeHTTP(w, r)
			}
		}
		proxy := httputil.NewSingleHostReverseProxy(remote)
		r.PathPrefix("/favicon.ico").HandlerFunc(handler(proxy))
	}

	// Frontend handlers
	r2.PathPrefix("/static/" + defaultFrontendVersion).Handler(http.StripPrefix("/static/", v))
	r2.PathPrefix("/static/").Handler(http.StripPrefix("/static/", v))
	r2.PathPrefix("/" + defaultFrontendVersion).Handler(v)
	r2.PathPrefix("/index.html").Handler(v)
	r2.PathPrefix("").Handler(v)

	s.httpServerHandler = r
	s.adminServerHandler = r2

	return &Controllers{}
}
