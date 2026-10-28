package main

import (
	"net/http"
)

// healthcheckHandler godoc
// @Summary Check API health
// @Tags System
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /healthcheck [get]
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
