package main

import (
	"html/template"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocraft/dbr"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// request & response
type (
	RegisterRequest struct {
		Hostname         string `json:"hostname"`
		NotificationDays int    `json:"notification_days"`
	}

	HostData struct {
		ID               int    `json:"id"`
		Hostname         string `json:"hostname"`
		Timelimit        string `json:"timelimit"`
		RemainingDays    int    `json:"remaining_days"`
		NotificationDays int    `json:"notification_days"`
		CreatedAt        string `json:"created_at"`
		UpdatedAt        string `json:"updated_at"`
	}

	ShowHosts []HostData

	Response struct {
		Response interface{} `json:"response"`
		Error    error       `json:"error"`
	}
)

// table name
const hostnames = "hostnames"

// hostnames table struct
type Hostnames struct {
	ID               int    `db:"id"`
	Hostname         string `db:"hostname"`
	Timelimit        string `db:"timelimit"`
	RemainingDays    int    `db:"remaining_days"`
	NotificationDays int    `db:"notification_days"`
	CreatedAt        string `db:"created_at"`
	UpdatedAt        string `db:"updated_at"`
}

func main() {
	// initialilze
	e := echo.New()

	// cors
	e.Use(middleware.CORS())

	// use static file
	e.Static("/", "public/assets")
	e.Renderer = &Template{
		templates: template.Must(template.ParseGlob("public/views/*.html")),
	}

	// view
	e.GET("/", index)

	// routing to api
	g := e.Group("/api")

	// show all hosts
	g.GET("/", func(c echo.Context) error { return showAllHosts(c) })
	g.GET("", func(c echo.Context) error { return showAllHosts(c) })

	// register host
	g.POST("/", func(c echo.Context) error { return registerHost(c) })
	g.POST("", func(c echo.Context) error { return registerHost(c) })

	// show specific host
	g.GET("/:hostname", func(c echo.Context) error { return showSpecificHosts(c) })

	// check deadline
	g.POST("/check/:hostname", func(c echo.Context) error { return checkDeadline(c) })

	// server start
	e.Logger.Fatal(e.Start(":1323"))
}

func connectDb() (sess *dbr.Session, err error) {
	// connection info
	username := os.Getenv("MYSQL_USER")
	password := os.Getenv("MYSQL_PASSWORD")
	host := os.Getenv("MYSQL_HOST")
	port := os.Getenv("MYSQL_PORT")
	database := os.Getenv("MYSQL_DATABASE")

	// try to connect
	dsn := username + ":" + password + "@tcp(" + host + ":" + port + ")/" + database
	conn, err := dbr.Open("mysql", dsn, nil)
	if err != nil {
		return nil, err
	}
	sess = conn.NewSession(nil)
	return sess, nil
}

func registerHost(c echo.Context) (err error) {
	// bind json
	r := &RegisterRequest{NotificationDays: 45}
	if err = c.Bind(r); err != nil {
		return c.JSON(http.StatusBadRequest, Response{Response: nil, Error: err})
	}

	// check cert
	timeLimit, remainingDays, err := checkCertLimit(r.Hostname)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{Response: nil, Error: err})
	}

	// insert database
	sess, err := connectDb()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Response: nil, Error: err})
	}
	_, err = sess.InsertInto(hostnames).
		Columns("hostname", "timelimit", "remaining_days", "notification_days").
		Values(r.Hostname, timeLimit, remainingDays, r.NotificationDays).
		Exec()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Response: nil, Error: err})
	}
	type status struct {
		Status bool `json:"status"`
	}
	return c.JSON(http.StatusCreated, Response{Response: status{true}, Error: err})
}

func showAllHosts(c echo.Context) (err error) {
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
	return c.JSON(http.StatusOK, Response{Response: s, Error: err})
}

func showSpecificHosts(c echo.Context) (err error) {
	// select from database
	sess, err := connectDb()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Response: nil, Error: err})
	}
	var d HostData
	hostname := c.Param("hostname")
	_, err = sess.Select("*").From("hostnames").Where("hostname=?", hostname).Load(&d)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Response: nil, Error: err})
	}
	return c.JSON(http.StatusOK, Response{Response: d, Error: err})
}

func updateQuery(hostname string) (err error) {
	// check cert
	timeLimit, remainingDays, err := checkCertLimit(hostname)
	if err != nil {
		return
	}

	// update
	sess, err := connectDb()
	if err != nil {
		return
	}
	_, err = sess.Update(hostnames).
		Set("timelimit", timeLimit).
		Set("remaining_days", remainingDays).
		Where("hostname = ?", hostname).
		Exec()
	if err != nil {
		return
	}
	sess, err = connectDb()
	if err != nil {
		return
	}
	var d HostData
	_, err = sess.Select("notification_days").From("hostnames").Where("hostname=?", hostname).Load(&d)
	if err != nil {
		return
	}
	if d.NotificationDays >= remainingDays {
		message := "https://" + hostname + "'s ssl deadline is " + timeLimit + ". " + strconv.Itoa(remainingDays) + " days left until the deadline."
		if err = notify(message); err != nil {
			return
		}
	}
	return nil
}

func checkDeadline(c echo.Context) (err error) {
	hostname := c.Param("hostname")
	return func() error {
		if hostname == "all" {
			sess, err := connectDb()
			if err != nil {
				return c.JSON(http.StatusInternalServerError, Response{Response: nil, Error: err})
			}
			var s ShowHosts
			_, err = sess.Select("hostname").
				From("hostnames").
				Load(&s)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, Response{Response: nil, Error: err})
			}
			for _, v := range s {
				if err = updateQuery(v.Hostname); err != nil {
					return c.JSON(http.StatusInternalServerError, Response{Response: nil, Error: err})
				}
			}
		} else {
			if err = updateQuery(hostname); err != nil {
				return c.JSON(http.StatusInternalServerError, Response{Response: nil, Error: err})
			}
		}
		return c.JSON(http.StatusOK, Response{Response: "ok", Error: err})
	}()
}
