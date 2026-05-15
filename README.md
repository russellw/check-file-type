# check-file-type
Verify consistency between file extension and file contents.

## Usage

```
check-file-type [file|dir ...]
```

Checks each file by reading its magic bytes and comparing against the extension. Directories are walked recursively. If no arguments are given, the current directory is scanned.

Prints one line per mismatch:

```
path/to/file.jpg: extension .jpg does not match content (PNG image)
```

Exits with code 1 if any mismatches are found, 0 otherwise.

## Supported formats

| Format | Extensions |
|--------|-----------|
| PNG | `.png` |
| JPEG | `.jpg` `.jpeg` |
| GIF | `.gif` |
| WebP | `.webp` |
| BMP | `.bmp` `.dib` |
| TIFF | `.tif` `.tiff` |
| PDF | `.pdf` |
| ZIP (and ZIP-based) | `.zip` `.jar` `.war` `.apk` `.docx` `.xlsx` `.pptx` `.odt` `.epub` … |
| GZIP | `.gz` `.tgz` `.gzip` `.svgz` |
| BZIP2 | `.bz2` `.tbz2` |
| XZ | `.xz` `.txz` |
| 7-Zip | `.7z` |
| RAR | `.rar` |
| TAR | `.tar` |
| MP3 | `.mp3` |
| FLAC | `.flac` |
| OGG | `.ogg` `.oga` `.ogv` `.opus` |
| WAV | `.wav` `.wave` |
| AVI | `.avi` |
| MP4/MOV | `.mp4` `.m4v` `.m4a` `.mov` |
| Matroska/WebM | `.mkv` `.webm` `.mka` |
| ELF binary | `.elf` `.so` `.out` |
| PE binary | `.exe` `.dll` `.sys` `.scr` |
| Java class | `.class` |
| PostScript | `.ps` `.eps` `.ai` |

## Build

```
go build
```

Produces a statically linked binary with no external dependencies.
