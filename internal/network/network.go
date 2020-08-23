// Package network interfaces with the linux network subsystem.
package network

import (
	"fmt"
	"net"
	"os"
	"os/exec"

	"go.eqrx.net/flatcni/internal/cni"
)

// AddConfiguration that contains required information for adding a network.
type AddConfiguration struct {
	cni.Configuration
	MAC          string
	InnerAddress net.IPNet
	OuterAddress net.IPNet
}

// NewAddConfiguration that contains a unique address and the gateway to set.
func NewAddConfiguration(c cni.Configuration) (a AddConfiguration, err error) {
	a.InnerAddress, a.OuterAddress, err = PickAddressPair(c.Subnet, c.TargetNet)
	return
}

func issueIPCommand(args ...string) error {
	cmd := exec.Command("ip", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, output)
	}
	return nil
}

// DestroyContainerNetwork of the container with the given ID..
func DestroyContainerNetwork(c cni.Configuration) {
	_ = issueIPCommand("link", "delete", c.HostInterfaceName)
	_ = os.Remove("/var/run/netns/" + c.ContainerID)
}

// SetupContainerNetwork with the given parameters.
func SetupContainerNetwork(c *AddConfiguration) (err error) {
	var symlinkCreated, interfaceCreated bool
	defer func() {
		if err != nil && symlinkCreated {
			_ = os.Remove("/var/run/netns/" + c.ContainerID)
		}
		if err != nil && interfaceCreated {
			_ = issueIPCommand("link", "remove", c.InterfaceName)
		}
	}()
	if err := os.Symlink(c.SandboxPath, "/var/run/netns/"+c.ContainerID); err != nil {
		return fmt.Errorf("could not create network namespace: %w", err)
	}
	if err := issueIPCommand("link", "add", c.HostInterfaceName, "type", "veth", "peer", c.TmpInterfaceName); err != nil {
		return fmt.Errorf("could not create veth interface %v, %v: %w", c.TmpInterfaceName, c.HostInterfaceName, err)
	}
	iface, err := net.InterfaceByName(c.HostInterfaceName)
	if err != nil {
		return fmt.Errorf("could not get veth interface: %w", err)
	}
	c.MAC = iface.HardwareAddr.String()
	if err := issueIPCommand("link", "set", c.TmpInterfaceName, "netns", c.ContainerID); err != nil {
		return fmt.Errorf("could not assign veth interface to namespace: %w", err)
	}

	if err := issueIPCommand("netns", "exec", c.ContainerID, "ip", "link", "set", c.TmpInterfaceName, "name", c.InterfaceName); err != nil {
		return fmt.Errorf("could not rename interface inside container to target name: %w", err)
	}
	if err := issueIPCommand("link", "set", c.HostInterfaceName, "up"); err != nil {
		return fmt.Errorf("could not put host interface up: %w", err)
	}
	if err := issueIPCommand("netns", "addr", "add", c.OuterAddress.String(), "dev", c.HostInterfaceName); err != nil {
		return fmt.Errorf("could not add address to container interface: %w", err)
	}

	if err := issueIPCommand("netns", "exec", c.ContainerID, "ip", "link", "set", c.InterfaceName, "up"); err != nil {
		return fmt.Errorf("could not put container interface up: %w", err)
	}
	if err := issueIPCommand("netns", "exec", c.ContainerID, "ip", "addr", "add", c.InnerAddress.String(), "dev", c.InterfaceName); err != nil {
		return fmt.Errorf("could not add address to container interface: %w", err)
	}
	if err := issueIPCommand("netns", "exec", c.ContainerID, "ip", "route", "add", "default", "via", c.OuterAddress.IP.String(), "dev", c.InterfaceName); err != nil {
		return fmt.Errorf("could not set gateway to container interface: %w", err)
	}

	return nil
}
