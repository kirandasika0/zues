package dispatch

import (
	"testing"
	"zues/util"

	"github.com/gorilla/websocket"
)

var dChan *Channel

func TestNewChannel(t *testing.T) {
	channelID := util.RandomString(16)
	newChannel, err := NewChannel(channelID)
	if err != nil {
		t.Fatalf("Error while creating the channel %s", channelID)
	}
	if newChannel.LCount() != 0 {
		t.Fail()
	}
	dChan = newChannel
}

func TestGetChannel(t *testing.T) {
	_, err := GetChannel(dChan.Name())
	if err != nil {
		t.Fail()
	}
	randomChanID := util.RandomString(15)
	_, err = GetChannel(randomChanID)
	if err == nil {
		t.Fail()
	}
}

func TestGetListenerCount(t *testing.T) {
	randomChanID := util.RandomString(16)
	lCount := GetListenerCount(randomChanID)
	if lCount != 0 {
		t.Fail()
	}
	lCount = GetListenerCount(dChan.Name())
	if lCount != 0 {
		t.Fail()
	}
}

func TestListeners(t *testing.T) {
	dChan.Listeners()
}

func TestAddListener(t *testing.T) {
	newListner := &websocket.Conn{}
	newChannel, _ := NewChannel(util.RandomString(10))
	isAdded, err := newChannel.AddListener(newListner)
	if err != nil || !isAdded {
		t.Fail()
	}

	isAdded, err = newChannel.AddListener(nil)
	if err == nil || isAdded {
		t.Fail()
	}
}
