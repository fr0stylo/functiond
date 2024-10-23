package worker

import (
	"context"
	"fmt"
	"log"

	gocni "github.com/containerd/go-cni"
)

const conflist = `{
  "cniVersion": "1.0.0",
  "name": "functiond",
  "plugins": [
    {
      "type": "bridge",
      "bridge": "cni0",
      "isGateway": true,
      "isDefaultGateway": true,
      "forceAddress": false,
     "ipMasq": true,
      "promiscMode": true,
      "mtu": 1500,
      "ipam": {
        "type": "host-local",
        "ranges": [
          [
            {
              "subnet": "10.22.0.0/24",
              "gateway": "10.22.0.1"
            }
          ]
        ]
      },
      "dns": {
	"nameservers": ["10.22.0.1", "8.8.8.8"]
	}
    },
    {
      "type": "firewall"
    },
    {
      "type": "portmap",
      "capabilities": {"portMappings": true},
      "snat": true
    }
  ]
}
`

type NetworkManager struct {
	cni       gocni.CNI
	netNSBase string
}

func NewNetworkManager() (*NetworkManager, error) {
	cni, err := gocni.New(
		gocni.WithMinNetworkCount(2),
		gocni.WithPluginConfDir("/etc/cni/net.d"),
		gocni.WithPluginDir([]string{"/opt/cni/bin"}),
		gocni.WithInterfacePrefix("eth"),
		gocni.WithConfListBytes([]byte(conflist)))
	if err != nil {
		return nil, err
	}

	return &NetworkManager{cni: cni, netNSBase: "/proc/%d/ns/net"}, nil
}

func (r *NetworkManager) getNetworkNameAndNS(pid uint32) (string, string) {
	return fmt.Sprintf(r.netNSBase, pid), fmt.Sprintf("%d", pid)
}

func (r *NetworkManager) Attach(ctx context.Context, pid uint32, labels map[string]string) error {
	if err := r.cni.Load(
		gocni.WithLoNetwork,
		gocni.WithDefaultConf); err != nil {
		log.Fatalf("failed to load CNI configuration: %v", err)
	}

	netNS, name := r.getNetworkNameAndNS(pid)

	_, err := r.cni.Setup(ctx, name, netNS,
		gocni.WithLabels(labels))
	if err != nil {
		log.Fatalf("failed to attach network to container: %v", err)
		return err
	}

	return nil
}

func (r *NetworkManager) Detach(ctx context.Context, pid uint32, labels map[string]string) error {
	netNS, name := r.getNetworkNameAndNS(pid)

	return r.cni.Remove(ctx, name, netNS,
		gocni.WithLabels(labels))
}
