package main

import (
	"errors"
	"net/http"

	"github.com/hadnguyen/greenlight/internal/data"
	"github.com/hadnguyen/greenlight/internal/validator"
)

type registerUserInput struct {
	Name     string `json:"name" binding:"required" example:"Jane Doe"`
	Email    string `json:"email" binding:"required" format:"email" example:"jane@example.com"`
	Password string `json:"password" binding:"required" example:"example-password"`
}

// registerUserHandler godoc
// @Summary Register a user
// @Description Creates an inactive user. Name must be at most 500 bytes; password must be between 8 and 72 bytes. Duplicate email addresses return a validation error.
// @Tags Users
// @Accept json
// @Produce json
// @Param user body registerUserInput true "User registration details"
// @Success 201 {object} object{user=data.User}
// @Failure 400 {object} map[string]interface{}
// @Failure 422 {object} map[string]interface{}
// @Failure 429 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input registerUserInput

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
