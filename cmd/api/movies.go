package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/hadnguyen/greenlight/internal/data"
	"github.com/hadnguyen/greenlight/internal/validator"
)

type movieCreateInput struct {
	Title   string       `json:"title" binding:"required" example:"The Breakfast Club"`
	Year    int32        `json:"year" binding:"required" example:"1985"`
	Runtime data.Runtime `json:"runtime" binding:"required" swaggertype:"string" example:"97 mins"`
	Genres  []string     `json:"genres" binding:"required" example:"comedy,drama"`
}

type movieUpdateInput struct {
	Title   *string       `json:"title" example:"The Breakfast Club"`
	Year    *int32        `json:"year" example:"1985"`
	Runtime *data.Runtime `json:"runtime" swaggertype:"string" example:"97 mins"`
	Genres  []string      `json:"genres" example:"comedy,drama"`
}

// listMoviesHandler godoc
// @Summary List movies
// @Tags Movies
// @Produce json
// @Description Returns a list of movies.
// @Param title query string false "Title filter"
// @Param genres query []string false "Genre filters" collectionFormat(csv)
// @Param page query int false "Page number" default(1) minimum(1) maximum(10000000)
// @Param page_size query int false "Page size" default(20) minimum(1) maximum(100)
// @Param sort query string false "Sort order" default(id) Enums(id, title, year, runtime, -id, -title, -year, -runtime)
// @Success 200 {object} object{movies=[]data.Movie}
// @Failure 422 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /movies [get]
func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	movies, metadata, err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	response := struct {
		Movies   []*data.Movie `json:"movies"`
		Metadata data.Metadata `json:"metadata"`
	}{
		Movies:   movies,
		Metadata: metadata,
	}

	err = app.writeJSON(w, http.StatusOK, response, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// createMovieHandler godoc
// @Summary Create a movie
// @Tags Movies
// @Produce json
// @Accept json
// @Param movie body movieCreateInput true "Movie details"
// @Failure 400 {object} map[string]interface{}
// @Failure 422 {object} map[string]interface{}
// @Success 201 {object} object{movie=data.Movie}
// @Failure 500 {object} map[string]interface{}
// @Router /movies [post]
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input movieCreateInput

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// showMovieHandler godoc
// @Summary Get a movie
// @Tags Movies
// @Produce json
// @Param id path int true "Movie ID" minimum(1)
// @Failure 404 {object} map[string]interface{}
// @Success 200 {object} object{movie=data.Movie}
// @Failure 500 {object} map[string]interface{}
// @Router /movies/{id} [get]
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// updateMovieHandler godoc
// @Summary Update a movie
// @Tags Movies
// @Produce json
// @Param id path int true "Movie ID" minimum(1)
// @Failure 404 {object} map[string]interface{}
// @Accept json
// @Param movie body movieUpdateInput true "Movie details"
// @Failure 400 {object} map[string]interface{}
// @Failure 422 {object} map[string]interface{}
// @Description Only provided fields are updated.
// @Failure 409 {object} map[string]interface{}
// @Success 200 {object} object{movie=data.Movie}
// @Failure 500 {object} map[string]interface{}
// @Router /movies/{id} [patch]
func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input movieUpdateInput

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		movie.Title = *input.Title
	}

	if input.Year != nil {
		movie.Year = *input.Year
	}

	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}

	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// deleteMovieHandler godoc
// @Summary Delete a movie
// @Tags Movies
// @Produce json
// @Param id path int true "Movie ID" minimum(1)
// @Failure 404 {object} map[string]interface{}
// @Success 200 {object} object{message=string}
// @Failure 500 {object} map[string]interface{}
// @Router /movies/{id} [delete]
func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
