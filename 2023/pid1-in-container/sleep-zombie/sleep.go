package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

func main() {
	var reap bool
	flag.BoolVar(&reap, "reap", false, "whether to reap children")

	pid := os.Getpid()
	ppid := os.Getppid()
	fmt.Printf("pid: %d, ppid: %d\n", pid, ppid)

	if reap { // reap children
		go func() {
			sig := make(chan os.Signal, 1)
			signal.Notify(sig)
			for s := range sig {
				if s == syscall.SIGCHLD {
					var status unix.WaitStatus
					unix.Wait4(-1, &status, unix.WNOHANG, nil)
				}
			}
		}()
	}

	for i := 1; i <= 60; i++ {
		fmt.Println(pid, ".", i)
		if _, isChild := os.LookupEnv("CHILD_ID"); !isChild {
			pwd, err := os.Getwd()
			if err != nil {
				log.Fatalf("getwd err: %s", err)
			}
			args := append(os.Args, fmt.Sprintf("#child_%d_of_%d", i, os.Getpid()))
			childENV := []string{
				fmt.Sprintf("CHILD_ID=%d", i),
			}
			syscall.ForkExec(args[0], args, &syscall.ProcAttr{
				Dir: pwd,
				Env: append(os.Environ(), childENV...),
				Sys: &syscall.SysProcAttr{
					Setsid: true,
				},
				Files: []uintptr{0, 1, 2}, // print message to the same pty
			})
		} else {
			os.Exit(0) // child exit directly, become zombie
		}
		time.Sleep(time.Second)
	}
}
