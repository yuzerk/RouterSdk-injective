// Package server provides JSON/RESTful RPC service.
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/libstring"
	"github.com/didip/tollbooth/v6/limiter"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
	rpcjson "github.com/gorilla/rpc/v2/json2"

	"github.com/anyswap/CrossChain-Router/v3/log"
	"github.com/anyswap/RouterSDK-injective/cmd/utils"
	"github.com/anyswap/RouterSDK-injective/config"
)

// StartAPIServer start api server
func StartAPIServer() {
	router := mux.NewRouter()
	initAPIRouter(router)

	addAuthenticationMiddleware(router)

	cfg := config.GetServerConfig()
	apiPort := cfg.Port
	allowedOrigins := cfg.AllowedOrigins
	maxRequestsLimit := cfg.MaxRequestsLimit
	if maxRequestsLimit <= 0 {
		maxRequestsLimit = 10 // default value
	}

	corsOptions := []handlers.CORSOption{
		handlers.AllowedMethods([]string{"GET", "POST"}),
	}
	if len(allowedOrigins) != 0 {
		corsOptions = append(corsOptions,
			handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type"}),
			handlers.AllowedOrigins(allowedOrigins),
		)
	}

	log.Info("JSON RPC service listen and serving", "port", apiPort, "allowedOrigins", allowedOrigins)
	lmt := tollbooth.NewLimiter(float64(maxRequestsLimit),
		&limiter.ExpirableOptions{
			DefaultExpirationTTL: 600 * time.Second,
		},
	)
	lmt.SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
		remoteIP := libstring.RemoteIP(lmt.GetIPLookups(), lmt.GetForwardedForIndexFromBehind(), r)
		remoteIP = libstring.CanonicalizeIP(remoteIP)
		log.Warnf("rpc limit reached: %v\n", remoteIP)
	})
	handler := tollbooth.LimitHandler(lmt, handlers.CORS(corsOptions...)(router))
	svr := http.Server{
		Addr:         fmt.Sprintf(":%v", apiPort),
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 300 * time.Second,
		Handler:      handler,
	}
	go func() {
		if err := svr.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) && utils.IsCleanuping() {
				return
			}
			log.Fatal("ListenAndServe error", "err", err)
		}
	}()

	utils.TopWaitGroup.Add(1)
	go utils.WaitAndCleanup(func() { doCleanup(&svr) })
}

func doCleanup(svr *http.Server) {
	defer utils.TopWaitGroup.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := svr.Shutdown(ctx); err != nil {
		log.Error("Server Shutdown failed", "err", err)
	}
	log.Info("Close http server success")
}

func initAPIRouter(r *mux.Router) {
	rpcserver := rpc.NewServer()
	rpcserver.RegisterCodec(rpcjson.NewCodec(), "application/json")
	err := rpcserver.RegisterService(new(ChainSupportAPI), "bridge")
	if err != nil {
		log.Fatal("start rpc service failed", "err", err)
	}

	r.Handle("/", rpcserver)
}
