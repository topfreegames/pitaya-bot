package bot

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/topfreegames/pitaya-bot/constants"
	"github.com/topfreegames/pitaya-bot/custom"
	"github.com/topfreegames/pitaya-bot/metrics"
	"github.com/topfreegames/pitaya-bot/models"
	"github.com/topfreegames/pitaya-bot/storage"
)

// SequentialBot defines the struct for the sequential bot that is going to run
type SequentialBot struct {
	client          *PClient
	config          *viper.Viper
	host            string
	id              int
	logger          logrus.FieldLogger
	metricsReporter []metrics.Reporter
	spec            *models.Spec
	storage         storage.Storage
}

// NewSequentialBot returns a new sequantial bot instance
func NewSequentialBot(
	config *viper.Viper,
	spec *models.Spec,
	id int,
	mr []metrics.Reporter,
	logger logrus.FieldLogger,
) (Bot, error) {
	store, err := storage.NewStorage(config)
	if err != nil {
		return nil, err
	}

	bot := &SequentialBot{
		config:          config,
		host:            config.GetString("server.host"),
		id:              id,
		logger:          logger,
		metricsReporter: mr,
		spec:            spec,
		storage:         store,
	}

	if err = bot.Connect(); err != nil {
		return nil, err
	}

	return bot, nil
}

// Initialize initializes the bot
func (b *SequentialBot) Initialize() error {
	b.logger.Debug("Initializing bot")
	pre, args := custom.GetPre(b.config, b.spec)
	storage, err := pre.Run(args)
	if err != nil {
		return err
	}
	b.logger.Debugf("Received storage: %+v", storage)
	b.storage = storage
	return nil
}

// Run runs the bot
func (b *SequentialBot) Run() (err error) {
	defer b.Disconnect()
	defer func() {
		if rec := recover(); rec != nil {
			b.logger.Errorf("Panic running sequential bot: %+v", rec)
			err = fmt.Errorf("panic")
		}
	}()

	steps := b.spec.SequentialOperations
	for _, step := range steps {
		err = b.runOperation(step)
		if err != nil {
			return
		}
	}

	return
}

func (b *SequentialBot) runRequest(op *models.Operation) error {
	b.logger.Debug("Executing request to: " + op.URI)
	route := op.URI
	args, err := buildArgByType(op.Args, "object", b.storage)
	if err != nil {
		return err
	}

	resp, rawResp, err := sendRequest(args, route, b.client, b.metricsReporter, b.logger)
	if err != nil {
		return err
	}

	b.logger.Debug("validating expectations")
	err = validateExpectations(op.Expect, resp, b.storage)
	if err != nil {
		return NewExpectError(err, rawResp, op.Expect)
	}
	b.logger.Debug("received valid response")

	b.logger.Debug("storing data")
	err = storeData(op.Store, b.storage, resp)
	if err != nil {
		return err
	}

	b.logger.Debug("all done")
	return nil
}

func (b *SequentialBot) runNotify(op *models.Operation) error {
	b.logger.Debug("Executing notify to: " + op.URI)
	route := op.URI
	args, err := buildArgByType(op.Args, "object", b.storage)
	if err != nil {
		return err
	}

	err = sendNotify(args, route, b.client)
	if err != nil {
		return err
	}

	b.logger.Debug("all done")
	return nil
}

func (b *SequentialBot) runFunction(op *models.Operation) error {
	fName := op.URI
	b.logger.Debug("Will execute internal function: ", fName)

	switch fName {
	case "disconnect":
		b.Disconnect()
	case "connect":
		host := b.host
		args, err := buildArgByType(op.Args, "object", b.storage)
		if err != nil {
			return err
		}
		mapArgs, ok := args.(map[string]interface{})
		if !ok {
			return constants.ErrMalformedObject
		}
		if val, ok := mapArgs["host"]; ok {
			b.logger.Debug("Connecting to custom host")
			if h, ok := val.(string); ok {
				host = h
			}
		}
		b.Connect(host)
	case "reconnect":
		b.Reconnect()
	default:
		return fmt.Errorf("Unknown function: %s", fName)
	}

	return nil
}

func (b *SequentialBot) listenToPush(op *models.Operation) error {
	b.logger.Debug("Waiting for push on route: " + op.URI)
	resp, err := b.client.ReceivePush(op.URI, op.Timeout)
	if err != nil {
		return err
	}

	b.logger.Debug("validating expectations")
	err = validateExpectations(op.Expect, resp, b.storage)
	if err != nil {
		return err
	}
	b.logger.Debug("received valid response")

	b.logger.Debug("storing data")
	err = storeData(op.Store, b.storage, resp)
	if err != nil {
		return err
	}

	b.logger.Debug("all done")
	return nil
}

// StartListening ...
func (b *SequentialBot) startListening() {
	b.client.StartListening()
}

// TODO - refactor
func (b *SequentialBot) runOperation(op *models.Operation) error {
	switch op.Type {
	case "request":
		return b.runRequest(op)
	case "notify":
		return b.runNotify(op)
	case "function":
		return b.runFunction(op)
	case "listen":
		return b.listenToPush(op)
	}

	return fmt.Errorf("Unknown type: %s", op.Type)
}

// Finalize finalizes the bot
func (b *SequentialBot) Finalize() error {
	b.logger.Debug("Finalizing bot")
	post, args := custom.GetPost(b.config, b.spec)
	if err := post.Run(args, b.storage); err != nil {
		return err
	}
	b.logger.Debugf("Saved storage")
	return nil
}

// Disconnect ...
func (b *SequentialBot) Disconnect() {
	b.client.Disconnect()
}

// Connect ...
func (b *SequentialBot) Connect(hosts ...string) error {
	if len(hosts) > 0 {
		b.host = hosts[0]
	}
	if b.client != nil && b.client.Connected() {
		b.logger.Fatal("Bot already connected")
	}

	pushinfoprotos := b.config.GetStringSlice("server.protobuffer.pushinfo.protos")
	pushinforoutes := b.config.GetStringSlice("server.protobuffer.pushinfo.routes")
	if len(pushinforoutes) != len(pushinfoprotos) {
		b.logger.Fatal("Invalid number of protos routes or protos.")
	}
	pushinfo := make(map[string]string)
	for i := range pushinfoprotos {
		pushinfo[pushinforoutes[i]] = pushinfoprotos[i]
	}

	servertype := b.config.GetString("server.serializer")
	docs := ""
	if servertype == "protobuffer" {
		docs = b.config.GetString("server.protobuffer.docs")
	}

	tls := b.config.GetBool("server.tls")
	client, err := NewPClient(b.host, tls, docs, pushinfo)
	if err != nil {
		b.logger.WithError(err).Error("Unable to create client...")
		return err
	}

	b.client = client
	b.startListening()
	return nil
}

// Reconnect ...
func (b *SequentialBot) Reconnect() {
	b.Disconnect()
	b.Connect()
	b.logger.Debug("Reconnect done")
}
