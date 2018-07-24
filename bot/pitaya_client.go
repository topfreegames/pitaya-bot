package bot

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/topfreegames/pitaya/client"
)

// FIXME - contants from interal pitaya package
const (
	MsgResponseType byte = 0x02
	MsgPushType     byte = 0x03
)

// PClient is a wrapper arund pitaya/client.
// The ideia is to be able to separeta request/responses
// from server pushes.
type PClient struct {
	client         *client.Client
	responsesMutex sync.Mutex
	responses      map[uint]chan []byte

	pushesMutex sync.Mutex
	pushes      map[string]chan []byte
}

// NewPClient is the PCLient constructor
func NewPClient(host string) *PClient {
	pclient := client.New(logrus.InfoLevel)
	err := pclient.ConnectTo(host)
	if err != nil {
		fmt.Println("Error connecting to server")
		fmt.Println(err)
	}

	return &PClient{
		client:    pclient,
		responses: make(map[uint]chan []byte),
		pushes:    make(map[string]chan []byte),
	}
}

// Disconnect disconnects the client
func (c *PClient) Disconnect() {
	c.client.Disconnect()
	c.client = nil
}

// Connected returns if the given client is connected or not
func (c *PClient) Connected() bool {
	return c.client != nil && c.client.Connected
}

func (c *PClient) getResponseChannelForID(id uint) chan []byte {
	c.responsesMutex.Lock()
	defer c.responsesMutex.Unlock()
	if _, ok := c.responses[id]; !ok {
		c.responses[id] = make(chan []byte)
	}

	return c.responses[id]
}

func (c *PClient) removeResponseChannelForID(id uint) {
	c.responsesMutex.Lock()
	defer c.responsesMutex.Unlock()

	delete(c.responses, id)
}

func (c *PClient) getPushChannelForRoute(route string) chan []byte {
	c.pushesMutex.Lock()
	defer c.pushesMutex.Unlock()
	if _, ok := c.pushes[route]; !ok {
		c.pushes[route] = make(chan []byte)
	}

	return c.pushes[route]
}

// Request ...
func (c *PClient) Request(route string, data []byte) (Response, []byte, error) {
	messageID, err := c.client.SendRequest(route, data)
	if err != nil {
		return nil, nil, err
	}

	ch := c.getResponseChannelForID(messageID)

	select {
	case responseData := <-ch:
		ret := make(Response)
		if err := json.Unmarshal(responseData, &ret); err != nil {
			err = fmt.Errorf("Error unmarshaling response: %s", err)
			return nil, nil, err
		}

		return ret, responseData, nil
	case <-time.After(time.Second):
		return nil, nil, fmt.Errorf("Timeout waiting for response on route %s", route)
	}

	return nil, nil, nil
}

// Notify sends a notify to the server
func (c *PClient) Notify(route string, data []byte) error {
	err := c.client.SendNotify(route, data)
	return err
}

// ReceivePush ...
func (c *PClient) ReceivePush(route string, timeout int) (Response, error) {
	ch := c.getPushChannelForRoute(route)

	select {
	case data := <-ch:
		ret := make(Response)
		if err := json.Unmarshal(data, &ret); err != nil {
			err = fmt.Errorf("Error unmarshaling response: %s", err)
			return nil, err
		}

		return ret, nil
	case <-time.After(time.Duration(timeout) * time.Millisecond):
		return nil, fmt.Errorf("Timeout waiting for push on route %s", route)
	}
}

// StartListening ...
func (c *PClient) StartListening() {
	fmt.Println("Listening...")
	go func() {
		for m := range c.client.IncomingMsgChan {
			t := byte(m.Type)
			switch t {
			case MsgResponseType:
				ch := c.getResponseChannelForID(m.ID)
				ch <- m.Data
				c.removeResponseChannelForID(m.ID)
			case MsgPushType:
				ch := c.getPushChannelForRoute(m.Route)
				ch <- m.Data
			default:
				panic("Unknown message type")
			}
		}
	}()
}
