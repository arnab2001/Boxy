package main

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "boxy",
	Short: "Tiny container CLI powered by containerd",
}

func Execute() { _ = rootCmd.Execute() }
