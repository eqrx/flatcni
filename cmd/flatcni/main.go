package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"go.eqrx.net/flatcni/internal/cni"
	"go.eqrx.net/flatcni/internal/network"
)

const (
	commandEnv         = "CNI_COMMAND"
	containerIDEnv     = "CNI_CONTAINERID"
	namespaceNameEnv   = "CNI_NETNS"
	interfaceNameEnv   = "CNI_IFNAME"
	podInterfacePrefix = "pod"
	tmpInterfacePrefix = "tmp"
)

func add(stdout *json.Encoder, cniConfig cni.Configuration) {
	var request cni.AddRequest
	if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil {
		_ = stdout.Encode(cni.NewErrorResponse(4, fmt.Sprintf("add request is invalid: %v", err)))
		os.Exit(1)
	}
	_, subnet, err := net.ParseCIDR(request.Subnet)
	if err != nil {
		_ = stdout.Encode(cni.NewErrorResponse(4, fmt.Sprintf("could not parse subnet: %v", err)))
		os.Exit(1)
	}
	cniConfig.Subnet = *subnet
	cniConfig.TargetNet = net.CIDRMask(request.TargetNet, 128)

	cfg, err := network.NewAddConfiguration(cniConfig)
	if err != nil {
		_ = stdout.Encode(cni.NewErrorResponse(0, err.Error()))
		os.Exit(1)
	}

	if err := network.SetupContainerNetwork(&cfg); err != nil {
		_ = stdout.Encode(cni.NewErrorResponse(0, err.Error()))
		os.Exit(1)
	}

	response := cni.AddResponse{
		CNIVersion: cni.Version,
		Interfaces: []cni.Interface{{Name: cfg.InterfaceName, MAC: cfg.MAC, Sandbox: cfg.ContainerID}},
		IPs:        []cni.IP{{Version: "6", Address: cfg.InnerAddress.String(), Gateway: cfg.OuterAddress.IP.String(), Interface: 0}},
	}

	_ = stdout.Encode(response)
}

func main() {
	stdout := json.NewEncoder(os.Stdout)
	command, _ := os.LookupEnv(commandEnv)
	switch command {
	case "VERSION":
		_ = stdout.Encode(cni.VersionResponse{
			CNIVersion:        cni.Version,
			SupportedVersions: []string{cni.Version}},
		)
		os.Exit(0)
	case "ADD":
	case "DEL":
	case "CHECK":
	default:
		_ = stdout.Encode(cni.NewErrorResponse(4, fmt.Sprintf("%s is invalid", commandEnv)))
		os.Exit(1)
	}

	containerID, envSet := os.LookupEnv(containerIDEnv)
	if !envSet {
		_ = stdout.Encode(cni.NewErrorResponse(4, fmt.Sprintf("%s is not set", containerIDEnv)))
		os.Exit(1)
	}
	interfaceName, envSet := os.LookupEnv(interfaceNameEnv)
	if !envSet {
		_ = stdout.Encode(cni.NewErrorResponse(4, fmt.Sprintf("%s is not set", interfaceNameEnv)))
		os.Exit(1)
	}
	sandboxPath, envSet := os.LookupEnv(namespaceNameEnv)
	if !envSet {
		_ = stdout.Encode(cni.NewErrorResponse(4, fmt.Sprintf("%s is not set", namespaceNameEnv)))
		os.Exit(1)
	}
	cniConfig := cni.Configuration{
		ContainerID:       containerID,
		InterfaceName:     interfaceName,
		SandboxPath:       sandboxPath,
		HostInterfaceName: podInterfacePrefix + containerID[:16-len(podInterfacePrefix)],
		TmpInterfaceName:  tmpInterfacePrefix + containerID[:16-len(tmpInterfacePrefix)],
	}

	switch command {
	case "ADD":
		add(stdout, cniConfig)
	case "DEL":
		network.DestroyContainerNetwork(cniConfig)
	case "CHECK":
	}
}
