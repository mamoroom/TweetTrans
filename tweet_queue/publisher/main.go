package main

import (
	//"github.com/mamoroom/tweet_translater/tweet_queue/publisher/lib/twitter_api_manager"
	"./lib"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	//_ "github.com/guregu/dynamo"
	"github.com/ChimeraCoder/anaconda"
)

type Conf struct {
	TwitterApiConf TwitterApiConf `json:"twitter_api"`
	GoogleApiConf  GoogleApiConf  `json:"google_api"`
}
type TwitterApiConf struct {
	ConsumerKey        string             `json:"consumer_key"`
	ConsumerSecret     string             `json:"consumer_secret"`
	AccessToken        string             `json:"access_token"`
	AccessTokenSecret  string             `json:"access_token_secret"`
	TwitterApiReqParam TwitterApiReqParam `json:"req_param"`
}

type TwitterApiReqParam struct {
	Track string `json:"track"`
}
type GoogleApiConf struct {
	ApiKey string `json:"api_key"`
}

func main() {
	// get conf
	var config Conf
	{
		file, err := ioutil.ReadFile("config.json")
		if err != nil {
			fmt.Fprint(os.Stderr, "Can't load json file")
			os.Exit(1)
		}
		err_json := json.Unmarshal(file, &config)
		if err_json != nil {
			fmt.Fprint(os.Stderr, "Fmt error of json")
			os.Exit(1)
		}
	}

	fmt.Println("Hello guys!")
	api_manager := twitter_api_manager.New(config.TwitterApiConf.ConsumerKey, config.TwitterApiConf.ConsumerSecret, config.TwitterApiConf.AccessToken, config.TwitterApiConf.AccessTokenSecret)
	go func() {
		for {
			stream := <-api_manager.StartPublicStream(config.TwitterApiConf.TwitterApiReqParam)
			switch tweet := stream.(type) {
			case anaconda.Tweet:
				fmt.Println("--------")
				rep := regexp.MustCompile(config.TwitterApiConf.TwitterApiReqParam.Track)
				_text := rep.ReplaceAllString(tweet.Text, "")
				fmt.Printf("%v:%s(%s): %s %s\n", tweet.Id, tweet.User.ScreenName, tweet.User.ProfileImageURL, _text, tweet.CreatedAt)
			case anaconda.StatusDeletionNotice:
				//pass
			default:
				fmt.Printf("unknown type(%T) : %v \n", tweet, tweet)
			}
		}
	}()
	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	exit_chan := make(chan int)
	go func() {
		for {
			s := <-signal_chan
			switch s {
			// kill -SIGHUP XXXX
			case syscall.SIGHUP:
				fmt.Println("hungup")
				exit_chan <- 0

			// kill -SIGINT XXXX or Ctrl+c
			case syscall.SIGINT:
				fmt.Println("Warikomi")
				api_manager.StopPublicStream()
				//exit_chan <- 0

			// kill -SIGTERM XXXX
			case syscall.SIGTERM:
				fmt.Println("force stop")
				exit_chan <- 0

			// kill -SIGQUIT XXXX
			case syscall.SIGQUIT:
				fmt.Println("stop and core dump")
				exit_chan <- 0

			default:
				fmt.Println("Unknown signal.")
				exit_chan <- 1
			}
		}
	}()

	code := <-exit_chan
	fmt.Println("Stop:", code)
	//os.Exit(code)
}
