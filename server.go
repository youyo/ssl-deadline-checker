package main

import (
	"crypto/tls"
	"math"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocraft/dbr"
	"github.com/labstack/echo"
)

// request & response
type (
	RegisterRequest struct {
		Hostname string `json:"hostname"`
	}

	HostData struct {
		ID            int    `json:"id"`
		Hostname      string `json:"hostname"`
		Timelimit     string `json:"timelimit"`
		RemainingDays int    `json:"remaining_days"`
		CreatedAt     string `json:"created_at"`
		UpdatedAt     string `json:"updated_at"`
	}

	ShowHosts []HostData

	Response struct {
		Response interface{} `json:"response"`
		Error    error       `json:"error"`
	}
)

// database
type Hostnames struct {
	ID            int    `db:"id"`
	Hostname      string `db:"hostname"`
	Timelimit     string `db:"timelimit"`
	RemainingDays int    `db:"remaining_days"`
	CreatedAt     string `db:"created_at"`
	UpdatedAt     string `db:"updated_at"`
}

const hostnames = "hostnames"

func main() {
	// initialilze
	e := echo.New()

	// routing
	g := e.Group("/ssl-deadline")

	// show all hosts
	g.GET("/", func(c echo.Context) error { return showAllHosts(c) })
	g.GET("", func(c echo.Context) error { return showAllHosts(c) })

	// register host
	g.POST("/", func(c echo.Context) error { return registerHost(c) })
	g.POST("", func(c echo.Context) error { return registerHost(c) })

	// show specific host
	g.GET("/:hostname", func(c echo.Context) error { return showSpecificHosts(c) })

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
	r := new(RegisterRequest)
	if err = c.Bind(r); err != nil {
		return c.JSON(http.StatusBadRequest, Response{Response: nil, Error: err})
	}

	// check cert
	timeLimit, remainingDays, err := checkCertLimit(r.Hostname)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Response: nil, Error: err})
	}

	// insert database
	sess, err := connectDb()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Response: nil, Error: err})
	}
	_, err = sess.InsertInto(hostnames).
		Columns("hostname", "timelimit", "remaining_days").
		Values(r.Hostname, timeLimit, remainingDays).
		Exec()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Response: nil, Error: err})
	}
	type status struct {
		Status bool `json:"status"`
	}
	return c.JSON(http.StatusCreated, Response{Response: status{true}, Error: err})
}

func checkCertLimit(hostname string) (timeLimit string, remainingDays int, err error) {
	conn, err := tls.Dial("tcp", hostname+":443", &tls.Config{})
	defer conn.Close()
	if err != nil {
		return
	}
	certs := conn.ConnectionState().PeerCertificates
	jst, _ := time.LoadLocation("Asia/Tokyo")
	t := certs[0].NotAfter.In(jst)
	timeLimit = t.Format("2006-01-02")
	duration := t.Sub(time.Now())
	remainingDays64 := math.Floor(duration.Hours() / 24)
	remainingDays = int(remainingDays64)
	return
}

func showAllHosts(c echo.Context) (err error) {
	// select database
	sess, err := connectDb()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Response: nil, Error: err})
	}
	var s ShowHosts
	_, err = sess.Select("*").From("hostnames").Load(&s)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{Response: nil, Error: err})
	}
	return c.JSON(http.StatusOK, Response{Response: s, Error: err})
}

func showSpecificHosts(c echo.Context) (err error) {
	// select database
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
