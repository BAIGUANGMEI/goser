package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/BAIGUANGMEI/goser/internal/client"
	"github.com/BAIGUANGMEI/goser/internal/config"
	"github.com/BAIGUANGMEI/goser/internal/model"
)

var cli *client.Client

func main() {
	cli = client.NewDefault()

	rootCmd := &cobra.Command{
		Use:   "goser",
		Short: "GoSer - Go Service Manager",
		Long:  "A non-blocking service manager for Windows, similar to systemd.",
	}

	// --- daemon commands ---
	daemonCmd := &cobra.Command{
		Use:   "daemon",
		Short: "Manage the GoSer daemon",
	}

	daemonCmd.AddCommand(
		&cobra.Command{
			Use:   "start",
			Short: "Start the daemon in background",
			RunE:  daemonStart,
		},
		&cobra.Command{
			Use:   "stop",
			Short: "Stop the daemon",
			RunE:  daemonStop,
		},
		&cobra.Command{
			Use:   "status",
			Short: "Check daemon status",
			RunE:  daemonStatus,
		},
	)

	// --- service commands ---
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all services with status",
		RunE:  listServices,
	}

	startCmd := &cobra.Command{
		Use:   "start <name>",
		Short: "Start a service",
		Args:  cobra.ExactArgs(1),
		RunE:  startService,
	}

	stopCmd := &cobra.Command{
		Use:   "stop <name>",
		Short: "Stop a service",
		Args:  cobra.ExactArgs(1),
		RunE:  stopService,
	}

	restartCmd := &cobra.Command{
		Use:   "restart <name>",
		Short: "Restart a service",
		Args:  cobra.ExactArgs(1),
		RunE:  restartService,
	}

	statusCmd := &cobra.Command{
		Use:   "status <name>",
		Short: "Get detailed service status",
		Args:  cobra.ExactArgs(1),
		RunE:  serviceStatus,
	}

	addCmd := &cobra.Command{
		Use:   "add <yaml-file>",
		Short: "Add a service from YAML file",
		Args:  cobra.ExactArgs(1),
		RunE:  addService,
	}

	removeCmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a service",
		Args:  cobra.ExactArgs(1),
		RunE:  removeService,
	}

	enableCmd := &cobra.Command{
		Use:   "enable <name>",
		Short: "Enable auto-start for a service",
		Args:  cobra.ExactArgs(1),
		RunE:  enableService,
	}

	disableCmd := &cobra.Command{
		Use:   "disable <name>",
		Short: "Disable auto-start for a service",
		Args:  cobra.ExactArgs(1),
		RunE:  disableService,
	}

	logsCmd := &cobra.Command{
		Use:   "logs <name>",
		Short: "View service logs",
		Args:  cobra.ExactArgs(1),
		RunE:  viewLogs,
	}
	logsCmd.Flags().IntP("lines", "n", 50, "Number of lines to show")
	logsCmd.Flags().BoolP("follow", "f", false, "Follow log output (not yet implemented)")

	rootCmd.AddCommand(daemonCmd, listCmd, startCmd, stopCmd, restartCmd, statusCmd, addCmd, removeCmd, enableCmd, disableCmd, logsCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// --- Daemon commands ---

func daemonStart(cmd *cobra.Command, args []string) error {
	// Check if already running
	_, err := cli.DaemonStatus()
	if err == nil {
		fmt.Println("Daemon is already running.")
		return nil
	}

	// Find the goserd binary
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	daemonPath := filepath.Join(filepath.Dir(exePath), "goserd.exe")
	if _, err := os.Stat(daemonPath); os.IsNotExist(err) {
		// Try in current directory
		daemonPath = "goserd.exe"
	}

	// Start daemon in background
	process := exec.Command(daemonPath)
	process.Stdout = nil
	process.Stderr = nil
	if err := process.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	fmt.Printf("Daemon started (PID: %d)\n", process.Process.Pid)
	// Detach
	_ = process.Process.Release()

	// Wait briefly and verify
	time.Sleep(time.Second)
	status, err := cli.DaemonStatus()
	if err != nil {
		fmt.Println("Warning: daemon may not have started correctly")
		return nil
	}
	fmt.Printf("Daemon is running. Services: %d total, %d running\n", status.ServiceCount, status.RunningCount)
	return nil
}

func daemonStop(cmd *cobra.Command, args []string) error {
	// Read PID file and send signal
	cfg := config.DefaultGlobalConfig()
	data, err := os.ReadFile(cfg.Daemon.PIDFile)
	if err != nil {
		return fmt.Errorf("daemon does not appear to be running (no PID file)")
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return fmt.Errorf("invalid PID file")
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("daemon process not found")
	}

	if err := process.Kill(); err != nil {
		return fmt.Errorf("failed to stop daemon: %w", err)
	}

	_ = os.Remove(cfg.Daemon.PIDFile)
	fmt.Println("Daemon stopped.")
	return nil
}

func daemonStatus(cmd *cobra.Command, args []string) error {
	status, err := cli.DaemonStatus()
	if err != nil {
		fmt.Println("Daemon is NOT running.")
		return nil
	}

	fmt.Println("Daemon Status:")
	fmt.Printf("  Running:  yes\n")
	fmt.Printf("  Uptime:   %s\n", status.Uptime)
	fmt.Printf("  Services: %d total, %d running, %d stopped, %d failed\n",
		status.ServiceCount, status.RunningCount, status.StoppedCount, status.FailedCount)
	return nil
}

// --- Service commands ---

func listServices(cmd *cobra.Command, args []string) error {
	services, err := cli.ListServices()
	if err != nil {
		return err
	}

	if len(services) == 0 {
		fmt.Println("No services configured.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tPID\tUPTIME\tRESTARTS\tCOMMAND")
	for _, svc := range services {
		pid := "-"
		if svc.PID > 0 {
			pid = strconv.Itoa(svc.PID)
		}
		uptime := "-"
		if svc.Uptime != "" {
			uptime = svc.Uptime
		}
		cmdStr := svc.Command
		if len(svc.Args) > 0 {
			cmdStr += " " + strings.Join(svc.Args, " ")
		}
		if len(cmdStr) > 40 {
			cmdStr = cmdStr[:37] + "..."
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
			svc.Name, colorState(svc.State), pid, uptime, svc.RestartCount, cmdStr)
	}
	w.Flush()
	return nil
}

func startService(cmd *cobra.Command, args []string) error {
	if err := cli.StartService(args[0]); err != nil {
		return err
	}
	fmt.Printf("Service '%s' started.\n", args[0])
	return nil
}

func stopService(cmd *cobra.Command, args []string) error {
	if err := cli.StopService(args[0]); err != nil {
		return err
	}
	fmt.Printf("Service '%s' stopped.\n", args[0])
	return nil
}

func restartService(cmd *cobra.Command, args []string) error {
	if err := cli.RestartService(args[0]); err != nil {
		return err
	}
	fmt.Printf("Service '%s' restarted.\n", args[0])
	return nil
}

func serviceStatus(cmd *cobra.Command, args []string) error {
	info, err := cli.GetService(args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Service: %s\n", info.Name)
	fmt.Printf("  Status:      %s\n", colorState(info.State))
	fmt.Printf("  Command:     %s %s\n", info.Command, strings.Join(info.Args, " "))
	if info.WorkingDir != "" {
		fmt.Printf("  Working Dir: %s\n", info.WorkingDir)
	}
	if info.PID > 0 {
		fmt.Printf("  PID:         %d\n", info.PID)
	}
	if info.Uptime != "" {
		fmt.Printf("  Uptime:      %s\n", info.Uptime)
	}
	if info.StartedAt != nil {
		fmt.Printf("  Started At:  %s\n", info.StartedAt.Format(time.RFC3339))
	}
	if info.StoppedAt != nil {
		fmt.Printf("  Stopped At:  %s\n", info.StoppedAt.Format(time.RFC3339))
	}
	fmt.Printf("  Auto Start:  %v\n", info.AutoStart)
	fmt.Printf("  Auto Restart:%v\n", info.AutoRestart)
	fmt.Printf("  Restarts:    %d\n", info.RestartCount)
	if info.ExitCode != nil {
		fmt.Printf("  Exit Code:   %d\n", *info.ExitCode)
	}
	if info.Error != "" {
		fmt.Printf("  Error:       %s\n", info.Error)
	}
	return nil
}

func addService(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var svc config.ServiceConfig
	if err := yaml.Unmarshal(data, &svc); err != nil {
		return fmt.Errorf("parse yaml: %w", err)
	}

	if err := cli.CreateService(&svc); err != nil {
		return err
	}

	fmt.Printf("Service '%s' added.\n", svc.Name)
	return nil
}

func removeService(cmd *cobra.Command, args []string) error {
	if err := cli.DeleteService(args[0]); err != nil {
		return err
	}
	fmt.Printf("Service '%s' removed.\n", args[0])
	return nil
}

func enableService(cmd *cobra.Command, args []string) error {
	info, err := cli.GetService(args[0])
	if err != nil {
		return err
	}
	svc := &config.ServiceConfig{
		Name:        info.Name,
		Command:     info.Command,
		Args:        info.Args,
		WorkingDir:  info.WorkingDir,
		Env:         info.Env,
		AutoStart:   true,
		AutoRestart: info.AutoRestart,
	}
	if err := cli.UpdateService(args[0], svc); err != nil {
		return err
	}
	fmt.Printf("Service '%s' enabled for auto-start.\n", args[0])
	return nil
}

func disableService(cmd *cobra.Command, args []string) error {
	info, err := cli.GetService(args[0])
	if err != nil {
		return err
	}
	svc := &config.ServiceConfig{
		Name:        info.Name,
		Command:     info.Command,
		Args:        info.Args,
		WorkingDir:  info.WorkingDir,
		Env:         info.Env,
		AutoStart:   false,
		AutoRestart: info.AutoRestart,
	}
	if err := cli.UpdateService(args[0], svc); err != nil {
		return err
	}
	fmt.Printf("Service '%s' disabled for auto-start.\n", args[0])
	return nil
}

func viewLogs(cmd *cobra.Command, args []string) error {
	n, _ := cmd.Flags().GetInt("lines")
	logs, err := cli.GetLogs(args[0], n)
	if err != nil {
		return err
	}

	if len(logs) == 0 {
		fmt.Println("No logs available.")
		return nil
	}

	for _, entry := range logs {
		ts := entry.Timestamp.Format("15:04:05")
		stream := entry.Stream
		if stream == "stderr" {
			stream = "ERR"
		} else {
			stream = "OUT"
		}
		fmt.Printf("[%s] [%s] %s\n", ts, stream, entry.Line)
	}
	return nil
}

// colorState adds ANSI color to state for terminal display.
func colorState(state model.ServiceState) string {
	switch state {
	case model.StateRunning:
		return "\033[32m" + string(state) + "\033[0m" // green
	case model.StateStopped:
		return "\033[90m" + string(state) + "\033[0m" // gray
	case model.StateFailed:
		return "\033[31m" + string(state) + "\033[0m" // red
	case model.StateStarting, model.StateStopping:
		return "\033[33m" + string(state) + "\033[0m" // yellow
	default:
		return string(state)
	}
}
