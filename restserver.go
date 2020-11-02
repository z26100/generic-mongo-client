package mongo

import (
	"context"
	"crypto/tls"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
	"strings"
	"time"
)

type RestServer struct {
	r      *mux.Router
	srv    *http.Server
	config ServerConfig
}

type ServerConfig struct {
	Listen                    string
	ProductionMode            bool
	CertFile, KeyFile         string
	ReadTimeout, WriteTimeout time.Duration
	TlsConfig                 *tls.Config
}

type Route struct {
	Path       string
	PathPrefix string
	HandlerFc  http.HandlerFunc
	Methods    string
}

/*****************
	REST Server
 *****************/
func RunRestServer(routes []Route, pathPrefix string, cors bool, restConfig ServerConfig) {
	server := NewDefaultServer(routes, restConfig)
	err := server.Listen(pathPrefix, cors)
	if err != nil {
		log.Fatal(err)
	}
	err = server.Shutdown(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func NewDefaultServer(routes []Route, config ServerConfig) *RestServer {
	s := RestServer{}
	log.Printf("Starting mux router with tls = %v", config.CertFile != "" && config.KeyFile != "")
	s.r = mux.NewRouter()
	for _, item := range routes {
		log.Printf("adding route %s ( %s)", item.Path, item.Methods)
		if item.Path != "" {
			s.r.Path(item.Path).HandlerFunc(item.HandlerFc).Methods(strings.Split(item.Methods, ",")...)
		} else if item.PathPrefix != "" {
			s.r.PathPrefix(item.PathPrefix).HandlerFunc(item.HandlerFc).Methods(strings.Split(item.Methods, ",")...)
		}
	}
	log.Println("Starting listener")
	s.config = config
	return &s
}

func (s *RestServer) Listen(pathPrefix string, corsAllowed bool) error {

	var handler http.Handler
	handler = s.r
	if pathPrefix != "" {
		handler = http.StripPrefix(pathPrefix, handler)
		log.Printf("path prefix = %s", pathPrefix)
	}
	if corsAllowed {
		log.Println("enable cors")
		c := cors.New(cors.Options{
			AllowedOrigins:   []string{"http://localhost:8081"},
			AllowCredentials: true,
			AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE"},
		})
		handler = c.Handler(handler)
	}

	s.srv = &http.Server{
		Addr:         s.config.Listen,
		Handler:      handler,
		TLSConfig:    s.config.TlsConfig,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	if s.config.TlsConfig == nil || s.config.CertFile == "" || s.config.KeyFile == "" {
		return s.srv.ListenAndServe()
	} else {
		return s.srv.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile)
	}
}

func (s RestServer) Shutdown(ctx context.Context) error {
	s.srv.Shutdown(ctx)
	return nil
}
