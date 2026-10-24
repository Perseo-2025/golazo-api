package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/Perseo-2025/golazo-api/internal/handler"
)

func main(){

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.HealthCheck)

	listener, err := net.Listen("tcp", ":8080")

	if err != nil {
		log.Fatal(err) 
	}

	fmt.Println("Servidor escuchando en", listener.Addr())

	if err := http.Serve(listener, mux); err != nil {
		log.Fatal(err)
	}

}
