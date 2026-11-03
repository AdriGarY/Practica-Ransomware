# PruebaClaude.go

## Descripción

`PruebaClaude.go` es un programa de prueba escrito en Go para Windows que
simula varias fases de un ransomware:

1. Elimina todas las copias Shadow del sistema.
2. Busca archivos de usuario con determinadas extensiones.
3. Cifra los archivos encontrados.
4. Sobrescribe los archivos originales con su contenido cifrado.
5. Muestra por consola un resumen del proceso.

El programa es destructivo: elimina mecanismos de recuperación del sistema y
modifica directamente los archivos encontrados. Debe utilizarse únicamente en
un laboratorio, una máquina virtual o un entorno expresamente autorizado.

## Requisitos y compatibilidad

- Sistema operativo Windows.
- Go instalado.
- Permisos suficientes para ejecutar `vssadmin`.
- Permisos de lectura y escritura sobre los archivos que se vayan a procesar.

El código no es multiplataforma. Utiliza letras de unidad de Windows (`A:` a
`Z:`), rutas con `\` y el comando `vssadmin`, por lo que no funcionará
correctamente en Linux o macOS.

## Librerías utilizadas

El programa utiliza únicamente librerías estándar de Go:

- `crypto/aes`: crea el cifrador AES.
- `crypto/cipher`: construye el modo autenticado GCM.
- `crypto/rand`: genera el nonce aleatorio.
- `fmt`: muestra mensajes, contadores y errores.
- `io`: permite rellenar el nonce con datos aleatorios.
- `os`: lee, escribe y consulta archivos y unidades.
- `os/exec`: ejecuta comandos externos de Windows.
- `path/filepath`: recorre directorios y obtiene extensiones y rutas.
- `strings`: normaliza extensiones y comprueba directorios excluidos.

No hay dependencias externas ni módulos adicionales.

## Flujo de ejecución

El flujo comienza en `main`:

```text
main
 ├── deleteShadowCopies()
 │    ├── Lista las copias Shadow
 │    └── Elimina todas las copias Shadow
 ├── searchFiles()
 │    ├── Detecta las unidades disponibles
 │    ├── Recorre sus directorios
 │    └── Devuelve los archivos seleccionados
 ├── printResults()
 │    └── Muestra la lista de archivos encontrados
 └── encryptFiles()
      └── Cifra y sobrescribe cada archivo encontrado
```

Si falla la eliminación de copias Shadow o la búsqueda de unidades, el
programa muestra el error y termina con código `1`. Si falla el cifrado de uno
o más archivos, continúa procesando los demás y finalmente también devuelve
un error.

## 1. Eliminación de copias Shadow

La función `deleteShadowCopies` ejecuta dos comandos:

```text
cmd /C vssadmin list shadows
cmd /C vssadmin delete shadows /all /quiet
```

Primero lista y muestra las copias existentes. Después solicita a Windows la
eliminación de **todas** ellas:

- `/all` selecciona todas las copias Shadow.
- `/quiet` evita las confirmaciones interactivas.

Esta acción es irreversible desde el programa y puede impedir la recuperación
de versiones anteriores de los archivos o del sistema.

## 2. Búsqueda de archivos

La función `searchFiles` busca archivos en todas las unidades detectadas. Las
extensiones configuradas son:

| Extensión | Tipo habitual |
|---|---|
| `.jpg` | Imagen JPEG |
| `.docx` | Documento de Word |
| `.xlsx` | Hoja de cálculo de Excel |
| `.pptx` | Presentación de PowerPoint |
| `.pdf` | Documento PDF |
| `.png` | Imagen PNG |

El recorrido se realiza con `filepath.Walk`. Para cada ruta:

1. Ignora los errores de acceso y continúa con el siguiente elemento.
2. Omite ciertos directorios del sistema.
3. Comprueba la extensión en minúsculas.
4. Convierte la ruta a absoluta.
5. Añade el archivo a la lista `foundFiles`.

Los directorios que se intentan excluir son:

```text
System Volume Information
$Recycle.Bin
pagefile.sys
hiberfil.sys
ProgramData
AppData
Temp
.git
node_modules
```

La exclusión se basa en comprobar si el nombre aparece dentro de la ruta
completa, sin distinguir entre mayúsculas y minúsculas.

## 3. Cifrado de archivos

La función `encryptFiles` recibe las rutas encontradas y utiliza la clave de
texto fija:

```text
ByClaude
```

Después llama a `expandKey` para convertirla en una clave de 32 bytes, el
tamaño utilizado por AES-256. La función completa cada posición restante
repitiendo los bytes de la clave original.

Para cada archivo:

1. Muestra el progreso en la consola.
2. Lee el archivo completo en memoria mediante `os.ReadFile`.
3. Crea un cifrador AES.
4. Construye AES-GCM.
5. Genera un nonce aleatorio.
6. Cifra el contenido con `gcm.Seal`.
7. Guarda el nonce seguido del texto cifrado.
8. Sobrescribe el archivo original mediante `os.WriteFile`.

El nonce se antepone al resultado cifrado porque será necesario para intentar
descifrar el contenido posteriormente. El modo GCM también incluye
autenticación de integridad, de modo que detecta modificaciones del texto
cifrado.

## 4. Resultados y errores

`encryptFiles` mantiene dos contadores:

- `successCount`: archivos cifrados correctamente.
- `failCount`: archivos que no se pudieron cifrar.

Si no hay archivos, muestra un mensaje y termina sin realizar operaciones de
cifrado. Si uno falla, el programa informa del error y continúa con el
siguiente. Al acabar, si `failCount` es mayor que cero, devuelve un error.

## Funciones auxiliares

### `getDrives`

Comprueba las letras de unidad de `A` a `Z` y conserva las que tienen una raíz
accesible, como `C:\` o `D:\`.

### `isSystemDirectory`

Devuelve `true` cuando la ruta contiene alguno de los nombres configurados
como directorios del sistema o directorios que pueden ralentizar el recorrido.
Cuando devuelve `true`, `filepath.Walk` utiliza `filepath.SkipDir`.

### `expandKey`

Crea un slice del tamaño solicitado, copia la clave recibida y repite sus
bytes hasta rellenar el tamaño final. En este programa se usa para producir
una clave AES de 32 bytes.

### `printResults`

Imprime el número total de archivos localizados y enumera sus rutas. No crea
ningún archivo de resultados.

## Compilación y ejecución

Desde PowerShell:

```powershell
go build -o PruebaClaude.exe .\PruebaClaude.go
.\PruebaClaude.exe
```

O directamente:

```powershell
go run .\PruebaClaude.go
```

Es posible que sea necesario ejecutar la terminal como administrador para que
`vssadmin` pueda listar y eliminar las copias Shadow.

## Ejemplo de salida

```text
=== Iniciando eliminación de copias shadow ===
Copias shadow encontradas:
...
✓ Copias shadow eliminadas exitosamente

=== Iniciando búsqueda de archivos ===
Buscando en unidades: [C:]

Escaneando unidad: C:\
  Encontrado: C:\Users\Usuario\Documents\informe.pdf

✓ Búsqueda completada. Total de archivos encontrados: 1

=== RESULTADOS DE LA BÚSQUEDA ===
Total de archivos encontrados: 1

1. C:\Users\Usuario\Documents\informe.pdf
=== Iniciando cifrado de archivos ===

Clave de cifrado: ByClaude
Tamaño de clave: 32 bytes (AES-256)

[1/1] Cifrando: C:\Users\Usuario\Documents\informe.pdf
  ✓ Archivo cifrado exitosamente
```

## Limitaciones importantes

- La clave está escrita directamente en el código y se imprime por consola.
- No existe una función de descifrado.
- Si se pierde la clave, no hay un mecanismo alternativo de recuperación.
- La clave se amplía mediante repetición, no mediante una función de derivación
  de claves criptográficamente adecuada.
- Cada archivo se carga completamente en memoria.
- Los archivos originales se sobrescriben; no se crean copias de seguridad.
- El programa no comprueba que el contenido cifrado pueda recuperarse.
- Los errores de acceso durante el recorrido se ignoran.
- No se guarda un registro persistente de los archivos procesados.
- El escaneo puede tardar mucho al recorrer unidades completas.
- La eliminación de Shadow Copies se realiza antes del cifrado.

## Uso seguro en laboratorio

Para probarlo sin poner en riesgo información real:

1. Utiliza una máquina virtual aislada.
2. Crea únicamente archivos de prueba.
3. Desconecta el entorno de redes no necesarias.
4. No ejecutes el programa en equipos de producción.
5. Comprueba previamente el código y los permisos.
6. Conserva una copia externa de cualquier dato importante.

