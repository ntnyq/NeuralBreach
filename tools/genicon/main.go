// genicon renders the shared SVG into desktop and documentation icons.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	sourcePath      = "resources/icon-macos.svg"
	manifestPath    = "tools/genicon/assets.json"
	rendererVersion = "0.48.1"
)

var inputPaths = []string{sourcePath, "tools/genicon/main.go"}
var outputPaths = []string{"resources/icon.png", "docs/public/icon.svg", "docs/public/icon.png", "docs/public/favicon.ico"}
var faviconSizes = []int{16, 32, 48, 64}

type manifest struct {
	Renderer string            `json:"renderer"`
	Hashes   map[string]string `json:"sha256"`
}

func main() {
	root := flag.String("root", ".", "repository root")
	check := flag.Bool("check", false, "verify committed assets without invoking the renderer")
	renderer := flag.String("renderer", "resvg", "path to resvg 0.48.1 (generation only)")
	flag.Parse()
	var err error
	if *check {
		err = checkAssets(*root)
	} else {
		err = generate(*root, *renderer)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func generate(root, renderer string) error {
	version, err := exec.Command(renderer, "--version").Output()
	if err != nil {
		return fmt.Errorf("icon generation needs resvg %s: %w (see docs/design/neural-breach/icon-vector.md)", rendererVersion, err)
	}
	if strings.TrimSpace(string(version)) != rendererVersion {
		return fmt.Errorf("expected %s, got %q", rendererVersion, strings.TrimSpace(string(version)))
	}
	tmp, err := os.MkdirTemp("", "neural-breach-icons-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	svg, err := os.ReadFile(filepath.Join(root, sourcePath))
	if err != nil {
		return err
	}
	// Render a snapshot so every size and the copied SVG use the same source bytes.
	source := filepath.Join(tmp, "icon.svg")
	if err := os.WriteFile(source, svg, 0o644); err != nil {
		return err
	}
	render := func(size int) ([]byte, error) {
		dest := filepath.Join(tmp, strconv.Itoa(size)+".png")
		if output, err := exec.Command(renderer, "--skip-system-fonts", "--width", strconv.Itoa(size), source, dest).CombinedOutput(); err != nil {
			return nil, fmt.Errorf("render %dpx icon: %w: %s", size, err, output)
		}
		data, err := os.ReadFile(dest)
		if err != nil {
			return nil, err
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil || cfg.Width != size || cfg.Height != size {
			return nil, fmt.Errorf("renderer returned an invalid %dpx PNG", size)
		}
		return data, nil
	}
	app, err := render(1024)
	if err != nil {
		return err
	}
	frames := make([][]byte, len(faviconSizes))
	for i, size := range faviconSizes {
		if frames[i], err = render(size); err != nil {
			return err
		}
	}
	assets := [][]byte{app, svg, app, encodeICO(frames)}
	for i, name := range outputPaths {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, assets[i], 0o644); err != nil {
			return err
		}
	}
	// Publish the manifest last: partial generation never passes -check.
	m, err := assetManifest(root)
	if err != nil {
		return err
	}
	m.Hashes[sourcePath] = fmt.Sprintf("%x", sha256.Sum256(svg))
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, manifestPath), append(data, '\n'), 0o644)
}

// ICO directory entries point to individually rendered, alpha-preserving PNGs.
func encodeICO(frames [][]byte) []byte {
	header := make([]byte, 6+16*len(frames))
	binary.LittleEndian.PutUint16(header[2:], 1)
	binary.LittleEndian.PutUint16(header[4:], uint16(len(frames)))
	offset := len(header)
	for i, frame := range frames {
		entry := header[6+16*i:]
		entry[0], entry[1] = byte(faviconSizes[i]), byte(faviconSizes[i])
		binary.LittleEndian.PutUint16(entry[4:], 1)
		binary.LittleEndian.PutUint16(entry[6:], 32)
		binary.LittleEndian.PutUint32(entry[8:], uint32(len(frame)))
		binary.LittleEndian.PutUint32(entry[12:], uint32(offset))
		offset += len(frame)
	}
	for _, frame := range frames {
		header = append(header, frame...)
	}
	return header
}

func assetManifest(root string) (manifest, error) {
	m := manifest{Renderer: rendererVersion, Hashes: make(map[string]string)}
	for _, paths := range [][]string{inputPaths, outputPaths} {
		for _, name := range paths {
			data, err := os.ReadFile(filepath.Join(root, name))
			if err != nil {
				return m, err
			}
			m.Hashes[name] = fmt.Sprintf("%x", sha256.Sum256(data))
		}
	}
	return m, nil
}

func checkAssets(root string) error {
	data, err := os.ReadFile(filepath.Join(root, manifestPath))
	if err != nil {
		return fmt.Errorf("read icon manifest (run make icon): %w", err)
	}
	var saved manifest
	if err := json.Unmarshal(data, &saved); err != nil {
		return err
	}
	current, err := assetManifest(root)
	if err != nil {
		return fmt.Errorf("read icon assets (run make icon): %w", err)
	}
	if saved.Renderer != current.Renderer || len(saved.Hashes) != len(current.Hashes) {
		return fmt.Errorf("icon manifest is outdated; run make icon")
	}
	for name, hash := range current.Hashes {
		if saved.Hashes[name] != hash {
			return fmt.Errorf("%s has changed; run make icon and commit all generated assets", name)
		}
	}
	return nil
}
