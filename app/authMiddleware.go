package app

import (
	"fmt"
	"qastack-workflows/domain"
	"qastack-workflows/errs"

	"github.com/gorilla/mux"

	"net/http"
	"strings"

	_ "github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type AuthMiddleware struct {
	repo domain.AuthRepository
}

// CORS Middleware
func (a AuthMiddleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Headers:", "*")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		currentRoute := mux.CurrentRoute(r)
		currentRouteVars := mux.Vars(r)
		authHeader := r.Header.Get("Authorization")
		log.Info(authHeader)
		log.Info("ok")
		if authHeader != "" {
			token := getTokenFromHeader(authHeader)
			log.Info("token", token)
			log.Info(currentRoute.GetName())
			log.Info(currentRouteVars)
			w.Header().Set("Access-Control-Allow-Headers:", "*")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "*")

			fmt.Println("ok")
			isAuthorized := a.repo.IsAuthorized(token, currentRoute.GetName(), currentRouteVars)
			log.Info("authorised:", isAuthorized)
			log.Info(r.Header.Get("Access-Control-Allow-Origin"))
			if isAuthorized {

				log.Info(r)
				next.ServeHTTP(w, r)
			} else {
				appError := errs.AppError{http.StatusForbidden, "Unauthorized"}
				WriteResponse(w, appError.Code, appError.AsMessage())
			}
		} else {
			WriteResponse(w, http.StatusUnauthorized, "missing token")
		}
		//// Set headers
		//w.Header().Set("Access-Control-Allow-Headers:", "*")
		//w.Header().Set("Access-Control-Allow-Origin", "*")
		//w.Header().Set("Access-Control-Allow-Methods", "*")
		//
		//if r.Method == "OPTIONS" {
		//	w.WriteHeader(http.StatusOK)
		//	return
		//}
		//
		//fmt.Println("ok")
		//
		//// Next
		//next.ServeHTTP(w, r)
		//return
	})
}

func (a AuthMiddleware) authorizationHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			currentRoute := mux.CurrentRoute(r)
			currentRouteVars := mux.Vars(r)
			authHeader := r.Header.Get("Authorization")

			if authHeader != "" {
				token := getTokenFromHeader(authHeader)
				log.Info("token", token)
				log.Info(currentRoute.GetName())
				log.Info(currentRouteVars)
				w.Header().Set("Access-Control-Allow-Headers:", "*")
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "*")

				fmt.Println("ok")
				isAuthorized := a.repo.IsAuthorized(token, currentRoute.GetName(), currentRouteVars)
				log.Info("authorised:", isAuthorized)
				log.Info(r.Header.Get("Access-Control-Allow-Origin"))
				if isAuthorized {

					log.Info(r)
					next.ServeHTTP(w, r)
				} else {
					appError := errs.AppError{http.StatusForbidden, "Unauthorized"}
					WriteResponse(w, appError.Code, appError.AsMessage())
				}
			} else {
				WriteResponse(w, http.StatusUnauthorized, "missing token")
			}
		})
	}
}

func getTokenFromHeader(header string) string {
	/*
	   token is coming in the format as below
	   "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50cyI6W.yI5NTQ3MCIsIjk1NDcyIiw"
	*/
	splitToken := strings.Split(header, "Bearer")
	if len(splitToken) == 2 {
		return strings.TrimSpace(splitToken[1])
	}
	return ""
}
