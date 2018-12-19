package bot

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/topfreegames/pitaya/client"
)

// information for the singleton
var instance *client.ProtoBufferInfo
var once sync.Once

// FIXME - constants from internal pitaya package
const (
	MsgResponseType byte = 0x02
	MsgPushType     byte = 0x03
)

// PClient is a wrapper around pitaya/client.
// The ideia is to be able to separate request/responses
// from server pushes
type PClient struct {
	client         client.PitayaClient
	responsesMutex sync.Mutex
	responses      map[uint]chan []byte

	pushesMutex sync.Mutex
	pushes      map[string]chan []byte
}

func getProtoInfo(host string, docs string, pushinfo map[string]string) *client.ProtoBufferInfo {
	once.Do(func() {
		cli := client.NewProto(docs, logrus.InfoLevel)
		for k, v := range pushinfo {
			cli.AddPushResponse(k, v)
		}
		err := cli.LoadServerInfo(host)
		if err != nil {
			fmt.Println("Unable to load server documentation.")
			fmt.Println(err)
		} else {
			instance = cli.ExportInformation()
		}
	})
	return instance
}

// NewPClient is the PCLient constructor
func NewPClient(host string, useTLS bool, docs string, pushinfo map[string]string) (*PClient, error) {

	var pclient client.PitayaClient
	if docs != "" {
		protoclient := client.NewProto(docs, logrus.InfoLevel)
		pclient = protoclient
		if err := protoclient.LoadInfo(getProtoInfo(host, docs, pushinfo)); err != nil {
			return nil, err
		}
	} else {
		pclient = client.New(logrus.InfoLevel)
	}

	if useTLS {
		if err := pclient.ConnectToTLS(host, true); err != nil {
			fmt.Println("Error connecting to server")
			fmt.Println(err)
			return nil, err
		}
	} else {
		if err := pclient.ConnectTo(host); err != nil {
			fmt.Println("Error connecting to server")
			fmt.Println(err)
			return nil, err
		}
	}

	return &PClient{
		client:    pclient,
		responses: make(map[uint]chan []byte),
		pushes:    make(map[string]chan []byte),
	}, nil
}

// Disconnect disconnects the client
func (c *PClient) Disconnect() {
	c.client.Disconnect()
	c.client = nil
}

// Connected returns if the given client is connected or not
func (c *PClient) Connected() bool {
	return c.client != nil && c.client.ConnectedStatus()
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
		var ret Response
		if err := json.Unmarshal(responseData, &ret); err != nil {
			err = fmt.Errorf("Error unmarshaling response: %s", err)
			return nil, nil, err
		}

		return ret, responseData, nil
	case <-time.After(5 * time.Second): // TODO - pass timeout as config
		return nil, nil, fmt.Errorf("Timeout waiting for response on route %s", route)
	}
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
		var ret Response
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
	channel := c.client.MsgChannel()
	go func() {
		for m := range channel {
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
