package main

import (
	"html/template"
	"io"
	"net/http"

	"github.com/labstack/echo"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func index(c echo.Context) error {
	// select from database
	sess, err := connectDb()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Response: nil, Error: err})
	}
	var s ShowHosts
	_, err = sess.Select("*").From("hostnames").OrderBy("remaining_days").Load(&s)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Response: nil, Error: err})
	}
	return c.Render(http.StatusOK, "index", s)
}
