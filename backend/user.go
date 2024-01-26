package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func RestUserAPI(router *mux.Router, database Database) {
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Must use GET", http.StatusMethodNotAllowed)
			return
		}
		user, err := CheckCookie(r, database)
		if err != nil {
			http.Error(w, "could not check cookie", http.StatusInternalServerError)
			return
		}
		if user == nil {
			user = &User{
				UserID: rand.Uint64(),
			}
			if err := database.CreateUser(*user); err != nil {
				http.Error(w, "could not create user", http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:     "userID",
				Value:    strconv.FormatUint(user.UserID, 10),
				MaxAge:   365 * 24 * 3600,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
				HttpOnly: true,
				Path:     "/",
			})
		}
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	})
}

func CheckCookie(r *http.Request, database Database) (*User, error) {
	cookie, err := r.Cookie("userID")
	if err != nil {
		return nil, nil
	}
	userID, err := strconv.ParseUint(cookie.Value, 10, 64)
	if err != nil {
		return nil, nil
	}
	user, err := database.ReadUser(userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// If returns nil, an error has been sent and must return from handler.
func Authenticate(w http.ResponseWriter, r *http.Request, database Database) *User {
	user, err := CheckCookie(r, database)
	if err != nil {
		http.Error(w, "invalid cookie", http.StatusInternalServerError)
		return nil
	}
	if user == nil {
		http.Error(w, "missing cookie or invalid user", http.StatusUnauthorized)
	}
	return user
}