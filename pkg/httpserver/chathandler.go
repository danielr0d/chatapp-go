package httpserver

import (
	"encoding/json"
	"net/http"
	"regexp"

	"chatapp/pkg/redisrepo"
)

// usernames end up inside RediSearch TAG queries, so keep them to safe characters
var validUsername = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

type userReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type response struct {
	Status  bool        `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func writeJSON(w http.ResponseWriter, code int, res response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(res)
}

func decodeUser(w http.ResponseWriter, r *http.Request) (userReq, bool) {
	var u userReq
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Message: "invalid request body"})
		return u, false
	}
	return u, true
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	u, ok := decodeUser(w, r)
	if !ok {
		return
	}

	if !validUsername.MatchString(u.Username) || u.Password == "" {
		writeJSON(w, http.StatusBadRequest,
			response{Message: "username must be 3-32 letters, digits or _ and password is required"})
		return
	}

	if redisrepo.IsUserExist(u.Username) {
		writeJSON(w, http.StatusConflict, response{Message: "username already taken"})
		return
	}

	if err := redisrepo.RegisterNewUser(u.Username, u.Password); err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Message: "error while registering user"})
		return
	}

	writeJSON(w, http.StatusCreated, response{Status: true, Message: "user created"})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	u, ok := decodeUser(w, r)
	if !ok {
		return
	}

	if err := redisrepo.IsUserAuthentic(u.Username, u.Password); err != nil {
		writeJSON(w, http.StatusUnauthorized, response{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, response{Status: true, Message: "login successful"})
}

func verifyContactHandler(w http.ResponseWriter, r *http.Request) {
	u, ok := decodeUser(w, r)
	if !ok {
		return
	}

	if !redisrepo.IsUserExist(u.Username) {
		writeJSON(w, http.StatusNotFound, response{Message: "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, response{Status: true})
}

// chatHistoryHandler: GET /chat-history?u1=john&u2=jane&from-ts=0&to-ts=+inf
func chatHistoryHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	u1, u2 := q.Get("u1"), q.Get("u2")
	fromTS, toTS := q.Get("from-ts"), q.Get("to-ts")

	if !validUsername.MatchString(u1) || !validUsername.MatchString(u2) {
		writeJSON(w, http.StatusBadRequest, response{Message: "invalid username"})
		return
	}
	if fromTS == "" {
		fromTS = "0"
	}
	if toTS == "" {
		toTS = "+inf"
	}

	chats, err := redisrepo.FetchChatBetween(u1, u2, fromTS, toTS)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Message: "error while fetching chats"})
		return
	}

	writeJSON(w, http.StatusOK, response{Status: true, Data: chats})
}

// contactListHandler: GET /contact-list?username=john
func contactListHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")

	contacts, err := redisrepo.FetchContactList(username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, response{Message: "error while fetching contact list"})
		return
	}

	writeJSON(w, http.StatusOK, response{Status: true, Data: contacts})
}
