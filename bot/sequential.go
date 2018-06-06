package bot

import (
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
}

// NewSequentialBot returns a new sequantial bot instance
func NewSequentialBot(config *viper.Viper, spec *models.Spec, id int) (Bot, error) {
	pclient := client.New(logrus.InfoLevel)
	host := config.GetString("server.host")
	err := pclient.ConnectTo(host)
	if err != nil {
		return nil, err
	}

	bot := &SequentialBot{
		client:  pclient,
		config:  config,
		spec:    spec,
		id:      id,
		storage: newStorage(config),
	}

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

func (b *SequentialBot) runOperation(op *models.Operation) error {
	route := op.Request
	args, err := buildArgs(op.Args, b.storage)
	if err != nil {
		return err
	}

	resp, err := sendRequest(args, route, b.client)
	if err != nil {
		return err
	}

	err = validateExpectations(op.Expect, resp)
	if err != nil {
		return err
	}

	err = storeData(op.Store, b.storage, resp)
	if err != nil {
		return err
	}

	// TODO: Store change

	return nil
}

// Finalize finalizes the bot
func (b *SequentialBot) Finalize() error {
	// TODO
	return nil
}
