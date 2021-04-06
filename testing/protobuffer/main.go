// Copyright (c) TFG Co. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"time"

	gogoproto "github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/topfreegames/pitaya"
	"github.com/topfreegames/pitaya/acceptor"
	"github.com/topfreegames/pitaya/component"
	"github.com/topfreegames/pitaya/constants"
	pitayaprotos "github.com/topfreegames/pitaya/protos"
	"github.com/topfreegames/pitaya/serialize/protobuf"

	"github.com/topfreegames/pitaya-bot/testing/protobuffer/protos"
)

// DocsHandler ...
type DocsHandler struct {
	component.Base
}

// PlayerHandler ...
type PlayerHandler struct {
	component.Base
}

// Player ...
type Player struct {
	PrivateID    uuid.UUID `json:"privateID"`
	AccessToken  uuid.UUID `json:"accessToken"`
	Name         string    `json:"name"`
	SoftCurrency int       `json:"softCurrency"`
	Trophies     int       `json:"trophies"`
}

var (
	player = &Player{
		PrivateID:    uuid.New(),
		AccessToken:  uuid.New(),
		Name:         "john doe",
		SoftCurrency: 100,
		Trophies:     2,
	}
)

// Proto transform Player in proto.Player
func (p *Player) getProto() *protos.Player {
	return &protos.Player{
		PrivateID:    p.PrivateID.String(),
		AccessToken:  p.AccessToken.String(),
		Name:         p.Name,
		SoftCurrency: int32(p.SoftCurrency),
		Trophies:     int32(p.Trophies),
	}
}

// Create ...
func (p *PlayerHandler) Create(ctx context.Context) (*protos.AuthResponse, error) {
	if err := bindSession(ctx, player.PrivateID); err != nil {
		panic(err)
	}

	return &protos.AuthResponse{
		Code:   "200",
		Player: player.getProto(),
	}, nil
}

func bindSession(ctx context.Context, uid uuid.UUID) error {
	return pitaya.GetSessionFromCtx(ctx).Bind(ctx, uid.String())
}

// Authenticate ...
func (p *PlayerHandler) Authenticate(ctx context.Context, arg *protos.AuthArg) (*protos.AuthResponse, error) {
	if err := bindSession(ctx, player.PrivateID); err != nil {
		panic(err)
	}
	return &protos.AuthResponse{
		Code:   "200",
		Player: player.getProto(),
	}, nil
}

// FindMatch ...
func (p *PlayerHandler) FindMatch(ctx context.Context, arg *protos.FindMatchArg) (*protos.FindMatchResponse, error) {
	go func() {
		time.Sleep(200 * time.Millisecond)
		response := &protos.FindMatchPush{
			Code: "200",
			IP:   "127.0.0.1",
			Port: 9090,
		}
		uuids := []string{player.PrivateID.String()}
		if _, err := pitaya.SendPushToUsers("connector.playerHandler.matchfound", response, uuids, "connector"); err != nil {
			panic(err)
		}
	}()

	return &protos.FindMatchResponse{
		Code: "200",
	}, nil
}

// Docs returns documentation
func (d *DocsHandler) Docs(ctx context.Context) (*pitayaprotos.Doc, error) {
	docs, err := pitaya.Documentation(true)
	if err != nil {
		return nil, err
	}

	jsonDocs, err := json.Marshal(docs)
	if err != nil {
		return nil, err
	}

	return &pitayaprotos.Doc{Doc: string(jsonDocs)}, nil
}

// Protos return protobuffers descriptors
func (d *DocsHandler) Protos(ctx context.Context, message *pitayaprotos.ProtoNames) (*pitayaprotos.ProtoDescriptors, error) {
	var descriptors [][]byte

	for _, name := range message.GetName() {
		protoDescriptor, err := pitaya.Descriptor(name)
		if err != nil {
			// Not a default pitaya proto, so we can probably find it here
			protoDescriptor, err = getDescriptorFromName(name)
			if err != nil { 
				return nil, constants.ErrProtodescriptor
			}
		}

		descriptors = append(descriptors, protoDescriptor)
	}

	return &pitayaprotos.ProtoDescriptors{
		Desc: descriptors,
	}, nil
}

// taken partly from pitaya/docgenerator/descriptors.go
func getDescriptorFromName(protoName string) ([]byte, error) {
	protoReflectTypePointer := gogoproto.MessageType(protoName)
	protoReflectType := protoReflectTypePointer.Elem()
	protoValue := reflect.New(protoReflectType)
	descriptorMethod, ok := protoReflectTypePointer.MethodByName("Descriptor")
	if !ok {
		return nil, constants.ErrProtodescriptor
	}

	descriptorValue := descriptorMethod.Func.Call([]reflect.Value{protoValue})
	protoDescriptor := descriptorValue[0].Bytes()

	return protoDescriptor, nil
}

func main() {
	port := flag.Int("port", 30124, "the port to listen")
	svType := flag.String("type", "connector", "the server type")
	sdPrefix := flag.String("sdprefix", "pitaya/", "prefix to discover other servers")
	debug := flag.Bool("debug", true, "turn on debug logging")

	flag.Parse()

	l := logrus.New()
	l.Formatter = &logrus.TextFormatter{}
	l.SetLevel(logrus.InfoLevel)
	if *debug {
		l.SetLevel(logrus.DebugLevel)
	}

	pitaya.SetLogger(l)

	tcp := acceptor.NewTCPAcceptor(fmt.Sprintf(":%d", *port))

	pitaya.Register(
		&PlayerHandler{},
		component.WithName("playerHandler"),
		component.WithNameFunc(strings.ToLower),
	)

	pitaya.Register(
		&DocsHandler{},
		component.WithName("docsHandler"),
		component.WithNameFunc(strings.ToLower),
	)

	pitaya.SetSerializer(protobuf.NewSerializer())

	l.Infof("Port: %d", *port)
	pitaya.AddAcceptor(tcp)
	cfg := viper.New()
	cfg.Set("pitaya.cluster.sd.etcd.prefix", *sdPrefix)
	isFrontend := true
	pitaya.Configure(isFrontend, *svType, pitaya.Cluster, map[string]string{}, cfg)

	pitaya.Start()
}
