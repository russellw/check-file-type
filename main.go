package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type fileType struct {
	name       string
	extensions []string // lowercase, without dot
}

type magicDetector struct {
	offset int
	magic  []byte
	ft     fileType
}

var magicDetectors = []magicDetector{
	{0, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, fileType{"PNG image", []string{"png"}}},
	{0, []byte{0xff, 0xd8, 0xff}, fileType{"JPEG image", []string{"jpg", "jpeg"}}},
	{0, []byte{'G', 'I', 'F', '8'}, fileType{"GIF image", []string{"gif"}}},
	{0, []byte{0x42, 0x4d}, fileType{"BMP image", []string{"bmp", "dib"}}},
	{0, []byte{'I', 'I', 0x2a, 0x00}, fileType{"TIFF image", []string{"tif", "tiff"}}},
	{0, []byte{'M', 'M', 0x00, 0x2a}, fileType{"TIFF image", []string{"tif", "tiff"}}},
	{0, []byte{0x00, 0x00, 0x01, 0x00}, fileType{"ICO image", []string{"ico"}}},
	{0, []byte{0x00, 0x00, 0x02, 0x00}, fileType{"CUR image", []string{"cur"}}},
	{0, []byte{'8', 'B', 'P', 'S'}, fileType{"Photoshop document", []string{"psd", "psb"}}},
	{0, []byte{'%', 'P', 'D', 'F'}, fileType{"PDF document", []string{"pdf"}}},
	{0, []byte{0x1f, 0x8b}, fileType{"GZIP archive", []string{"gz", "gzip", "tgz", "svgz"}}},
	{0, []byte{'B', 'Z', 'h'}, fileType{"BZIP2 archive", []string{"bz2", "tbz2"}}},
	{0, []byte{0xfd, '7', 'z', 'X', 'Z', 0x00}, fileType{"XZ archive", []string{"xz", "txz"}}},
	{0, []byte{'7', 'z', 0xbc, 0xaf, 0x27, 0x1c}, fileType{"7-Zip archive", []string{"7z"}}},
	{0, []byte{'R', 'a', 'r', '!', 0x1a, 0x07}, fileType{"RAR archive", []string{"rar"}}},
	{0, []byte{0x28, 0xb5, 0x2f, 0xfd}, fileType{"Zstandard archive", []string{"zst"}}},
	{0, []byte{0x04, 0x22, 0x4d, 0x18}, fileType{"LZ4 archive", []string{"lz4"}}},
	{0, []byte{0x7f, 'E', 'L', 'F'}, fileType{"ELF binary", []string{"elf", "so", "out"}}},
	{0, []byte{'M', 'Z'}, fileType{"PE binary", []string{"exe", "dll", "sys", "scr", "com"}}},
	{0, []byte{0xfe, 0xed, 0xfa, 0xce}, fileType{"Mach-O binary", []string{"dylib", "o"}}},
	{0, []byte{0xce, 0xfa, 0xed, 0xfe}, fileType{"Mach-O binary", []string{"dylib", "o"}}},
	{0, []byte{0xfe, 0xed, 0xfa, 0xcf}, fileType{"Mach-O binary", []string{"dylib", "o"}}},
	{0, []byte{0xcf, 0xfa, 0xed, 0xfe}, fileType{"Mach-O binary", []string{"dylib", "o"}}},
	{0, []byte{0xca, 0xfe, 0xba, 0xbe}, fileType{"Java class or Mach-O fat binary", []string{"class", "dylib"}}},
	{0, []byte{0x00, 'a', 's', 'm'}, fileType{"WebAssembly binary", []string{"wasm"}}},
	{0, []byte{'f', 'L', 'a', 'C'}, fileType{"FLAC audio", []string{"flac"}}},
	{0, []byte{'O', 'g', 'g', 'S'}, fileType{"OGG audio/video", []string{"ogg", "oga", "ogv", "opus"}}},
	{0, []byte{'I', 'D', '3'}, fileType{"MP3 audio", []string{"mp3"}}},
	{0, []byte{0xff, 0xfb}, fileType{"MP3 audio", []string{"mp3"}}},
	{0, []byte{0xff, 0xf3}, fileType{"MP3 audio", []string{"mp3"}}},
	{0, []byte{0xff, 0xf2}, fileType{"MP3 audio", []string{"mp3"}}},
	{0, []byte{0x1a, 0x45, 0xdf, 0xa3}, fileType{"Matroska/WebM", []string{"mkv", "webm", "mka", "mks"}}},
	{0, []byte{0x25, 0x21}, fileType{"PostScript", []string{"ps", "eps", "ai"}}},
	{0, []byte{'S', 'Q', 'L', 'i', 't', 'e', ' ', 'f', 'o', 'r', 'm', 'a', 't', ' ', '3', 0x00}, fileType{"SQLite database", []string{"db", "sqlite", "sqlite3", "db3"}}},
	{0, []byte{0xed, 0xab, 0xee, 0xdb}, fileType{"RPM package", []string{"rpm"}}},
	{0, []byte{'!', '<', 'a', 'r', 'c', 'h', '>'}, fileType{"AR archive", []string{"a", "deb"}}},
	// RIFF, ZIP, MP4, TAR handled by dedicated functions below
}

func matchesMagic(data []byte, offset int, magic []byte) bool {
	if offset+len(magic) > len(data) {
		return false
	}
	for i, b := range magic {
		if data[offset+i] != b {
			return false
		}
	}
	return true
}

func detectRIFF(data []byte) *fileType {
	if len(data) < 12 || string(data[0:4]) != "RIFF" {
		return nil
	}
	switch string(data[8:12]) {
	case "WEBP":
		return &fileType{"WebP image", []string{"webp"}}
	case "WAVE":
		return &fileType{"WAV audio", []string{"wav", "wave"}}
	case "AVI ":
		return &fileType{"AVI video", []string{"avi"}}
	}
	return nil
}

func detectAIFF(data []byte) bool {
	return len(data) >= 12 && string(data[0:4]) == "FORM" &&
		(string(data[8:12]) == "AIFF" || string(data[8:12]) == "AIFC")
}

// detectZIPBased scans local file headers to identify the specific ZIP-based format.
func detectZIPBased(data []byte) *fileType {
	pos := 0
	for i := 0; i < 30 && pos+30 <= len(data); i++ {
		if data[pos] != 'P' || data[pos+1] != 'K' || data[pos+2] != 3 || data[pos+3] != 4 {
			break
		}
		comprMethod := int(data[pos+8]) | int(data[pos+9])<<8
		compressedSize := int(data[pos+18]) | int(data[pos+19])<<8 | int(data[pos+20])<<16 | int(data[pos+21])<<24
		filenameLen := int(data[pos+26]) | int(data[pos+27])<<8
		extraLen := int(data[pos+28]) | int(data[pos+29])<<8

		filenameEnd := pos + 30 + filenameLen
		if filenameEnd > len(data) {
			break
		}
		filename := string(data[pos+30 : filenameEnd])

		// ODF and EPUB store an uncompressed "mimetype" file as the first entry
		if filename == "mimetype" && comprMethod == 0 && compressedSize > 0 {
			mimeStart := filenameEnd + extraLen
			mimeEnd := mimeStart + compressedSize
			if mimeEnd <= len(data) {
				mime := string(data[mimeStart:mimeEnd])
				switch {
				case mime == "application/epub+zip":
					return &fileType{"EPUB ebook", []string{"epub"}}
				case strings.HasPrefix(mime, "application/vnd.oasis.opendocument.text"):
					return &fileType{"ODF text document", []string{"odt", "fodt"}}
				case strings.HasPrefix(mime, "application/vnd.oasis.opendocument.spreadsheet"):
					return &fileType{"ODF spreadsheet", []string{"ods", "fods"}}
				case strings.HasPrefix(mime, "application/vnd.oasis.opendocument.presentation"):
					return &fileType{"ODF presentation", []string{"odp", "fodp"}}
				case strings.HasPrefix(mime, "application/vnd.oasis.opendocument.graphics"):
					return &fileType{"ODF drawing", []string{"odg", "fodg"}}
				case strings.HasPrefix(mime, "application/vnd.oasis.opendocument"):
					return &fileType{"ODF document", []string{"odt", "ods", "odp", "odg"}}
				}
			}
		}

		switch {
		case strings.HasPrefix(filename, "word/"):
			return &fileType{"Word document", []string{"docx", "docm"}}
		case strings.HasPrefix(filename, "xl/"):
			return &fileType{"Excel spreadsheet", []string{"xlsx", "xlsm", "xlsb"}}
		case strings.HasPrefix(filename, "ppt/"):
			return &fileType{"PowerPoint presentation", []string{"pptx", "pptm"}}
		case filename == "AndroidManifest.xml":
			return &fileType{"Android package", []string{"apk", "xapk"}}
		case filename == "META-INF/MANIFEST.MF":
			return &fileType{"Java archive", []string{"jar", "war", "ear"}}
		}

		next := filenameEnd + extraLen + compressedSize
		if next <= pos {
			break
		}
		pos = next
	}
	return &fileType{"ZIP archive", []string{"zip", "cbz", "ipa", "xpi", "crx"}}
}

var heicBrands = map[string]bool{
	"heic": true, "heis": true, "hevc": true, "hevx": true,
	"heim": true, "hevm": true, "hevs": true, "mif1": true, "msf1": true,
}

var avifBrands = map[string]bool{
	"avif": true, "avis": true,
}

func detectISOBMFF(data []byte) *fileType {
	if len(data) < 12 || string(data[4:8]) != "ftyp" {
		return nil
	}
	brand := string(data[8:12])
	if heicBrands[brand] {
		return &fileType{"HEIC/HEIF image", []string{"heic", "heif"}}
	}
	if avifBrands[brand] {
		return &fileType{"AVIF image", []string{"avif"}}
	}
	if brand == "qt  " {
		return &fileType{"QuickTime video", []string{"mov", "qt"}}
	}
	return &fileType{"MP4/M4x video", []string{"mp4", "m4v", "m4a", "m4b", "m4p"}}
}

func detectTAR(data []byte) bool {
	return len(data) >= 262 && string(data[257:262]) == "ustar"
}

func detect(data []byte) *fileType {
	// ZIP-based formats need special handling before the generic ZIP match
	if matchesMagic(data, 0, []byte{'P', 'K', 3, 4}) {
		return detectZIPBased(data)
	}
	for _, d := range magicDetectors {
		if matchesMagic(data, d.offset, d.magic) {
			ft := d.ft
			return &ft
		}
	}
	if matchesMagic(data, 0, []byte{'R', 'I', 'F', 'F'}) {
		if ft := detectRIFF(data); ft != nil {
			return ft
		}
	}
	if detectAIFF(data) {
		return &fileType{"AIFF audio", []string{"aiff", "aif", "aifc"}}
	}
	if ft := detectISOBMFF(data); ft != nil {
		return ft
	}
	if detectTAR(data) {
		return &fileType{"TAR archive", []string{"tar"}}
	}
	return nil
}

// checkFile returns the detected fileType and current extension if they mismatch, nil otherwise.
func checkFile(path string) (ext string, ft *fileType) {
	ext = strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	if ext == "" {
		return "", nil
	}

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	buf := make([]byte, 4096)
	n, _ := f.Read(buf)
	if n == 0 {
		return "", nil
	}
	buf = buf[:n]

	ft = detect(buf)
	if ft == nil {
		return "", nil
	}
	for _, e := range ft.extensions {
		if e == ext {
			return "", nil
		}
	}
	return ext, ft
}

func handleMismatch(path, ext string, ft *fileType, rename bool) (counted bool) {
	if !rename {
		fmt.Printf("%s: extension .%s does not match content (%s)\n", path, ext, ft.name)
		return true
	}
	newPath := strings.TrimSuffix(path, filepath.Ext(path)) + "." + ft.extensions[0]
	if _, err := os.Stat(newPath); err == nil {
		fmt.Printf("%s: would rename to %s but destination already exists\n", path, newPath)
		return true
	}
	if err := os.Rename(path, newPath); err != nil {
		panic(err)
	}
	fmt.Printf("renamed %s -> %s\n", path, newPath)
	return true
}

func walkPath(path string, rename bool) (mismatches int) {
	info, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	if !info.IsDir() {
		ext, ft := checkFile(path)
		if ft != nil {
			handleMismatch(path, ext, ft, rename)
			return 1
		}
		return 0
	}
	err = filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if fi.IsDir() {
			return nil
		}
		ext, ft := checkFile(p)
		if ft != nil {
			handleMismatch(p, ext, ft, rename)
			mismatches++
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return mismatches
}

func main() {
	rename := flag.Bool("rename", false, "rename files to the correct extension")
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		args = []string{"."}
	}
	mismatches := 0
	for _, arg := range args {
		mismatches += walkPath(arg, *rename)
	}
	if mismatches > 0 {
		os.Exit(1)
	}
}
