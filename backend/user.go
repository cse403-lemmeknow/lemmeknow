package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// User sent over JSON.
type GetUserResponse struct {
	UserID UserID    `json:"userId"`
	Name   string    `json:"name"`
	Status string    `json:"status"`
	Groups []GroupID `json:"groups"`
}

// User edit sent over JSON.
type PatchUserRequest struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type SubjectUserKeyType struct{}

// Used for looking up user, the subject of a specific request, out of request context.
var SubjectUserKey = SubjectUserKeyType(struct{}{})

// API's related to a specific group.
func RestSpecificUserAPI(router *mux.Router, database Database) {
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := ParseUint64PathParameter(w, r, "userID")
			if !ok {
				return
			}
			subjectUser, err := database.ReadUser(userID)
			if err != nil {
				http.Error(w, "could not read user", http.StatusInternalServerError)
				return
			}
			if subjectUser == nil {
				http.Error(w, "no such user", http.StatusNotFound)
				return
			}
			rWithContext := r.WithContext(context.WithValue(r.Context(), SubjectUserKey, subjectUser))
			next.ServeHTTP(w, rWithContext)
		})
	})
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		subjectUser := r.Context().Value(SubjectUserKey).(*User)
		switch r.Method {
		case http.MethodGet:
			WriteJSON(w, GetUserResponse{
				UserID: subjectUser.UserID,
				Name:   subjectUser.Name,
				// omit Groups: subjectUser.Groups for privacy!
				Status: subjectUser.Status,
			})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

// User-related API's.
func RestUserAPI(router *mux.Router, database Database, notification Notification) {
	RestSpecificUserAPI(AddHandler(router, "/{userID}"), database)
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		user, err := CheckCookie(r, database)
		if err != nil {
			http.Error(w, "could not check cookie", http.StatusInternalServerError)
			return
		}
		switch r.Method {
		case http.MethodGet:
			if user == nil {
				user = &User{
					UserID: GenerateID(),
				}
				if err := database.CreateUser(*user); err != nil {
					http.Error(w, "could not create user", http.StatusInternalServerError)
					return
				}
				http.SetCookie(w, &http.Cookie{
					Name:     "userID",
					Value:    strconv.FormatUint(user.UserID, 10),
					MaxAge:   365 * 24 * 3600,
					Secure:   isOnLambda(),
					SameSite: http.SameSiteStrictMode,
					HttpOnly: true,
					Path:     "/",
				})
			}
			WriteJSON(w, GetUserResponse{
				UserID: user.UserID,
				Name:   user.Name,
				Groups: append([]GroupID{}, user.Groups...),
				Status: user.Status,
			})
		case http.MethodPatch:
			var request PatchUserRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "could not decode body", http.StatusBadRequest)
				return
			}
			if err := updateUserAndNotifyGroups(user.UserID, func(user *User) error {
				if request.Name != "" {
					user.Name = request.Name
				}
				if request.Status != "" {
					user.Status = request.Status
				}
				return nil
			}, database, notification); err != nil {
				http.Error(w, "could not update user", http.StatusInternalServerError)
				return
			}
			WriteJSON(w, nil)
		default:
			http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		}
	})
}

// Gets possibly-nil user from cookie. Only possible error is database error.
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

type UserKeyType struct{}

// Used for looking up user out of request context.
var UserKey = UserKeyType(struct{}{})

// Adds user to request context or returns error if it doesn't exist.
func AuthenticateMiddleware(database Database) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := Authenticate(w, r, database)
			if user == nil {
				return
			}
			rWithContext := r.WithContext(context.WithValue(r.Context(), UserKey, user))
			next.ServeHTTP(w, rWithContext)
		})
	}
}
