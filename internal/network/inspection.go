package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os/exec"
)

type networkInterfaceInfo struct {
	Addresses []struct {
		Family  string `json:"family"`
		Address string `json:"local"`
		Mask    int    `json:"prefixlen"`
	} `json:"addr_info"`
}

var (
	// ErrAddressPollExhaused mean that the given address pool does not contain any more
	// free IP addresses.
	ErrAddressPollExhaused = errors.New("address pool exhaused")
	// ErrInvalidMasterConfiguration means that the network configuration of the host is unexpected.
	ErrInvalidMasterConfiguration = errors.New("master interface is invalid")
)

func jsonFromIPCommand(destination interface{}, args ...string) error {
	cmd := exec.Command("ip", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, output)
	}
	return json.Unmarshal(output, destination)
}

// CurrentSubnets that are in use within the given parent prefix.
func CurrentSubnets(parentNet net.IPNet) ([]net.IPNet, error) {
	addresses := []net.IPNet{}
	networkInterfaceInfos := []networkInterfaceInfo{}
	if err := jsonFromIPCommand(&networkInterfaceInfos, "--json", "address", "show"); err != nil {
		return addresses, err
	}
	for _, interfaceInfo := range networkInterfaceInfos {
		for _, addressInfo := range interfaceInfo.Addresses {
			if addressInfo.Family != "inet6" {
				continue
			}
			address := net.ParseIP(addressInfo.Address)
			if !address.IsGlobalUnicast() {
				continue
			}
			if parentNet.Contains(address) {
				addresses = append(addresses, net.IPNet{IP: address, Mask: net.CIDRMask(addressInfo.Mask, 128)})
			}
		}
	}
	return addresses, nil
}

// PickAddressPair for one container.
func PickAddressPair(parentNet net.IPNet, targetMask net.IPMask) (inner net.IPNet, outer net.IPNet, err error) {
	inner.Mask = targetMask
	outer.Mask = targetMask
	subnets, err := CurrentSubnets(parentNet)
	isOccupied := func(address net.IPNet) bool {
		for _, c := range subnets {
			if address.IP.Equal(c.IP) {
				return true
			}
		}
		return false
	}
	ones, bits := targetMask.Size()

	subnetIncrement := big.NewInt(2 << (bits - ones))
	if err != nil {
		return
	}
	i := big.NewInt(0)
	i.SetBytes(parentNet.IP)
	for {
		i.Add(i, subnetIncrement)
		outer.IP = i.Bytes()
		if !isOccupied(outer) {
			i.Add(i, big.NewInt(1))
			inner.IP = i.Bytes()
			return
		}

		if !parentNet.Contains(inner.IP) {
			err = ErrAddressPollExhaused
			return
		}
	}
}
