package tdm

import (
	"fmt"
	"net/url"
	"strconv"

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

func (tdm *Tdm) GetDms(p DmParams) ([]anaconda.DirectMessage, error) {
	return tdm.api.GetDirectMessages(p.ToValues())
}

func (tdm *Tdm) GetDm(id string) (anaconda.DirectMessage, error) {
	v := url.Values{}
	v.Set("id", id)

	dms, err := tdm.api.GetDirectMessagesShow(v)
	if err != nil || len(dms) != 1 {
		return anaconda.DirectMessage{}, err
	}

	return dms[0], nil
}

func (tdm *Tdm) GetSentDms(p DmParams) ([]anaconda.DirectMessage, error) {
	return tdm.api.GetDirectMessagesSent(p.ToValues())
}

func (tdm *Tdm) GetFriends() ([]anaconda.User, error) {
	// TODO: Cross-reference GetFollowersListAll and GetFriendsListAll
	// (and maybe info about followed users with open DMs)
	// to get a list of followed people avialable for DM

	return nil, nil
}

type DmParams struct {
	SinceId string
	MaxId   string
	Count   int
}

func (p *DmParams) ToValues() url.Values {
	v := url.Values{}
	if p.SinceId != "" {
		v.Set("since_id", p.SinceId)
	}
	if p.MaxId != "" {
		v.Set("max_id", p.MaxId)
	}
	if p.Count > 0 {
		v.Set("count", strconv.Itoa(p.Count))
	}

	return v
}

func NewDmParams(options ...DmOption) (*DmParams, error) {
	p := &DmParams{}
	for _, op := range options {
		if err := op(p); err != nil {
			return nil, err
		}
	}

	return p, nil
}

type DmOption func(*DmParams) error
