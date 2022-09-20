package main

import (
	"os"
	"os/exec"
)

type Command interface {
	Start() error
	Finished() <-chan error
	Process() *os.Process
}

type AsyncCommand struct {
	cmd          *exec.Cmd
	finishedChan chan error
}

func NewAsyncCommand(cmd *exec.Cmd) *AsyncCommand {
	return &AsyncCommand{cmd: cmd, finishedChan: make(chan error)}
}

func (c *AsyncCommand) Start() error {
	err := c.cmd.Start()
	if err != nil {
		return err
	}
	go func() {
		c.finishedChan <- c.cmd.Wait()
	}()
	return nil
}

func (c *AsyncCommand) Process() *os.Process {
	return c.cmd.Process
}

func (c *AsyncCommand) Finished() <-chan error {
	return c.finishedChan
}
