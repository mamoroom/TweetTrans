package twitter_api_manager

import (
	_ "fmt"
	"github.com/ChimeraCoder/anaconda"
	"net/url"
	"reflect"
	"strings"
)

type TwitterApiManager struct {
	TwitterApi anaconda.TwitterApi
	ApiStream  *anaconda.Stream
}

func New(client_key string, client_secret string, access_token string, access_token_secret string) TwitterApiManager {
	anaconda.SetConsumerKey(client_key)
	anaconda.SetConsumerSecret(client_secret)
	m := new(TwitterApiManager)
	m.TwitterApi = *anaconda.NewTwitterApi(access_token, access_token_secret)
	SetLogger(anaconda.BasicLogger)
	return *m
}

func (m *TwitterApiManager) StartPublicStream(req_param interface{}) chan interface{} {
	v_url := url.Values{}
	v := reflect.ValueOf(req_param)
	rt := v.Type()
	r := reflect.Indirect(v)
	for i := 0; i < rt.NumField(); i++ {
		name := rt.Field(i).Name
		v_url.Set(strings.ToLower(name), r.FieldByName(name).String())
	}
	m.ApiStream = m.TwitterApi.PublicStreamFilter(v_url)
	return m.ApiStream.C
}

func (m *TwitterApiManager) StopPublicStream() {
	m.ApiStream.Stop()
}
