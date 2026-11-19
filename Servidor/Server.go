package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
)

func claveGenerator(semilla string) string {
	hash := sha256.Sum256([]byte(semilla))
	return hex.EncodeToString(hash[:])
}

func main() {
	// Escuchar en el puerto 9090
	listener, err := net.Listen("tcp", "127.0.0.1:9090")
	if err != nil {
		fmt.Println("Error al abrir el puerto:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Esperando conexión en el puerto 9090...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		ip, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		if err != nil {
			fmt.Println("Error al obtener la IP del cliente:", err)
			conn.Close()
			continue
		}
		go func() {
			defer conn.Close()

			_, _ = io.WriteString(conn, claveGenerator(ip)+"\n")
		}()
		fmt.Println("La IP: ", conn.RemoteAddr().String(), " ha solicitado la clave de cifrado:", claveGenerator(ip))
	}
}
