package http

import (
	"net/http"

	"time"

	"crypto/tls"
	"net"

	"sync"

	"github.com/Tecsisa/foulkon/api"
	"github.com/Tecsisa/foulkon/foulkon"
	"github.com/julienschmidt/httprouter"
	"github.com/kylelemons/godebug/pretty"
)

type ReloadHandlerFunc func(watch *ProxyServer)

// ProxyServer struct with reload Handler extension
type ProxyServer struct {
	certFile string
	keyFile  string

	resourceLock sync.Mutex
	reloadFunc   ReloadHandlerFunc
	refreshTime  time.Duration

	reloadServe      chan struct{}
	currentResources []api.ProxyResource
	http.Server
}

// WorkerServer struct
type WorkerServer struct {
	certFile string
	keyFile  string

	http.Server
}

// Server interface that WorkerServer and ProxyServer have to implement
type Server interface {
	Run() error
	Configuration() error
}

// Run starts an HTTP WorkerServer
func (ws *WorkerServer) Run() error {
	var err error
	if ws.certFile != "" || ws.keyFile != "" {
		err = ws.ListenAndServeTLS(ws.certFile, ws.keyFile)
	} else {
		err = ws.ListenAndServe()
	}

	return err
}

// Configuration an HTTP ProxyServer with a given address
func (ps *ProxyServer) Configuration() error {
	if ps.certFile != "" || ps.keyFile != "" {
		if ps.Addr == "" {
			ps.Addr = ":https"
		}

		if !strSliceContains(ps.TLSConfig.NextProtos, "http/1.1") {
			ps.TLSConfig.NextProtos = append(ps.TLSConfig.NextProtos, "http/1.1")
		}

		configHasCert := len(ps.TLSConfig.Certificates) > 0 || ps.TLSConfig.GetCertificate != nil

		if !configHasCert || ps.certFile != "" || ps.keyFile != "" {
			var err error
			ps.TLSConfig.Certificates = make([]tls.Certificate, 1)
			ps.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(ps.certFile, ps.keyFile)
			if err != nil {
				return err
			}
		}
	}

	if ps.Addr == "" {
		ps.Addr = ":http"
	}

	return nil
}

// Configuration an HTTP WorkerServer
func (ws *WorkerServer) Configuration() error { return nil }

// Run starts an HTTP ProxyServer
func (ps *ProxyServer) Run() error {
	// Call reloadFunc every refreshTime
	timer := time.NewTicker(ps.refreshTime)
	// now wait for the other times when we needed to
	go func() {
		for range timer.C {
			// change the handler
			ps.reloadFunc(ps)
			ps.reloadServe <- struct{}{} // reset the listening binding
		}
	}()

	var err error
	ln, err := net.Listen("tcp", ps.Addr)
	if err != nil {
		return err
	}

	for {
		l := ln.(*net.TCPListener)
		defer l.Close()
		go func(l net.Listener) {
			err = ps.Serve(l)
		}(l)
		if err != nil {
			return err
		}
		<-ps.reloadServe
	}
}

// NewProxy returns a new ProxyServer
func NewProxy(proxy *foulkon.Proxy) Server {
	// Initialization
	ps := new(ProxyServer)
	ps.reloadServe = make(chan struct{}, 1)
	ps.TLSConfig = &tls.Config{}

	// Set Proxy parameters
	ps.certFile = proxy.CertFile
	ps.keyFile = proxy.KeyFile

	ps.Addr = proxy.Host + ":" + proxy.Port
	ps.refreshTime = proxy.RefreshTime
	ps.reloadFunc = ps.RefreshResources(proxy)

	ps.reloadFunc(ps)

	return ps
}

// NewWorker returns a new WorkerServer
func NewWorker(worker *foulkon.Worker, h http.Handler) Server {
	ws := new(WorkerServer)
	ws.certFile = worker.CertFile
	ws.keyFile = worker.KeyFile
	ws.Addr = worker.Host + ":" + worker.Port

	ws.Handler = h

	return ws
}

// RefreshResources implements reloadFunc
func (ps *ProxyServer) RefreshResources(proxy *foulkon.Proxy) func(s *ProxyServer) {
	return func(srv *ProxyServer) {
		router := httprouter.New()

		proxyHandler := ProxyHandler{proxy: proxy, client: http.DefaultClient}

		// Get proxy resources
		result, err := proxy.ProxyApi.GetProxyResources()
		if err != nil {
			proxy.Logger.Errorf("Unexpected error reading proxy resources from database %v", err)
			return
		}

		if diff := pretty.Compare(srv.currentResources, result); diff != "" {

			defer srv.resourceLock.Unlock()
			srv.resourceLock.Lock()
			// writer lock
			ps.currentResources = result

			proxy.Logger.Info("Updating resources ...")
			for _, res := range result {
				// Check if Url is empty. Handle doesn't accept empty path
				if res.Url == "" {
					res.Url = "/"
				}
				router.Handle(res.Method, res.Url, proxyHandler.HandleRequest(res))
			}
			ps.Server.Handler = router
		}
	}
}

func strSliceContains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}