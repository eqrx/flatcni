// Package cni contains types that are used to interface with the CNI caller.
package cni

import "net"

const (
	// Version of the supported CNI.
	Version = "0.4.0"
)

// VersionResponse when queried for it.
type VersionResponse struct {
	CNIVersion        string   `json:"cniVersion"`
	SupportedVersions []string `json:"supportedVersions"`
}

// ErrorResponse when an error arises.
type ErrorResponse struct {
	CNIVersion string `json:"cniVersion"`
	Code       int    `json:"code"`
	Message    string `json:"msg"`
}

// NewErrorResponse creates a new ErrorResponse.
func NewErrorResponse(code int, message string) ErrorResponse {
	return ErrorResponse{Version, code, message}
}

// IP definition of the response to an ADD command.
type IP struct {
	Version   string `json:"version"`
	Address   string `json:"address"`
	Gateway   string `json:"gateway"`
	Interface int    `json:"interface"`
}

// Interface definition of the response the an ADD command.
type Interface struct {
	Name    string `json:"name"`
	MAC     string `json:"mac"`
	Sandbox string `json:"sandbox"`
}

// AddResponse is the data written out to stdout for an ADD command.
type AddResponse struct {
	CNIVersion string      `json:"cniVersion"`
	Interfaces []Interface `json:"interfaces"`
	IPs        []IP        `json:"ips"`
}

// AddRequest contains the configuration needed for the ADD command.
type AddRequest struct {
	Subnet    string `json:"subnet"`
	TargetNet int    `json:"targetNet"`
}

// Configuration needed for network management.
type Configuration struct {
	Subnet            net.IPNet
	TargetNet         net.IPMask
	ContainerID       string
	InterfaceName     string
	SandboxPath       string
	HostInterfaceName string
	TmpInterfaceName  string
}
