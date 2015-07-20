package main

import (
	"fmt"
	"github.com/jinzhu/now"
	"github.com/parnurzeal/gorequest"
	"github.com/robfig/cron"
	"log"
	"os"
	"strconv"
	"time"
)

const URL = "http://1337online.com/manager.php"
const TimeFormat = "2006-01-02 15:04:05.000000"
const TimeZone = "Europe/Amsterdam"

type requestBody struct {
	Action string `json:"action"`
	Data   string `json:"data"`
}

func main() {
	manualCorrection, _ := strconv.Atoi(os.Getenv("CORRECTION"))

	c := cron.New()
	c.AddFunc("0 35 13 * * *", func() {
		fmt.Println("Determine timeout")
		duration, err := determineTimeout(200, time.Millisecond*50)
		if err != nil {
			log.Fatal(err)
			return
		}
		startTime := now.MustParse("13:37:00").Add(-(*duration)).Add(time.Millisecond * time.Duration(manualCorrection))
		<-time.After(startTime.Sub(time.Now()))
		gorequest.New().Post(URL).Type("form").Send(requestBody{
			Action: "new",
			Data:   os.Getenv("USERNAME"),
		}).End()
		fmt.Println("Posted!")
	})
	c.Start()
	fmt.Printf("1337 Bot cron started at %s\n", time.Now())

	// Never quit...
	select {}
}

func getServerTime() (*time.Time, error) {
	_, body, errors := gorequest.New().Post(URL).Type("form").Send(requestBody{
		Action: "getServerTime",
	}).End()

	if len(errors) > 0 {
		return nil, fmt.Errorf("Error in communication with server")
	}
	loc, _ := time.LoadLocation(TimeZone)

	serverTime, err := time.ParseInLocation(TimeFormat, body, loc)
	return &serverTime, err
}

func determineTimeout(samples int, interval time.Duration) (*time.Duration, error) {
	var minDuration *time.Duration
	for i := 0; i < samples; i++ {
		<-time.After(interval)
		start := time.Now()
		serverTime, err := getServerTime()
		if err != nil {
			return nil, err
		}
		duration := serverTime.Sub(start)
		if minDuration == nil || duration < *minDuration {
			minDuration = &duration
		}
	}

	return minDuration, nil
}
