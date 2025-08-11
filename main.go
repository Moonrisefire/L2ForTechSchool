package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

func main() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)

	scanner := bufio.NewScanner(os.Stdin)

	var currentCmds []*exec.Cmd

	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			fmt.Println("\nexit")
			return
		}

		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		go func() {
			<-sigc
			for _, c := range currentCmds {
				if c.Process != nil {
					_ = c.Process.Signal(syscall.SIGINT)
				}
			}
		}()

		status := runLine(line, &currentCmds)
		if status < 0 {
			return
		}
	}
}

func runLine(line string, currentCmds *[]*exec.Cmd) int {

	parts := splitByLogicalOperators(line)

	lastStatus := 0
	for i, part := range parts {

		if i > 0 {
			if parts[i-1].op == "&&" && lastStatus != 0 {
				continue
			}
			if parts[i-1].op == "||" && lastStatus == 0 {
				continue
			}
		}

		status := runPipeline(part.cmd, currentCmds)
		lastStatus = status
	}

	return lastStatus
}

type cmdPart struct {
	cmd string
	op  string
}

func splitByLogicalOperators(line string) []cmdPart {
	var res []cmdPart

	line = strings.TrimSpace(line)
	for len(line) > 0 {
		var idxAnd = strings.Index(line, "&&")
		var idxOr = strings.Index(line, "||")

		var idx int
		var op string
		if idxAnd == -1 && idxOr == -1 {
			res = append(res, cmdPart{cmd: strings.TrimSpace(line), op: ""})
			break
		} else if idxAnd != -1 && (idxOr == -1 || idxAnd < idxOr) {
			idx = idxAnd
			op = "&&"
		} else {
			idx = idxOr
			op = "||"
		}

		res = append(res, cmdPart{cmd: strings.TrimSpace(line[:idx]), op: op})
		line = strings.TrimSpace(line[idx+2:])
	}
	return res
}

func runPipeline(line string, currentCmds *[]*exec.Cmd) int {
	commands := splitByPipe(line)
	n := len(commands)

	var cmds []*exec.Cmd
	var pipes []io.ReadCloser
	var writers []io.WriteCloser

	for _, cmdstr := range commands {
		args := parseArgs(cmdstr)
		if len(args) == 0 {
			return 1
		}
		for i, a := range args {
			args[i] = substituteEnv(a)
		}

		if isBuiltin(args[0]) && n == 1 {
			return runBuiltin(args)
		}

		cmd := exec.Command(args[0], args[1:]...)
		cmds = append(cmds, cmd)
	}

	for i := 0; i < n-1; i++ {
		r, w, err := os.Pipe()
		if err != nil {
			fmt.Fprintln(os.Stderr, "pipe error:", err)
			return 1
		}
		pipes = append(pipes, r)
		writers = append(writers, w)
	}

	for i, cmd := range cmds {
		if i > 0 {
			cmd.Stdin = pipes[i-1]
		} else {
			cmd.Stdin = os.Stdin
		}
		if i < n-1 {
			cmd.Stdout = writers[i]
		} else {
			cmd.Stdout = os.Stdout
		}
		cmd.Stderr = os.Stderr
	}

	for _, cmd := range cmds {
		err := cmd.Start()
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to start:", err)
			return 1
		}
	}

	for _, w := range writers {
		w.Close()
	}
	for _, r := range pipes {
		r.Close()
	}

	*currentCmds = cmds

	status := 0
	for _, cmd := range cmds {
		err := cmd.Wait()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				status = exitErr.ExitCode()
			} else {
				status = 1
			}
		}
	}

	*currentCmds = nil
	return status
}

func splitByPipe(line string) []string {
	return strings.Split(line, "|")
}

func parseArgs(cmd string) []string {
	cmd = strings.TrimSpace(cmd)
	return strings.Fields(cmd)
}

func substituteEnv(s string) string {
	if strings.HasPrefix(s, "$") && len(s) > 1 {
		return os.Getenv(s[1:])
	}
	return s
}

func isBuiltin(cmd string) bool {
	switch cmd {
	case "cd", "pwd", "echo", "kill", "ps", "exit":
		return true
	}
	return false
}

func runBuiltin(args []string) int {
	switch args[0] {
	case "cd":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "cd: missing argument")
			return 1
		}
		err := os.Chdir(args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, "cd:", err)
			return 1
		}
		return 0

	case "pwd":
		dir, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, "pwd:", err)
			return 1
		}
		fmt.Println(dir)
		return 0

	case "echo":
		fmt.Println(strings.Join(args[1:], " "))
		return 0

	case "kill":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "kill: missing pid")
			return 1
		}
		pid, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, "kill: invalid pid")
			return 1
		}
		process, err := os.FindProcess(pid)
		if err != nil {
			fmt.Fprintln(os.Stderr, "kill:", err)
			return 1
		}
		err = process.Kill()
		if err != nil {
			fmt.Fprintln(os.Stderr, "kill:", err)
			return 1
		}
		return 0

	case "ps":
		cmd := exec.Command("ps")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, "ps:", err)
			return 1
		}
		return 0

	case "exit":
		os.Exit(0)
	}

	return 1
}
