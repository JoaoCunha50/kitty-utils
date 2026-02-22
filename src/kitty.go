package src

type KittyInstance struct {
	Tabs   []string
	Socket string
}

var socket string = "mykitty"

func InitKitty(socket string) *KittyInstance {
	return &KittyInstance{
		Tabs:   []string{},
		Socket: socket,
	}
}

type Kitty interface {
	GetTabs() []string
	SetTabs([]string)
	GetSocket() string
	SetSocket(string)
}

func (k *KittyInstance) GetTabs() []string {
	return k.Tabs
}

func (k *KittyInstance) SetTabs(tabs []string) {
	k.Tabs = tabs
}

func (k *KittyInstance) GetSocket() string {
	return k.Socket
}

func (k *KittyInstance) SetSocket(socket string) {
	k.Socket = socket
}
