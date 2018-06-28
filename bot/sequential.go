package bot

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/topfreegames/pitaya-bot/models"
)

// SequentialBot defines the struct for the sequential bot that is going to run
type SequentialBot struct {
	client  *PClient
	config  *viper.Viper
	id      int
	spec    *models.Spec
	storage *storage
	logger  logrus.FieldLogger
}

// NewSequentialBot returns a new sequantial bot instance
func NewSequentialBot(config *viper.Viper, spec *models.Spec, id int) (Bot, error) {
	bot := &SequentialBot{
		config:  config,
		spec:    spec,
		id:      id,
		storage: newStorage(config),
		logger:  logrus.New(),
	}

	bot.Connect()

	return bot, nil
}

// Initialize initializes the bot
func (b *SequentialBot) Initialize() error {
	// TODO
	return nil
}

// Run runs the bot
func (b *SequentialBot) Run() error {
	steps := b.spec.SequentialOperations

	for _, step := range steps {
		err := b.runOperation(step)
		if err != nil {
			// TODO: Treat errors
			return err
		}
	}

	return nil
}

func (b *SequentialBot) runRequest(op *models.Operation) error {
	b.logger.Info("Executing request to: " + op.URI)
	route := op.URI
	args, err := buildArgs(op.Args, b.storage)
	if err != nil {
		return err
	}

	resp, err := sendRequest(args, route, b.client)
	if err != nil {
		return err
	}

	b.logger.Info("validating expectations")
	err = validateExpectations(op.Expect, resp, b.storage)
	if err != nil {
		return err
	}
	b.logger.Info("received valid response")

	b.logger.Info("storing data")
	err = storeData(op.Store, b.storage, resp)
	if err != nil {
		return err
	}

	b.logger.Info("all done")
	return nil
}

func (b *SequentialBot) runFunction(op *models.Operation) error {
	b.logger.Info("Will execute internal function: ", op.URI)
	switch op.URI {
	case "disconnect":
		b.Disconnect()
	case "connect":
		b.Connect()
	case "reconnect":
		b.Reconnect()
	}

	return nil
}

func (b *SequentialBot) listenToPush(op *models.Operation) error {
	b.logger.Info("Waiting for push on route: " + op.URI)
	resp, err := b.client.ReceivePush(op.URI)
	if err != nil {
		return err
	}

	b.logger.Info("validating expectations")
	err = validateExpectations(op.Expect, resp, b.storage)
	if err != nil {
		return err
	}
	b.logger.Info("received valid response")

	b.logger.Info("storing data")
	err = storeData(op.Store, b.storage, resp)
	if err != nil {
		return err
	}

	b.logger.Info("all done")
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
	case "function":
		return b.runFunction(op)
	case "listen":
		return b.listenToPush(op)
	}

	return nil
}

// Finalize finalizes the bot
func (b *SequentialBot) Finalize() error {
	// TODO
	return nil
}

// Disconnect ...
func (b *SequentialBot) Disconnect() {
	fmt.Println("Disconnect")
	b.client.Disconnect()
}

// Connect ...
func (b *SequentialBot) Connect() {
	fmt.Println("Connect")
	host := b.config.GetString("server.host")
	if b.client != nil && b.client.Connected() {
		b.logger.Fatal("Bot already connected")
	}

	b.client = NewPClient(host)
	b.startListening()
}

// Reconnect ...
func (b *SequentialBot) Reconnect() {
	b.Disconnect()
	b.Connect()
	b.logger.Info("Reconnect done")
}
