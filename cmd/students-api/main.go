// main.go is where the program starts.
//
// It lives in cmd/students-api/ (not the project root). This is a common Go
// layout: cmd/<app-name>/main.go for each runnable program, and the rest of
// the code in internal/.
package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/students-api/internal/config"
	"example.com/students-api/internal/http/handlers/student"
	"example.com/students-api/internal/storage/sqlite"
)

func main() {
	// load config
	cfg := config.MustLoad()
	// database setup

	storage, err := sqlite.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("storage initialized", slog.String("env", cfg.Env), slog.String("version", "1.0.0"))

	// setup router
	router := http.NewServeMux()

	// student.New(storage) is CALLED here and returns the real handler
	// function (see the closure explanation in student.go).
	router.HandleFunc("POST /api/students", student.New(storage))
	router.HandleFunc("GET /api/students/{id}", student.GetById(storage))
	router.HandleFunc("GET /api/students", student.GetList(storage))
	// setup server

	// An http.Server struct gives more control than http.ListenAndServe,
	// and it is needed for Shutdown() below.
	server := http.Server{
		Addr:    cfg.Addr, // promoted field from the embedded HTTPServer
		Handler: router,
	}

	slog.Info("server started", slog.String("address", cfg.Addr))

	// ---- GRACEFUL SHUTDOWN ----
	// A CHANNEL is a pipe that goroutines use to send values to each other.
	// This one carries OS signals. Buffer size 1 so no signal is missed.
	done := make(chan os.Signal, 1)

	// "When the user presses Ctrl+C (SIGINT) or the system asks us to stop
	// (SIGTERM), send that signal into the done channel."
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// `go func() {...}()` starts a GOROUTINE: this function runs in the
	// background, so main() can continue to the next line.
	go func() {
		err := server.ListenAndServe()

		// FIX: after server.Shutdown(), ListenAndServe ALWAYS returns
		// http.ErrServerClosed. That is the normal "I stopped" signal, not a
		// failure, so we must not call log.Fatal for it. log.Fatal would exit
		// the whole program before the shutdown below could finish.
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("failed to start server")
		}
	}()

	// <-done BLOCKS (waits) until a value arrives in the channel,
	// i.e. until Ctrl+C. The server keeps running meanwhile.
	<-done

	slog.Info("shutting down the server")

	// A CONTEXT with a 5-second deadline: Shutdown gets at most 5 seconds.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // free the context's resources when main ends

	// Shutdown stops accepting new requests and waits for the running ones
	// to finish (or for the 5 seconds to run out).
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("failed to shutdown server", slog.String("error", err.Error()))
	}

	slog.Info("server shutdown successfully")

}
