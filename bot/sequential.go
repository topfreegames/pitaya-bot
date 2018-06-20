package bot

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/topfreegames/pitaya-bot/models"
	"github.com/topfreegames/pitaya/client"
)

// SequentialBot defines the struct for the sequential bot that is going to run
type SequentialBot struct {
	client  *client.Client
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

// TODO - refactor
func (b *SequentialBot) runOperation(op *models.Operation) error {
	if op.Request != "" {
		route := op.Request
		b.logger.Info("Will send request: ", route)

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

	b.logger.Info("Will execute internal function: ", op.Function)
	switch op.Function {
	case "disconnect":
		b.Disconnect()
	case "connect":
		b.Connect()
	case "reconnect":
		b.Reconnect()
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
	b.client.Disconnect()
	b.client = nil
}

// Connect ...
func (b *SequentialBot) Connect() {
	if b.client != nil {
		if b.client.Connected {
			fmt.Println("Client already connected")
			return
		}

		b.client = nil
	}

	pclient := client.New(logrus.InfoLevel)
	host := b.config.GetString("server.host")
	err := pclient.ConnectTo(host)
	if err != nil {
		fmt.Println("Error connecting to server")
		fmt.Println(err)
	}

	b.client = pclient
}

// Reconnect ...
func (b *SequentialBot) Reconnect() {
	b.Disconnect()
	b.Connect()
	b.logger.Info("done")
}
