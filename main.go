package main

import (
	"flag"
	"fmt"
	"github.com/avast/retry-go"
	"go.uber.org/zap"
	"log"
	"net/url"
	"os/exec"
	"time"
)

const RETRY_DELAY = 1 * time.Minute
const RETRY_MAX = 100

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	sugar := logger.Sugar()
	sugar.Info("Reserve youtube-d; until target time.")

	// setup flag
	flag.Parse()
	args := flag.Args()

	if len(args) < 2 {
		sugar.Info("usage: $reserve-dl \"2022-01-01 12:00\" \"https://www.youtube.com/user/dommune/live\"\n")
		return
	}

	reserveTime, err := parseTime(args[0])
	if err != nil {
		sugar.Error(err)
		return
	}
	sugar.Info(fmt.Printf("reserve time: %+v\n", reserveTime))

	targetUrl, err := url.Parse(args[1])
	if err != nil {
		sugar.Error(err)
		return
	}
	sugar.Info(fmt.Printf("target url: %+v\n", targetUrl.String()))

	// sleep until reserve time
	sugar.Info("sleep until: %v", reserveTime)
	sleep(reserveTime)

	// try download
	cmd := youtubeDL(targetUrl.String())
	err = retry.Do(func() error {
		sugar.Info(fmt.Sprintf("try at: %+v\n", time.Now()))
		result, err := cmd.Output()
		if err != nil {
			sugar.Error(err)
			return err
		}
		sugar.Info(fmt.Printf("youtube-dl result: %s\n", result))
		return nil
	},
		retry.Delay(RETRY_DELAY),
		retry.Attempts(RETRY_MAX),
	)
	if err != nil {
		sugar.Error(err)
	}

	sugar.Info("finish download: %s\n", targetUrl.String())
}

func sleep(targetTime time.Time) {
	for time.Now().Sub(targetTime) < 0 {
		time.Sleep(1 * time.Second)
	}
}

func parseTime(ts string) (time.Time, error) {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	layout := "2006-1-2 15:4"
	return time.ParseInLocation(layout, ts, loc)
}

func youtubeDL(url string) *exec.Cmd {
	return exec.Command("youtube-dl", url)
}
