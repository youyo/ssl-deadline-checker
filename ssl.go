package main

import (
	"crypto/tls"
	"math"
	"time"
)

func checkCertLimit(hostname string) (timeLimit string, remainingDays int, err error) {
	conn, err := tls.Dial("tcp", hostname+":443", &tls.Config{})
	if err != nil {
		return
	}
	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return
	}
	t := certs[0].NotAfter.In(jst)
	timeLimit = t.Format("2006-01-02")
	duration := t.Sub(time.Now())
	remainingDays64 := math.Floor(duration.Hours() / 24)
	remainingDays = int(remainingDays64)
	return
}
