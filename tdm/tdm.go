package tdm

import (
	"fmt"
	"net/url"

	"github.com/ChimeraCoder/anaconda"
)

type Tdm struct {
	api    *anaconda.TwitterApi
	stream *anaconda.Stream
}

func NewTdm(consumerKey, consumerSecret, accessToken, accessTokenSecret string) *Tdm {
	anaconda.SetConsumerKey(consumerKey)
	anaconda.SetConsumerSecret(consumerSecret)
	api := anaconda.NewTwitterApi(accessToken, accessTokenSecret)
	tdm := new(Tdm)
	tdm.api = api

	return tdm
}

func (tdm *Tdm) OpenStream() {
	v := url.Values{}
	v.Set("with", "user")
	tdm.stream = tdm.api.UserStream(v)
}

func (tdm *Tdm) CloseStream() {
	tdm.stream.Stop()
}

func (tdm *Tdm) GetDmStream() (chan interface{}, error) {
	if tdm.stream == nil {
		return nil, fmt.Errorf("No open stream")
	}

	return tdm.stream.C, nil
}

func (tdm *Tdm) GetDms() ([]anaconda.DirectMessage, error) {
	return tdm.api.GetDirectMessages(nil)
}
