package main

import (
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
	{0, []byte{'%', 'P', 'D', 'F'}, fileType{"PDF document", []string{"pdf"}}},
	// ZIP covers many container formats
	{0, []byte{'P', 'K', 0x03, 0x04}, fileType{"ZIP archive", []string{
		"zip", "jar", "war", "ear", "apk", "ipa", "xpi", "crx",
		"docx", "xlsx", "pptx", "odt", "ods", "odp", "odg",
		"epub", "cbz",
	}}},
	{0, []byte{0x1f, 0x8b}, fileType{"GZIP archive", []string{"gz", "gzip", "tgz", "svgz"}}},
	{0, []byte{'B', 'Z', 'h'}, fileType{"BZIP2 archive", []string{"bz2", "tbz2"}}},
	{0, []byte{0xfd, '7', 'z', 'X', 'Z', 0x00}, fileType{"XZ archive", []string{"xz", "txz"}}},
	{0, []byte{'7', 'z', 0xbc, 0xaf, 0x27, 0x1c}, fileType{"7-Zip archive", []string{"7z"}}},
	{0, []byte{'R', 'a', 'r', '!', 0x1a, 0x07}, fileType{"RAR archive", []string{"rar"}}},
	{0, []byte{0x7f, 'E', 'L', 'F'}, fileType{"ELF binary", []string{"elf", "so", "out"}}},
	{0, []byte{'M', 'Z'}, fileType{"PE binary", []string{"exe", "dll", "sys", "scr", "com"}}},
	{0, []byte{'f', 'L', 'a', 'C'}, fileType{"FLAC audio", []string{"flac"}}},
	{0, []byte{'O', 'g', 'g', 'S'}, fileType{"OGG audio/video", []string{"ogg", "oga", "ogv", "opus"}}},
	{0, []byte{'I', 'D', '3'}, fileType{"MP3 audio", []string{"mp3"}}},
	{0, []byte{0x1a, 0x45, 0xdf, 0xa3}, fileType{"Matroska/WebM", []string{"mkv", "webm", "mka", "mks"}}},
	{0, []byte{0x42, 0x4d}, fileType{"BMP image", []string{"bmp", "dib"}}},
	{0, []byte{'I', 'I', 0x2a, 0x00}, fileType{"TIFF image", []string{"tif", "tiff"}}},
	{0, []byte{'M', 'M', 0x00, 0x2a}, fileType{"TIFF image", []string{"tif", "tiff"}}},
	{0, []byte{0xca, 0xfe, 0xba, 0xbe}, fileType{"Java class", []string{"class"}}},
	{0, []byte{0x25, 0x21}, fileType{"PostScript", []string{"ps", "eps", "ai"}}},
	// RIFF container handled below by detectRIFF
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

func detectMP4(data []byte) bool {
	// ISO base media: ftyp box at offset 4
	return len(data) >= 8 && string(data[4:8]) == "ftyp"
}

func detectTAR(data []byte) bool {
	return len(data) >= 262 && string(data[257:262]) == "ustar"
}

func detect(data []byte) *fileType {
	for _, d := range magicDetectors {
		if d.ft.name == "" {
			continue // skip RIFF placeholder
		}
		if matchesMagic(data, d.offset, d.magic) {
			ft := d.ft
			return &ft
		}
	}
	// RIFF container: check after magic byte match
	if matchesMagic(data, 0, []byte{'R', 'I', 'F', 'F'}) {
		if ft := detectRIFF(data); ft != nil {
			return ft
		}
	}
	if detectMP4(data) {
		ft := fileType{"MP4/M4x video", []string{"mp4", "m4v", "m4a", "m4b", "m4p", "mov"}}
		return &ft
	}
	if detectTAR(data) {
		ft := fileType{"TAR archive", []string{"tar"}}
		return &ft
	}
	return nil
}

func checkFile(path string) (mismatch bool, msg string) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	if ext == "" {
		return false, ""
	}

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	if n == 0 {
		return false, ""
	}
	buf = buf[:n]

	ft := detect(buf)
	if ft == nil {
		return false, ""
	}
	for _, e := range ft.extensions {
		if e == ext {
			return false, ""
		}
	}
	return true, fmt.Sprintf("%s: extension .%s does not match content (%s)", path, ext, ft.name)
}

func walkPath(path string) (mismatches int) {
	info, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	if !info.IsDir() {
		mismatch, msg := checkFile(path)
		if mismatch {
			fmt.Println(msg)
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
		mismatch, msg := checkFile(p)
		if mismatch {
			fmt.Println(msg)
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
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"."}
	}
	mismatches := 0
	for _, arg := range args {
		mismatches += walkPath(arg)
	}
	if mismatches > 0 {
		os.Exit(1)
	}
}
