package main

import (
	//"github.com/mamoroom/tweet_translater/tweet_queue/publisher/lib/twitter_api_manager"
	"./lib/dynamo_db_manager"
	"./lib/tweet_api_manager"
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"html"
	"io/ioutil"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
	//_ "github.com/guregu/dynamo"
	"cloud.google.com/go/translate"
	"github.com/ChimeraCoder/anaconda"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
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

var TARGET_LANGUAGES = []string{"ja", "en", "es", "fr", "it", "de", "ko"}

func main() {
	// get conf
	var config Conf
	{
		file, err := ioutil.ReadFile("../../config.json")
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

	// init google translate api <- あとで引っ越し //
	ctx := context.Background()
	g_t_api, err := translate.NewClient(ctx, option.WithAPIKey(config.GoogleApiConf.ApiKey))
	if err != nil {
		fmt.Fprint(os.Stderr, "Fmt error of json")
		g_t_api.Close()
		os.Exit(1)
	}
	/////////////////////////////

	fmt.Println("Hello guys!")
	api_manager := twitter_api_manager.New(config.TwitterApiConf.ConsumerKey, config.TwitterApiConf.ConsumerSecret, config.TwitterApiConf.AccessToken, config.TwitterApiConf.AccessTokenSecret)
	twitter_stream := api_manager.GetPublicStream(config.TwitterApiConf.TwitterApiReqParam)
	go func() {
		for {
			stream := <-twitter_stream.C
			switch tweet := stream.(type) {
			case anaconda.Tweet:
				fmt.Println("--------")
				rep := regexp.MustCompile(config.TwitterApiConf.TwitterApiReqParam.Track)
				_text := rep.ReplaceAllString(tweet.Text, "")
				fmt.Printf("%v:%s(%s): %s %s\n", tweet.Id, tweet.User.ScreenName, tweet.User.ProfileImageURL, _text, tweet.CreatedAt)

				time_tweeted, err := tweet.CreatedAtTime()
				if err != nil {
					fmt.Println("%v", err)
					return
				}
				for _, target_lang := range TARGET_LANGUAGES {
					trans_text := get_translated_text(_text, target_lang, g_t_api, ctx)
					//fmt.Println(trans_text)

					ddm := dynamo_db_manager.New()
					boolean := ddm.PutTrans(target_lang, time_tweeted.Unix(), tweet.Id, tweet.User.ScreenName, _text, trans_text, tweet.User.ProfileImageURL)
					fmt.Println("->" + target_lang + ":" + strconv.FormatBool(boolean))
				}
			case anaconda.StatusDeletionNotice:
				//pass
			default:
				//fmt.Printf("unknown type(%T) : %v \n", tweet, tweet)
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
	is_can_exit := false
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
				fmt.Println("Ctl+c")
				if !is_can_exit {
					twitter_stream.Stop()
					is_can_exit = true
					fmt.Println("Now you can exit")
				} else {
					exit_chan <- 0
				}

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
	fmt.Println("Stop:")
	os.Exit(code)
}

func get_translated_text(text string, target_lang string, g_t_api *translate.Client, ctx context.Context) string {
	lang, err := language.Parse(target_lang)
	if err != nil {
		return "NaN"
	}
	resp, err := g_t_api.Translate(ctx, []string{text}, lang, nil)
	if err != nil {
		return "NaN"
	}

	return html.UnescapeString(resp[0].Text)
}
