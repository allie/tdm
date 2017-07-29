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

func (tdm *Tdm) SendDmToUsername(text, screenName string) (anaconda.DirectMessage, error) {
	return tdm.api.PostDMToScreenName(text, screenName)
}

func (tdm *Tdm) SendDmToId(text string, userId int64) (anaconda.DirectMessage, error) {
	return tdm.api.PostDMToUserId(text, userId)
}

func (tdm *Tdm) DeleteDm(id int64) (anaconda.DirectMessage, error) {
	return tdm.api.DeleteDirectMessage(id, false)
}

func (tdm *Tdm) OpenStream() {
	v := url.Values{}
	v.Set("with", "user")
	tdm.stream = tdm.api.UserStream(v)
}

func (tdm *Tdm) CloseStream() {
	tdm.stream.Stop()
}

func (tdm *Tdm) GetDmStream() (chan anaconda.DirectMessage, error) {
	if tdm.stream == nil {
		return nil, fmt.Errorf("No open stream")
	}

	dms := make(chan anaconda.DirectMessage)

	filter := func(streamEvents <-chan interface{}) {
		for event := range streamEvents {
			if dm, ok := event.(anaconda.DirectMessage); ok {
				dms <- dm
			}
		}
	}

	go filter(tdm.stream.C)

	return dms, nil
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
	// TODO: Use chans and a goroutine to fetch following info
	// for every 100 friends as they come in

	friends := make([]anaconda.User, 1, 100)

	v := url.Values{}
	v.Set("skip_status", "true")
	v.Set("include_user_entities", "false")

	followingPages := tdm.api.GetFriendsListAll(v)
	for page := range followingPages {
		friends = append(friends, page.Friends...)
	}

	return friends, nil
}

type DmParams struct {
	SinceId int64
	MaxId   int64
	Count   int
}

func (p *DmParams) ToValues() url.Values {
	v := url.Values{}
	if p.SinceId > 0 {
		v.Set("since_id", strconv.FormatInt(p.SinceId, 10))
	}
	if p.MaxId > 0 {
		v.Set("max_id", strconv.FormatInt(p.MaxId, 10))
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
