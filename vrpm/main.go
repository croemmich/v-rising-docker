package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/nxadm/tail"
)

const appId = "1829350"
const serverDir = "/mnt/vrising/server"
const persistentDir = "/mnt/vrising/persistent"

const serverLogPath = persistentDir + "/VRisingServer.log"
const serverPath = serverDir + "/VRisingServer.exe"

const steamcmdPath = "/usr/bin/steamcmd"
const xvfbPath = "/usr/bin/Xvfb"
const wine64Path = "/usr/bin/wine64"
const wineserverPath = "/usr/bin/wineserver"

func main() {
	// install/update/validate the server
	err := upstall()
	if err != nil {
		panic(err)
	}

	// run the server
	sigChan := make(chan os.Signal)
	finishedChan := run(sigChan)

	// wait for SIGINT or SIGTERM
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	fmt.Printf("Received %s signal\n", sig)

	// notify "run" that the server should stop
	sigChan <- sig

	// wait for the server to stop
	err = <-finishedChan

	if err != nil {
		panic(err)
	}
}

func upstall() error {
	cmd := exec.Command(steamcmdPath, "+@sSteamCmdForcePlatformType", "windows", "+force_install_dir", serverDir, "+login", "anonymous", "+app_update", appId, "validate", "+quit")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func makeXvfbCommand() Command {
	cmd := exec.CommandContext(context.Background(), xvfbPath, ":0", "-screen", "0", "1024x768x16", "-nolisten", "tcp", "-nolisten", "unix")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}
	return NewAsyncCommand(cmd)
}

func makeServerCommand() Command {
	cmd := exec.CommandContext(context.Background(), wine64Path, serverPath, "-persistentDataPath", persistentDir, "-logFile", serverLogPath)
	cmd.Env = append(os.Environ(), "DISPLAY=:0.0", "WINEDEBUG=error")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}
	return NewAsyncCommand(cmd)
}

func makeWineserverCommand() Command {
	cmd := exec.CommandContext(context.Background(), wineserverPath, "--wait")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}
	return NewAsyncCommand(cmd)
}

func run(sigChan chan os.Signal) <-chan error {
	finishedChan := make(chan error, 1)
	go func() {
		xvfbCmd := makeXvfbCommand()
		serverCmd := makeServerCommand()
		wineserverCmd := makeWineserverCommand()

		// Start Xvfb
		if err := xvfbCmd.Start(); err != nil {
			finishedChan <- err
			return
		}
		println("Started Xvfb")

		// Start the server
		if err := serverCmd.Start(); err != nil {
			finishedChan <- err
			return
		}
		println("Started server")

		// Tail the server log file
		println("Tail server logs...")
		t, err := tailServerLog()
		defer t.Stop()
		if err != nil {
			finishedChan <- err
			return
		}

		// Watch for commands to finish or a message on the sigChan
		for {
			select {
			case <-xvfbCmd.Finished():
				finishedChan <- fmt.Errorf("xvfb exited unexpectedly")
				return
			case err := <-serverCmd.Finished():
				if err != nil {
					finishedChan <- fmt.Errorf("server exited with error: %w", err)
					return
				}
				err = wineserverCmd.Start()
				if err != nil {
					println("Failed to wait for wineserver, waiting 30 seconds and exiting...")
					time.Sleep(time.Second * 30)
					finishedChan <- fmt.Errorf("failed to wait for wineserver: %w", err)
					return
				}
			case err := <-wineserverCmd.Finished():
				if err != nil {
					finishedChan <- fmt.Errorf("wineserver exited unexpectedly: %w", err)
				} else {
					close(finishedChan)
				}
				return
			case <-sigChan:
				err := serverCmd.Process().Signal(syscall.SIGINT)
				if err != nil {
					finishedChan <- fmt.Errorf("failed to signal server: %w", err)
					return
				}
			}
		}
	}()
	return finishedChan
}

func tailServerLog() (*tail.Tail, error) {
	t, err := tail.TailFile(serverLogPath, tail.Config{Follow: true, ReOpen: true, Poll: true})
	if err != nil {
		return nil, err
	}
	go func() {
		for line := range t.Lines {
			println(line.Text)
		}
	}()
	return t, nil
}
