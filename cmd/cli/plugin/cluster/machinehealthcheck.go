package main

import (
	"github.com/spf13/cobra"
)

var machineHealthCheckCmd = &cobra.Command{
	Use:   "machinehealthcheck",
	Short: "Get,set, or delete a MachineHealthCheck object",
	Long:  `Get,set, or delete a MachineHealthCheck object for a Tanzu Kubernetes cluster`,
}