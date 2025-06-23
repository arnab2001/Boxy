package main

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "boxy",
	Short: "Container runtime with Docker-style port publishing powered by containerd",
}

func Execute() { _ = rootCmd.Execute() }
