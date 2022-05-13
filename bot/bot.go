package bot

// Bot defines the interface the bots must implement
type Bot interface {
	Initialize() error
	Run() error
	Finalize() error
	Connect(...string) error
	Disconnect()
	Reconnect()
	Sleep(int)
}
