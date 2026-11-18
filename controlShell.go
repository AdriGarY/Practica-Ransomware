package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	// Escuchar en el puerto 9090
	listener, err := net.Listen("tcp", "127.0.0.1:9090")
	if err != nil {
		fmt.Println("Error al abrir el puerto:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Esperando conexión en el puerto 9090...")

	// Aceptar la conexión entrante de ShellReverse.go
	con, err := listener.Accept()
	if err != nil {
		fmt.Println("Error al aceptar conexión:", err)
		return
	}
	defer con.Close()
	fmt.Println("¡Cliente conectado! Escribe comandos y pulsa Enter:")

	// Goroutine en segundo plano para mostrar todo lo que responda la shell
	go func() {
		io.Copy(os.Stdout, con)
	}()

	// Leer comandos desde tu teclado y enviarlos al cliente
	teclado := bufio.NewReader(os.Stdin)
	for {
		comando, err := teclado.ReadString('\n')
		if err != nil {
			break
		}
		con.Write([]byte(comando))
	}
}
