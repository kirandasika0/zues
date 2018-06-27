// Package dispatch implements a basic pubsub model to allow
// other parts of the system to handle websockets connections with grace
package dispatch

// TODO: Need to update this package and abstract any type of websocket connection type

import (
	"fmt"
	"sync"

	"github.com/kataras/golog"

	"github.com/gorilla/websocket"
)

var channels = map[string]*Channel{}

// Channel is a struct that represents a channel is a pubsub environment
type Channel struct {
	name        string
	lCount      uint32
	listeners   []*Listener
	updateMutex sync.RWMutex
}

// Listener is a struct that represents a lister on a particular channel
type Listener struct {
	wsConn *websocket.Conn
}

// NewChannel return a new channel with the given name
func NewChannel(channelName string) (*Channel, error) {
	if channelName == "" {
		return nil, fmt.Errorf("please provide a channel name")
	}
	_, ok := channels[channelName]
	if ok {
		// Channel already exists
		return nil, fmt.Errorf("channel with name %s already exists", channelName)
	}
	var newChannel = Channel{
		name:        channelName,
		lCount:      0,
		updateMutex: sync.RWMutex{},
	}
	// Add it to the cached channels
	channels[channelName] = &newChannel
	return &newChannel, nil
}

// GetChannel return a channel from cache
func GetChannel(channelName string) (*Channel, error) {
	channel, ok := channels[channelName]
	if !ok {
		return nil, fmt.Errorf("channel with name %s does not exist", channelName)
	}
	return channel, nil
}

// GetListenerCount is a class level method to get the number of listeners on a channel
func GetListenerCount(channelName string) uint32 {
	c, ok := channels[channelName]
	if !ok {
		return 0
	}
	return c.LCount()
}

// Name returns the channel name
func (c *Channel) Name() string {
	return c.name
}

// LCount return the listener count for a channel
func (c *Channel) LCount() uint32 {
	return c.lCount
}

// Listeners returns the slice of listeners on a channel
func (c *Channel) Listeners() []*Listener {
	return c.listeners
}

// AddListener adds a listener to the current channel
func (c *Channel) AddListener(wsConn *websocket.Conn) (bool, error) {
	if wsConn == nil {
		return false, fmt.Errorf("please pass a websocket connection")
	}
	c.updateMutex.Lock()
	c.listeners = append(c.listeners, &Listener{wsConn})
	c.lCount++
	c.updateMutex.Unlock()
	return true, nil
}

// Broadcast write a message to all the listeners waiting on a channel
func (c *Channel) Broadcast(message interface{}) error {
	for i, l := range c.listeners {
		err := websocket.WriteJSON(l.wsConn, message)
		if err != nil {
			// Either the client has closed the connection or the connection was lost
			// Removing the connection from the listeners
			golog.Infof("removing wsConn from listeners slice.")
			c.listeners = append(c.listeners[:i], c.listeners[i+1:]...)
			c.lCount--
			continue
		}
	}
	return nil
}

// CloseChannel closes all the currently open socket connections
// returns the number of closed socket connections and an error
// if something goes wrong
func CloseChannel(channelName string) (uint32, error) {
	if channelName == "" {
		return 0, fmt.Errorf("channel name need to close a channel listeners")
	}
	c, ok := channels[channelName]
	if !ok {
		return 0, fmt.Errorf("channel with name %s not found", channelName)
	}
	var closedConn uint32
	for _, l := range c.listeners {
		l.wsConn.Close()
		closedConn++
	}
	// Delete the channel
	delete(channels, channelName)
	return closedConn, nil
}
