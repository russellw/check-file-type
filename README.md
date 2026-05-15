# check-file-type
Verify consistency between file extension and file contents.

## Usage

```
check-file-type [-rename] [file|dir ...]
```

Checks each file by reading its magic bytes and comparing against the extension. Directories are walked recursively. If no arguments are given, the current directory is scanned.

Prints one line per mismatch:

```
path/to/file.jpg: extension .jpg does not match content (PNG image)
```

Exits with code 1 if any mismatches are found, 0 otherwise.

### Options

`-rename` — rename each mismatched file to use the correct extension:

```
renamed path/to/file.jpg -> path/to/file.png
```

If the destination filename already exists, the file is left in place and a warning is printed.

## Supported formats

| Format | Extensions |
|--------|-----------|
| PNG | `.png` |
| JPEG | `.jpg` `.jpeg` |
| GIF | `.gif` |
| WebP | `.webp` |
| BMP | `.bmp` `.dib` |
| TIFF | `.tif` `.tiff` |
| ICO | `.ico` |
| CUR | `.cur` |
| HEIC/HEIF | `.heic` `.heif` |
| AVIF | `.avif` |
| Photoshop | `.psd` `.psb` |
| PDF | `.pdf` |
| Word document | `.docx` `.docm` |
| Excel spreadsheet | `.xlsx` `.xlsm` `.xlsb` |
| PowerPoint presentation | `.pptx` `.pptm` |
| ODF text | `.odt` |
| ODF spreadsheet | `.ods` |
| ODF presentation | `.odp` |
| ODF drawing | `.odg` |
| EPUB | `.epub` |
| ZIP | `.zip` `.cbz` `.ipa` `.xpi` `.crx` |
| GZIP | `.gz` `.tgz` `.gzip` `.svgz` |
| BZIP2 | `.bz2` `.tbz2` |
| XZ | `.xz` `.txz` |
| 7-Zip | `.7z` |
| RAR | `.rar` |
| TAR | `.tar` |
| Zstandard | `.zst` |
| LZ4 | `.lz4` |
| MP3 | `.mp3` |
| FLAC | `.flac` |
| OGG | `.ogg` `.oga` `.ogv` `.opus` |
| WAV | `.wav` `.wave` |
| AIFF | `.aiff` `.aif` `.aifc` |
| AVI | `.avi` |
| MP4/M4x | `.mp4` `.m4v` `.m4a` `.m4b` `.m4p` |
| QuickTime | `.mov` `.qt` |
| Matroska/WebM | `.mkv` `.webm` `.mka` |
| ELF binary | `.elf` `.so` `.out` |
| PE binary | `.exe` `.dll` `.sys` `.scr` |
| Mach-O binary | `.dylib` `.o` |
| WebAssembly | `.wasm` |
| Java class | `.class` |
| Java archive | `.jar` `.war` `.ear` |
| Android package | `.apk` |
| PostScript | `.ps` `.eps` `.ai` |
| SQLite database | `.db` `.sqlite` `.sqlite3` `.db3` |
| RPM package | `.rpm` |
| AR/DEB archive | `.a` `.deb` |

## Build

```
go build
```

Produces a statically linked binary with no external dependencies.
