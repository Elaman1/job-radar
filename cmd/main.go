package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

//import (
//	"context"
//	"errors"
//	"fmt"
//	"job-radar/internal/config"
//	"job-radar/internal/databases/postgres"
//	apihttp "job-radar/internal/transport/http"
//	"job-radar/internal/transport/http/handlers"
//	"job-radar/logger"
//	"log"
//	"net/http"
//	"os/signal"
//	"syscall"
//	"time"
//)

func main() {
	apiKey := "ecf7326dcb0221bd42d91e218f614f6e"

	userIP, err := getPublicIP()
	if err != nil {
		panic(err)
	}

	fmt.Println("public ip:", userIP)

	params := url.Values{}

	params.Set("locale_code", "en_GB")
	params.Set("keywords", "golang")
	params.Set("page", "1")
	params.Set("page_size", "100")
	params.Set("sort", "date")

	params.Set("user_ip", userIP)
	params.Set("user_agent", "Mozilla/5.0")

	endpoint := "https://search.api.careerjet.net/v4/query?" + params.Encode()

	ctx := context.Background()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		panic(err)
	}

	req.SetBasicAuth(apiKey, "")

	req.Header.Set(
		"Referer",
		"https://job-radar-2u8b.onrender.com/",
	)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("status:", resp.StatusCode)
	fmt.Println(string(body))
	fmt.Println("endpoint:", endpoint)
	fmt.Println("authorization:", req.Header.Get("Authorization"))
	fmt.Println("referer:", req.Header.Get("Referer"))

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

func getPublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
