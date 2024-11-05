package kaspa

import (
	"context"
	eosclient "cryp-kaspad/internal/libs/kaspa/client"
	"fmt"
	"time"

	_ "github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/cmd/kaspawallet/daemon/server"
	_ "github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/infrastructure/config"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
)

type Eos struct {
	Client    *eosclient.Client
	RpcClient *rpcclient.RPCClient
}

func NewClient(ctx context.Context, urls []string) (*Eos, error) {

	for _, url := range urls {
		client, err := eosclient.Dial(url)
		if err != nil {
			//log.Warn("eosclient.Dial failed", "url", url, "error", err)
			continue
		}

		if !Ping(client) {
			//log.Warn("eosclient.Dial ping failed", "url", url)
			continue
		}

		return &Eos{Client: client}, nil
	}

	return nil, fmt.Errorf("eosclient.Dial all url failed, urls: %v", urls)
}

func Ping(client *eosclient.Client) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.GetInfo(ctx)
	return err == nil
}
func StartDeamon(ctx context.Context, urls []string) (*Eos, error) {

	fmt.Println("rpcServer:", urls[0])
	client, err := rpcclient.NewRPCClient(urls[0])
	if err != nil {
		fmt.Println("Error connecting to the server:", err)
	}
	return &Eos{RpcClient: client}, nil

}

type startDaemonConfig struct {
	KeysFile  string `long:"keys-file" short:"f" description:"Keys file location (default: ~/.kaspawallet/keys.json (*nix), %USERPROFILE%\\AppData\\Local\\Kaspawallet\\key.json (Windows))"`
	Password  string `long:"password" short:"p" description:"Wallet password"`
	RPCServer string `long:"rpcserver" short:"s" description:"RPC server to connect to"`
	Listen    string `long:"listen" short:"l" description:"Address to listen on (default: 0.0.0.0:8082)"`
	Timeout   uint32 `long:"wait-timeout" short:"w" description:"Waiting timeout for RPC calls, seconds (default: 30 s)"`
	Profile   string `long:"profile" description:"Enable HTTP profiling on given port -- NOTE port must be between 1024 and 65536"`
	config.NetworkFlags
}

func startDaemon(conf *startDaemonConfig) error {
	return server.Start(conf.NetParams(), conf.Listen, conf.RPCServer, conf.KeysFile, conf.Profile, conf.Timeout)
}
