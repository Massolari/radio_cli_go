package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type Player struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	IsPlaying bool
}

func NewPlayer(url string) (*Player, error) {
	p := &Player{}

	p.cmd = exec.Command("vlc", "-I", "rc")

	var err error
	p.stdin, err = p.cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	if err := p.cmd.Start(); err != nil {
		return nil, err
	}

	err = p.Play(url)
	if err != nil {
		return nil, err
	}
	p.IsPlaying = true

	return p, nil
}

func (p *Player) Play(url string) error {
	_, err := p.stdin.Write([]byte("clear\n"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return err
	}
	_, err = p.stdin.Write([]byte(fmt.Sprintf("add %s\n", url)))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return err
	}

	return p.Resume()
}

func (p *Player) Resume() error {
	if p.cmd.ProcessState != nil && p.cmd.ProcessState.Exited() {
		fmt.Println("Process has already exited.")
		os.Exit(1)
		return errors.New("cannot resume: process has exited")
	}
	_, err := p.stdin.Write([]byte("play\n"))
	if err != nil {
		fmt.Println(p.cmd.ProcessState)
		fmt.Println(err)
		os.Exit(1)
		return err
	}
	p.IsPlaying = true
	return nil
}

func (p *Player) Stop() error {
	_, err := p.stdin.Write([]byte("stop\n"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return err
	}
	p.IsPlaying = false
	return nil
}

func (p Player) Quit() error {
	if err := p.stdin.Close(); err != nil {
		return err
	}
	return p.cmd.Wait()
}
