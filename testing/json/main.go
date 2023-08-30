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
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/acceptor"
	"github.com/topfreegames/pitaya/v2/component"
	"github.com/topfreegames/pitaya/v2/config"
	pitayalogrus "github.com/topfreegames/pitaya/v2/logger/logrus"
	"github.com/topfreegames/pitaya/v2/serialize/json"
)

// PlayerHandler ...
type PlayerHandler struct {
	component.Base
}

// AuthArg ...
type AuthArg struct {
	AccessToken uuid.UUID `json:"accessToken"`
}

// findMatchArg ...
type findMatchArg struct {
	RoomType string `json:"roomType"`
}

// Player ...
type Player struct {
	PrivateID    uuid.UUID `json:"privateID"`
	AccessToken  uuid.UUID `json:"accessToken"`
	Name         string    `json:"name"`
	SoftCurrency int       `json:"softCurrency"`
	Trophies     int       `json:"trophies"`
}

// AuthResponse ...
type AuthResponse struct {
	Code   string  `json:"code"`
	Msg    string  `json:"msg"`
	Player *Player `json:"player"`
}

// FindMatchResponse ...
type FindMatchResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

type findMatchPush struct {
	Code string `json:"code"`
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

var (
	player = &Player{
		PrivateID:    uuid.New(),
		AccessToken:  uuid.New(),
		Name:         "john doe",
		SoftCurrency: 100,
		Trophies:     2,
	}

	app pitaya.Pitaya
)

// Create ...
func (p *PlayerHandler) Create(ctx context.Context) (*AuthResponse, error) {
	if err := bindSession(ctx, player.PrivateID); err != nil {
		panic(err)
	}

	return &AuthResponse{
		Code:   "200",
		Player: player,
	}, nil
}

func bindSession(ctx context.Context, uid uuid.UUID) error {
	return app.GetSessionFromCtx(ctx).Bind(ctx, uid.String())
}

// Authenticate ...
func (p *PlayerHandler) Authenticate(ctx context.Context, arg *AuthArg) (*AuthResponse, error) {
	if err := bindSession(ctx, player.PrivateID); err != nil {
		panic(err)
	}

	return &AuthResponse{
		Code:   "200",
		Player: player,
	}, nil
}

// FindMatch ...
func (p *PlayerHandler) FindMatch(ctx context.Context, arg *findMatchArg) (*FindMatchResponse, error) {
	go func() {
		time.Sleep(200 * time.Millisecond)
		response := findMatchPush{
			Code: "200",
			IP:   "127.0.0.1",
			Port: 9090,
		}
		if _, err := app.SendPushToUsers("connector.playerHandler.matchfound", response, []string{player.PrivateID.String()}, "connector"); err != nil {
			panic(err)
		}
	}()

	return &FindMatchResponse{
		Code: "200",
	}, nil
}

func main() {
	port := flag.Int("port", 30123, "the port to listen")
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

	tcp := acceptor.NewTCPAcceptor(fmt.Sprintf(":%d", *port))

	plogger := pitayalogrus.NewWithEntry(l.WithField("app", "pitaya-bot-testing-sv"))
	pitaya.SetLogger(plogger)

	l.Infof("Port: %d", *port)
	cfg := viper.New()
	cfg.Set("pitaya.cluster.sd.etcd.prefix", *sdPrefix)

	builder := pitaya.NewBuilderWithConfigs(true, *svType, pitaya.Cluster, map[string]string{}, config.NewConfig(cfg))
	builder.AddAcceptor(tcp)
	builder.Serializer = json.NewSerializer()
	app = builder.Build()

	app.Register(
		&PlayerHandler{},
		component.WithName("playerHandler"),
		component.WithNameFunc(strings.ToLower),
	)

	app.Start()
}
