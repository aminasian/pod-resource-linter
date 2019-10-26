package server

import (
	"context"
	"crypto/tls"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	// (https://github.com/kubernetes/kubernetes/issues/57982)
	defaulter = runtime.ObjectDefaulter(runtimeScheme)
)

type Env struct{}

func StartServer() {

	tlsCertPath := os.Getenv("TLS_CERT_PATH")
	tlsKeyPath := os.Getenv("TLS_KEY_PATH")

	pair, err := tls.LoadX509KeyPair(tlsCertPath, tlsKeyPath)
	if err != nil {
		log.Fatalf("Failed to load tls cert pair: %v", err)
	}

	srv := &http.Server{
		Addr:      ":8080",
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
	}

	e := Env{}

	mux := http.NewServeMux()
	mux.HandleFunc("/validate", e.handleValidationWebHookRequest)

	srv.Handler = mux

	go func() {
		log.Printf("Starting Http Server now")
		err := srv.ListenAndServeTLS("", "")
		if err != nil {
			log.Fatalf("Starting http server failed: %v", err)
		}
	}()

	log.Printf("Successfully started HTTP Server")

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Print("Got OS shutdown signal, shutting down webhook server gracefully...")
	srv.Shutdown(context.Background())
}
