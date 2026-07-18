package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"Sentinel/internal/render"
	"Sentinel/pkg/contextbuilder"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run local system checks for Sentinel",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(render.HeaderBox("🩺 SENTINEL DOCTOR"))

		uid := os.Getuid()
		socketPath := fmt.Sprintf("/tmp/sentinel_%d.sock", uid)
		allPassed := true

		fmt.Println(render.Section("Infrastructure Checks"))
		var infraLines []string

		conn, err := net.DialTimeout("unix", socketPath, 1*time.Second)
		if err != nil {
			allPassed = false
			infraLines = append(infraLines, render.Item(render.Error, "Sentinel Agent Daemon", fmt.Sprintf("Disconnected. Missing socket: %s", socketPath)))
		} else {
			conn.Close()
			infraLines = append(infraLines, render.Item(render.Success, "Sentinel Agent Daemon", "Connected successfully via UNIX socket."))
		}

		if uid == 0 {
			infraLines = append(infraLines, render.Item(render.Success, "User Permissions", "Running as root (Full infrastructure access)."))
		} else {
			infraLines = append(infraLines, render.Item(render.Success, "User Permissions", fmt.Sprintf("Running as regular user (UID: %d).", uid)))
		}
		fmt.Println(render.SubBlock(strings.Join(infraLines, "\n")))

		fmt.Println(render.Section("Environment Metadata"))
		metaData := fmt.Sprintf("%s\n%s\n%s",
			render.KeyValue("Socket Path", socketPath),
			render.KeyValue("User UID", strconv.Itoa(uid)),
			render.KeyValue("Collectors", strconv.Itoa(len(contextbuilder.Registry))),
		)
		fmt.Println(render.SubBlock(metaData))

		fmt.Println(render.Section("Registered Context Collectors"))
		var tableRows [][]string
		for name, collector := range contextbuilder.Registry {
			statusSymbol := "⚠ Disabled"
			if collector.Supports() {
				statusSymbol = "✓ Active"
			}
			tableRows = append(tableRows, []string{name, statusSymbol})
		}

		t := render.NewTable([]string{"Collector Specialist", "Status"}, tableRows)
		fmt.Println(render.SubBlock(t.Render()))

		if allPassed {
			fmt.Println(render.Banner(render.Success, " 🛡️  SENTINEL IS HEALTHY - READY TO DIAGNOSE PRODUCTION INCIDENTS "))
		} else {
			fmt.Println(render.Banner(render.Error, " ⚠️  SENTINEL DOCTOR DETECTED ENVIRONMENT ISSUES - CHECK LOGS ABOVE "))
			os.Exit(1)
		}
	},
}
