package main
import (
	"errors"
	"net/http"
	"time"
)

// jsonResponse is the type used for generic JSON responses
type jsonResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type envelope map[string]interface{}

// Login is the handler used to attempt to log a user into the api
func (app *application) Login(w http.ResponseWriter, r *http.Request) {
	type credentials struct {
		UserName string `json:"email"`
		Password string `json:"password"`
	}

	var creds credentials
	var payload jsonResponse

	err := app.readJSON(w, r, &creds)
	if err != nil {
		app.errorLog.Println(err)
		payload.Error = true
		payload.Message = "invalid json supplied, or json missing entirely"
		_ = app.writeJSON(w, http.StatusBadRequest, payload)
	}

	// TODO authenticate
	

	// look up the user by email
	user, err := app.models.Student.GetByEmail(creds.UserName)
	if err != nil {
		app.errorJSON(w, errors.New("user does not exist"))
		return
	}

	// validate the user's password
	validPassword, err := user.PasswordMatches(creds.Password)
	if err != nil || !validPassword {
		app.errorJSON(w, errors.New("invalid username/password"))
		return
	}

	app.infoLog.Println(creds.UserName , " is logged in successfully")

	// we have a valid user, so generate a token
	token, err := app.models.Token.GenerateToken(*user, 24*time.Hour)
	if err != nil {
		app.errorJSON(w, err)
		return
	}


	// save it to the database
	err = app.models.Token.Insert(*token, *user)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.infoLog.Println("Token generated successfully")

	// send back a response
	payload = jsonResponse{
		Error: false,
		Message: "logged in",
		Data: envelope{"token": token, "user":user},
	}

	err = app.writeJSON(w, http.StatusOK, payload)
	if err != nil {
		app.errorLog.Println(err)
	}
}

// Logout will log the user out and delete the token associated with it
func (app *application) Logout(w http.ResponseWriter, r *http.Request) {
		var requestPayload struct  {
			Token string `json:"token"`
		}
		err:= app.readJSON(w,r,&requestPayload)
		if err!= nil{
			app.errorJSON(w, errors.New("invalid json "))
			return
		}

		err=app.models.Token.DeleteByToken(requestPayload.Token)
		if err != nil{
			app.errorJSON(w, errors.New("invalid json recieved"))
			return
		}
		
		payload:= jsonResponse {
			Error: false,
			Message: "logged out",
		}
		_= app.writeJSON(w,http.StatusOK,payload)
}

// GetCourses will get all the registered, feedbacks, and filled feedbacks
func (app *application) GetCoursesData(w http.ResponseWriter, r *http.Request) {
		id:=2

		data,err:=app.models.Student.GetCourses(id)
		if err != nil{
			app.errorJSON(w, err)
			return
		}
		
		payload:= jsonResponse {
			Error: false,
			Message: "successfully get the data",
			Data: envelope{"registred_courses":data.RegisteredCourses,"feedbacks":data.Feedbacks, "filled_feedbacks":data.FilledFeedbacks},
		}
		_= app.writeJSON(w,http.StatusOK,payload)
}

