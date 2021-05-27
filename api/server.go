package api

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/lucas-clemente/quic-go/http3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	rpcUtils "pm.tcfw.com.au/source/ataas/internal/utils/rpc"
)

type APIServer struct {
	router *mux.Router

	Stop func()
}

func NewAPIServer(ctx context.Context) (*APIServer, error) {
	mux := mux.NewRouter()

	apiR, err := newRouter(ctx)
	if err != nil {
		return nil, err
	}

	mux.PathPrefix("/").HandlerFunc(apiR.ServeHTTP)

	return &APIServer{
		router: mux,
	}, nil
}

func (s *APIServer) Serve() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error)
	defer close(errCh)

	go func(ctx context.Context) {
		err := s.serveHTTPS(ctx)
		if err != nil {
			errCh <- err
		}
	}(ctx)

	go func(ctx context.Context) {
		err := s.serveGRPC(ctx)
		if err != nil {
			errCh <- err
		}
	}(ctx)

	return <-errCh
}

func (s *APIServer) serveGRPC(ctx context.Context) error {
	lis, err := net.Listen("tcp", viper.GetString("grpc.addr"))
	if err != nil {
		return err
	}

	start, stop, grpc := newGRPCServer(ctx, rpcUtils.DefaultServerOptions()...)

	s.Stop = stop

	if viper.GetBool("services.start") {
		start()
	}

	return grpc.Serve(lis)
}

func (s *APIServer) serveHTTPS(ctx context.Context) error {
	httpServ := &http.Server{
		Addr:           viper.GetString("https.addr"),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	h3Serv := &http3.Server{Server: httpServ}

	s.router.Use(h3Headers(h3Serv))

	if viper.GetBool("gw.enableAuth") {
		s.router.Use(authHandler)
	}

	httpServ.Handler = s.router

	go func() {
		err := h3Serv.ListenAndServeTLS(viper.GetString("tls.cert"), viper.GetString("tls.key"))
		if err != nil {
			logrus.New().Errorf("[http3] %s", err)
		}
	}()

	err := httpServ.ListenAndServeTLS(viper.GetString("tls.cert"), viper.GetString("tls.key"))

	return err
}

func h3Headers(h3serv *http3.Server) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h3serv.SetQuicHeaders(w.Header())
			next.ServeHTTP(w, r)
		})
	}
}
