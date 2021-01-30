package main

import (
	"net/http"
	"strconv"

	"github.com/starshine-sys/berry/db"

	"github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func (a *api) list(c echo.Context) (err error) {
	// get all terms except the ones listed as "hidden"
	terms, err := a.db.GetTerms(db.FlagListHidden)
	if err != nil {
		// if no rows were returned, return no content
		if errors.Cause(err) == pgx.ErrNoRows {
			return c.NoContent(http.StatusNoContent)
		}
		// otherwise, internal server error
		return c.NoContent(http.StatusInternalServerError)
	}

	// if no rows were returned, return no content
	if len(terms) == 0 {
		return c.NoContent(http.StatusNoContent)
	}
	return c.JSON(http.StatusOK, terms)
}

func (a *api) listCategory(c echo.Context) (err error) {
	// parse the ID
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	// get all terms from that category
	terms, err := a.db.GetCategoryTerms(id, db.FlagListHidden)
	if err != nil {
		// if no rows were returned, return no content
		if errors.Cause(err) == pgx.ErrNoRows {
			return c.NoContent(http.StatusNoContent)
		}
		return c.NoContent(http.StatusInternalServerError)
	}

	// if no rows were returned, return no content
	if len(terms) == 0 {
		return c.NoContent(http.StatusNoContent)
	}
	return c.JSON(http.StatusOK, terms)
}

func (a *api) categories(c echo.Context) (err error) {
	// get all categories
	categories, err := a.db.GetCategories()
	if err != nil {
		// if no rows were returned, return no content
		if errors.Cause(err) == pgx.ErrNoRows {
			return c.NoContent(http.StatusNoContent)
		}
		return c.NoContent(http.StatusInternalServerError)
	}
	if len(categories) == 0 {
		return c.NoContent(http.StatusNoContent)
	}

	// if no rows were returned, return no content
	return c.JSON(http.StatusOK, categories)
}

func (a *api) explanations(c echo.Context) (err error) {
	// get all explanations
	explanations, err := a.db.GetAllExplanations()
	if err != nil {
		// if no rows were returned, return no content
		if errors.Cause(err) == pgx.ErrNoRows {
			return c.NoContent(http.StatusNoContent)
		}
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, explanations)
}
