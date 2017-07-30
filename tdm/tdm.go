package tdm

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"sync"

	"github.com/ChimeraCoder/anaconda"
)

type Tdm struct {
	sync.RWMutex
	api    *anaconda.TwitterApi
	stream *anaconda.Stream
	user   anaconda.User
	chats  map[int64]*Chat
}

type Chat struct {
	username    string
	userID      int64
	receivedDms []anaconda.DirectMessage
	sentDms     []anaconda.DirectMessage
}

func NewTdm(consumerKey, consumerSecret, accessToken, accessTokenSecret string) (*Tdm, error) {
	anaconda.SetConsumerKey(consumerKey)
	anaconda.SetConsumerSecret(consumerSecret)

	tdm := new(Tdm)
	tdm.api = anaconda.NewTwitterApi(accessToken, accessTokenSecret)
	if err := tdm.getUser(); err != nil {
		return nil, err
	}
	if err := tdm.initChats(); err != nil {
		return nil, err
	}

	if err := tdm.FetchChats(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed fetching chats: %v", err)
	}

	tdm.OpenStream()

	dmStream, err := tdm.GetDmStream()
	if err != nil {
		return nil, err
	}

	go func() {
		for dm := range dmStream {
			tdm.addDm(dm)
			tdm.Log()
		}
	}()

	return tdm, nil
}

func (tdm *Tdm) getUser() error {
	v := url.Values{}
	v.Set("include_entities", "false")
	v.Set("skip_status", "true")

	user, err := tdm.api.GetSelf(v)
	if err != nil {
		return err
	}

	tdm.user = user

	return nil
}

func (tdm *Tdm) initChats() error {
	// TODO fetch from cache on disk or something

	tdm.chats = make(map[int64]*Chat, 1)

	return nil
}

func (tdm *Tdm) FetchChats() error {
	// TODO: Use goroutines
	dms, err := tdm.GetDms(DmParams{Count: 200})
	if err != nil {
		return err
	}

	for _, dm := range dms {
		tdm.addDm(dm)
	}

	dms, err = tdm.GetSentDms(DmParams{Count: 200})
	if err != nil {
		return err
	}

	for _, dm := range dms {
		tdm.addDm(dm)
	}

	return nil
}

func (tdm *Tdm) addDm(dm anaconda.DirectMessage) error {
	var chatterID int64
	var chatterUsername string
	if dm.SenderId == tdm.user.Id {
		chatterID = dm.RecipientId
		chatterUsername = dm.RecipientScreenName
	} else {
		chatterID = dm.SenderId
		chatterUsername = dm.SenderScreenName
	}

	tdm.Lock()
	chat, ok := tdm.chats[chatterID]
	if !ok {
		newChat := Chat{username: chatterUsername, userID: chatterID}
		tdm.chats[chatterID] = &newChat
		chat = &newChat
	}

	if dm.SenderId == tdm.user.Id {
		chat.sentDms = append(chat.sentDms, dm)
	} else {
		chat.receivedDms = append(chat.receivedDms, dm)
	}
	tdm.Unlock()

	return nil
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

func (tdm *Tdm) Log() {
	tdm.RLock()

	fmt.Printf("{\n")
	fmt.Printf("\tUser: %+v\n", tdm.user.ScreenName)
	fmt.Printf("\tChats:\n")
	for _, chat := range tdm.chats {
		chat.Log()
	}
	fmt.Printf("}\n\n")

	tdm.RUnlock()
}

func (chat *Chat) Log() {
	fmt.Printf("\t\tChatter: %v\n", chat.username)
	fmt.Printf("\t\tSent: \n")
	for _, sentDm := range chat.sentDms {
		fmt.Printf("\t\t\tText: %v\n", sentDm.Text)
	}

	fmt.Printf("\t\tReceived: \n")
	for _, receivedDm := range chat.receivedDms {
		fmt.Printf("\t\t\tText: %v\n", receivedDm.Text)
	}
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
