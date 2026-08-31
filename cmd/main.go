package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Job Radar")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}

	//conf, err := config.InitConf()
	//if err != nil {
	//	log.Fatalf("Error load configuration %v", err)
	//	return
	//}
	//
	//ctx, stop := signal.NotifyContext(
	//	context.Background(),
	//	syscall.SIGINT,
	//	syscall.SIGTERM,
	//)
	//defer stop()
	//
	//newLogger := logger.NewLogger(conf.Logger)
	//database, err := postgres.NewPostgres(ctx, conf.Postgres)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//healthHandler := handlers.NewHealthHandler()
	//checkerHandler := handlers.NewCheckerHandler(database.Pool(), conf)
	//
	//router := apihttp.NewRouter(
	//	newLogger,
	//	healthHandler,
	//	checkerHandler,
	//)
	//
	//srv := http.Server{
	//	Addr:         conf.Server.Address,
	//	Handler:      router,
	//	ReadTimeout:  conf.Server.ReadTimeout,
	//	WriteTimeout: conf.Server.WriteTimeout,
	//}
	//
	//go func() {
	//	if errSrv := srv.ListenAndServe(); errSrv != nil && !errors.Is(errSrv, http.ErrServerClosed) {
	//		log.Fatal(errSrv)
	//	}
	//}()
	//
	//fmt.Println("Listening on " + conf.Server.Address)
	//<-ctx.Done()
	//
	//fmt.Println("Shutting down...")
	//
	//ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	//
	//if errShutdown := srv.Shutdown(ctxTimeout); errShutdown != nil {
	//	fmt.Println("Error shutting down http server ", errShutdown)
	//}
	//
	//fmt.Println("server shutdown")
}
