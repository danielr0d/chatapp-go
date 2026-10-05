package httpserver

import (
	"log"
	"net/http"

	"chatapp/pkg/ws"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func NewRouter() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}).Methods(http.MethodGet)

	r.HandleFunc("/register", registerHandler).Methods(http.MethodPost)
	r.HandleFunc("/login", loginHandler).Methods(http.MethodPost)
	r.HandleFunc("/verify-contact", verifyContactHandler).Methods(http.MethodPost)
	r.HandleFunc("/chat-history", chatHistoryHandler).Methods(http.MethodGet)
	r.HandleFunc("/contact-list", contactListHandler).Methods(http.MethodGet)
	r.HandleFunc("/ws", ws.ServeWs)

	return cors.Default().Handler(r)
}

func StartHTTPServer(addr string) {
	log.Println("http server listening on", addr)
	log.Fatal(http.ListenAndServe(addr, NewRouter()))
}
