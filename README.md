# Practica-Ransomware

Proyecto académico en Go para estudiar, dentro de un laboratorio controlado,
el comportamiento de un ransomware de prueba y su flujo de recuperación.

> [!CAUTION]
> **Este proyecto es destructivo.** Algunas variantes recorren las unidades
> de Windows, eliminan Shadow Copies, sobrescriben archivos con contenido
> cifrado y crean notas de rescate. No lo ejecutes en tu equipo principal, en
> una red corporativa ni sobre datos reales. El autor no se responsabiliza de
> pérdida de información, indisponibilidad del sistema o daños derivados de
> un uso fuera de un entorno autorizado.

## Objetivos didácticos

- Identificar las fases típicas de un ransomware en un entorno aislado.
- Practicar recorrido de sistemas de archivos en Windows.
- Observar el uso de AES-GCM y nonces aleatorios.
- Separar una variante de cifrado de su correspondiente herramienta de
  descifrado.
- Probar una comunicación HTTP local para obtener una clave de laboratorio.
- Analizar indicadores de compromiso y limitaciones de una implementación
  experimental.

## Arquitectura

El repositorio contiene **varios programas independientes**, cada uno con su
propia función `main`. Por ello, los archivos `.go` se compilan por separado;
no se debe ejecutar `go build .` sobre todo el directorio.

```text
                          ┌────────────────────┐
                          │     Server.go      │
                          │ HTTP :9090 /clave  │
                          │ devuelve ByClaude  │
                          └─────────┬──────────┘
                                    │ HTTP local
                                    ▼
┌──────────────────────┐   ┌──────────────────────┐
│      Virus.go        │   │  Virus_Reverse.go    │
│ Variante de prueba   │   │ Variante inversa     │
│ destructiva          │   │ de recuperación      │
└──────────┬───────────┘   └──────────┬───────────┘
           │                          │
           ▼                          ▼
  Busca y cifra archivos       Busca .encrypted
  Crea nota de rescate         Descifra y renombra
  Muestra GIF/audio            archivos originales

                    ┌──────────────────────┐
                    │      Vacuna.go       │
                    │ Descifrado local     │
                    │ sin servidor HTTP    │
                    └──────────────────────┘
```

### Componentes

| Archivo | Responsabilidad | Dependencias |
|---|---|---|
| [`Virus.go`](./Virus.go) | Variante principal: elimina Shadow Copies, busca extensiones de usuario, cifra archivos, crea notas de rescate y muestra una ventana multimedia. | Windows, `vssadmin`, PowerShell y recursos de `media/`. |
| [`Virus_Reverse.go`](./Virus_Reverse.go) | Variante principal con obtención de clave mediante HTTP local. Realiza las mismas fases destructivas y consulta `Server.go`. | Windows, servidor local en `:9090`. |
| [`Vacuna.go`](./Vacuna.go) | Busca archivos `.encrypted`, los descifra con la clave fija de laboratorio y elimina la extensión `.encrypted`. | Windows. |
| [`Server.go`](./Server.go) | Servidor HTTP mínimo que expone `GET /clave` y responde `ByClaude`. | Cualquier sistema con Go; está pensado para uso local. |
| [`media/`](./media/) | GIF y audio utilizados por `Virus.go` y `Virus_Reverse.go`. | Windows Forms, PowerShell y Windows Media Player. |
| [`recursos_Prueba/`](./recursos_Prueba/) | Material para preparar un laboratorio: archivos de prueba, una sandbox y archivos comprimidos de ejemplo. | Windows Sandbox para el `.wsb`. |
| [`Vacuna.exe`](./Vacuna.exe) | Binario precompilado incluido en el repositorio. | No sustituye a una compilación reproducible. |

## Flujo de ejecución

### Variante `Virus.go`

1. Ejecuta `vssadmin` para listar y eliminar las Shadow Copies.
2. Detecta las unidades de `A:` a `Z:` que estén disponibles.
3. Recorre las unidades y busca `.jpg`, `.jpeg`, `.gif`, `.docx`, `.xlsx`,
   `.pptx`, `.pdf` y `.png`.
4. Omite rutas que contienen nombres de directorios del sistema o de
   herramientas (`Windows`, `System32`, `Program Files`, `.git`, etc.).
5. Cifra cada archivo con AES-GCM y antepone el nonce al resultado.
6. Sobrescribe el archivo y añade la extensión `.encrypted`.
7. Crea `README.txt` en los directorios donde encontró archivos.
8. Abre una ventana con el GIF y reproduce el audio de `media/`.

### Variante `Virus_Reverse.go`

Tiene el mismo flujo destructivo, pero obtiene la clave con:

```text
GET http://localhost:9090/clave
```

Debe utilizarse junto con `Server.go`. Si el servidor no está disponible, la
variante no dispone de una clave válida para cifrar.

### Variante `Vacuna.go`

1. Busca archivos con extensión `.encrypted` en las unidades disponibles.
2. Extrae el nonce almacenado al principio de cada archivo.
3. Descifra con AES-GCM usando `ByClaude`.
4. Sobrescribe el contenido y elimina `.encrypted` del nombre.

La herramienta de recuperación solo puede funcionar con archivos generados
por una variante compatible y con la clave correcta. No reemplaza una copia de
seguridad.

## Criptografía utilizada

- Algoritmo: AES-GCM.
- Tamaño de clave usado por el código: 32 bytes (AES-256).
- Clave de laboratorio actual: `ByClaude`.
- Derivación actual: repetición de los bytes de la cadena hasta alcanzar
  32 bytes; **no** es una KDF segura.
- Formato de archivo cifrado: `nonce || ciphertext || tag`.
- Los archivos se cargan completos en memoria antes de procesarse.

La clave está incrustada o se transmite por HTTP sin autenticación según la
variante. Esto es intencionado únicamente para la práctica y no debe usarse
como diseño criptográfico real.

## Requisitos

- Go instalado y disponible en `PATH`.
- Windows para ejecutar `Virus.go`, `Virus_Reverse.go` o `Vacuna.go`.
- PowerShell, Windows Forms y Windows Media Player para la ventana multimedia.
- Permisos suficientes para las operaciones de laboratorio con `vssadmin`.
- Una máquina virtual o Windows Sandbox aislado, sin datos personales.

`Server.go` usa solo la biblioteca estándar y puede compilarse en otros
sistemas, aunque su cliente (`Virus_Reverse.go`) está diseñado para Windows.

## Compilación

Desde la raíz del repositorio. Los comandos generan ejecutables Windows
de 64 bits y siguen el formato de compilación cruzada solicitado:

```bash
GOOS=windows GOARCH=amd64 go build -o Server.exe Server.go
GOOS=windows GOARCH=amd64 go build -o Virus.exe Virus.go
GOOS=windows GOARCH=amd64 go build -o Virus_Reverse.exe Virus_Reverse.go
GOOS=windows GOARCH=amd64 go build -o Vacuna-Lab.exe Vacuna.go
```

Para comprobar la sintaxis de un único archivo en el sistema actual:

```bash
go build Server.go
```

No compiles todos los archivos juntos: contienen varias funciones `main` y
declaraciones duplicadas entre variantes.

## Ejecución segura en laboratorio

1. Crea una máquina virtual o Windows Sandbox **desechable**.
2. Desactiva carpetas compartidas, portapapeles compartido y unidades
   montadas del equipo anfitrión.
3. Desconecta la red o permite únicamente `localhost`.
4. Crea archivos de prueba sin valor dentro del entorno aislado.
5. Copia `media/` junto al ejecutable si vas a probar la ventana multimedia.
6. Compila y ejecuta una sola variante cada vez.
7. Conserva una instantánea de la máquina virtual antes de la prueba.
8. Restaura o destruye la máquina virtual al terminar.

### Prueba del flujo con servidor local

En una terminal de laboratorio, inicia primero el servidor:

```powershell
.\Server.exe
```

En otra terminal del mismo entorno, ejecuta la variante que consulta la clave:

```powershell
.\Virus_Reverse.exe
```

Para intentar la recuperación en el mismo entorno:

```powershell
.\Vacuna-Lab.exe
```

Estas instrucciones **no** convierten el programa en una herramienta de
recuperación segura: el código sobrescribe archivos y no crea copias de
seguridad.

## Indicadores observables

En un entorno de análisis, pueden observarse:

- Ejecución de `cmd.exe` y `vssadmin.exe`.
- Recorrido de letras de unidad y lectura masiva del sistema de archivos.
- Escrituras sobre documentos e imágenes.
- Renombrado de archivos a `.encrypted`.
- Creación de `README.txt` en directorios afectados.
- Peticiones HTTP locales a `/clave`.
- Ejecución de PowerShell y acceso a los recursos multimedia.


## Licencia y uso

No se declara una licencia en este repositorio. Hasta que se añada una,
considéralo código de uso restringido para fines educativos y de laboratorio.
No lo redistribuyas ni lo ejecutes contra sistemas sin autorización expresa.
