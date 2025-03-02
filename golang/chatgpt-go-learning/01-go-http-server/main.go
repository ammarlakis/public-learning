package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to %s!", r.Host)
}

func main() {
	srv := &http.Server{
		Addr:    ":8080",
		Handler: http.HandlerFunc(handler),
	}

	// Channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Println("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("Error starting server:", err)
		}

	}()

	<-stop
	fmt.Println("\nShutting down server...")

	// create a context with a timeout for cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Println("Error shutting down server:", err)
	}

	fmt.Println("Server gracefully stopped")

}
