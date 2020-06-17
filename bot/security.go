package bot

import (
	"net"

	"github.com/pkg/errors"
)

func newIPChecker(cidrs ...string) (func(ip string) bool, error) {
	networks := make([]*net.IPNet, len(cidrs))

	for i, cidr := range cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		networks[i] = network
	}

	return func(ipRaw string) bool {
		ip := net.ParseIP(ipRaw)

		for _, network := range networks {
			if network.Contains(ip) {
				return true
			}
		}

		return false
	}, nil
}

var isTelegramIP func(ip string) bool

func init() {
	var err error

	isTelegramIP, err = newIPChecker(
		"149.154.160.0/20",
		"91.108.4.0/22",
	)

	if err != nil {
		panic(errors.Wrap(err, "create ip checker"))
	}

}
