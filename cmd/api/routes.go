package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rashidalam9678/nba_management_system/internal/database"
)

func (app *application) routes() http.Handler{
	mux:= chi.NewRouter()
	mux.Use(middleware.Recoverer)
	mux.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins:   []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	  }))

	mux.Post("/users/login", app.Login)
	mux.Post("/users/logout", app.Logout)
	mux.Get("/user/get-courses/2", app.GetCoursesData)
	
	mux.Route("/admin",func(r chi.Router) {
		r.Use(app.AuthTokenMiddleware)

		mux.Post("/foo", func(w http.ResponseWriter, r *http.Request) {
			payload:= jsonResponse{
				Error: false,
				Message: "You are now in authenticated area",
			}
			_= app.writeJSON(w,http.StatusOK,payload)
		})
	})

	mux.Get("/users/all",func (w http.ResponseWriter, r *http.Request)  {
		var users database.Student
		all, err:= users.GetAll()
		if err!=nil{
			app.errorLog.Println(err)
			return
		}
		payload:=jsonResponse{
			Error:false,
			Message: "success",
			Data: envelope{"users":all},
		}
		app.writeJSON(w,http.StatusOK,payload)
	})
	mux.Get("/users/add",func(w http.ResponseWriter, r *http.Request) {
		var u =database.Student{
			Email:"admin@gmail.com",
			FirstName: "Mohd",
			LastName: "Rashid",
			Password: "mraauthentication",
		}
		fmt.Println("adding user .....")
		id,err:= app.models.Student.Insert(u)
		if err != nil{
			app.errorLog.Println(err)
			app.errorJSON(w,err,http.StatusForbidden)
			return
		}
		app.infoLog.Println("User inserted with id : ",id)
		 newUser, err:= app.models.Student.GetOne(id)

		 if err!= nil{
			app.infoLog.Println("no user found")
		 }
		 app.writeJSON(w,http.StatusOK,newUser)
		
	})


	

	return mux
}