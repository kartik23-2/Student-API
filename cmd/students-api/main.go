package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
    "context"
    "time"

	"github.com/kartik23-2/Student-API/internal/storage/sqlite"
    "github.com/kartik23-2/Student-API/internal/config"
	"github.com/kartik23-2/Student-API/internal/http/handlers/student"
)

func main() {
	//load config
	cfg := config.MustLoad()
	//store database

    _, err:=sqlite.New(cfg)
	if err!=nil{
		slog.Error("failed to initialize database", slog.String("error", err.Error()))
		return
	}
	

	//setup router
	router := http.NewServeMux()
	router.HandleFunc("POST /api/students", student.New(store))

	//setup server
	server:=http.Server{
		Addr: cfg.Address,
		Handler: router,
	}

	slog.Info("server started", slog.String("address", cfg.Address))

	done:=make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	
	go func(){
		err:= server.ListenAndServe()
		if err!=nil{
			log.Fatal("error occured")
		}
	}()

	<-done

	slog.Info("server stopped")

	ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second)
	defer cancel()

	if err:=server.Shutdown(ctx); err!=nil{
		slog.Error("failed to shutdown error", slog.String("error", err.Error()))
	}
	slog.Info("server stopped successfully")

	fmt.Println("hello")

	
}
