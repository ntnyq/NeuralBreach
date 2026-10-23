package main

import (
	"bytes"
	"encoding/binary"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestCommittedIcons(t *testing.T) {
	root := filepath.Join("..", "..")
	if err := checkAssets(root); err != nil {
		t.Fatal(err)
	}
	app, err := os.ReadFile(filepath.Join(root, "resources/icon.png"))
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(app))
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != 1024 || img.Bounds().Dy() != 1024 {
		t.Fatalf("app icon dimensions: %v", img.Bounds())
	}
	if _, _, _, alpha := img.At(0, 0).RGBA(); alpha != 0 {
		t.Fatal("outside the base must be transparent")
	}
	for _, point := range [][2]int{{352, 304}, {352, 480}, {528, 480}, {704, 480}, {704, 656}} {
		c := color.NRGBAModel.Convert(img.At(point[0], point[1])).(color.NRGBA)
		if c.A != 255 || c.R > 100 || c.G > 100 || c.B > 100 {
			t.Errorf("recess at %v must expose the opaque dark base, got %v", point, c)
		}
	}
	web, err := os.ReadFile(filepath.Join(root, "docs/public/icon.png"))
	if err != nil || !bytes.Equal(app, web) {
		t.Fatal("web and desktop PNGs differ", err)
	}
	ico, err := os.ReadFile(filepath.Join(root, "docs/public/favicon.ico"))
	if err != nil {
		t.Fatal(err)
	}
	if len(ico) < 70 || binary.LittleEndian.Uint16(ico[0:]) != 0 || binary.LittleEndian.Uint16(ico[2:]) != 1 || binary.LittleEndian.Uint16(ico[4:]) != 4 {
		t.Fatal("invalid four-frame ICO header")
	}
	end := 70
	for i, size := range []int{16, 32, 48, 64} {
		entry := ico[6+i*16:]
		length := int(binary.LittleEndian.Uint32(entry[8:]))
		offset := int(binary.LittleEndian.Uint32(entry[12:]))
		if offset != end || length <= 0 || offset+length > len(ico) || int(entry[0]) != size || int(entry[1]) != size {
			t.Fatalf("invalid ICO directory entry %d", i)
		}
		frame, err := png.Decode(bytes.NewReader(ico[offset : offset+length]))
		if err != nil || frame.Bounds().Dx() != size || frame.Bounds().Dy() != size {
			t.Fatalf("invalid %dpx ICO payload: %v", size, err)
		}
		end = offset + length
	}
	if end != len(ico) {
		t.Fatal("unexpected ICO trailing bytes")
	}
}

func TestCheckRejectsChangedAssets(t *testing.T) {
	root := t.TempDir()
	paths := append([]string{manifestPath}, inputPaths...)
	paths = append(paths, outputPaths...)
	for _, name := range paths {
		data, err := os.ReadFile(filepath.Join("..", "..", name))
		if err != nil {
			t.Fatal(err)
		}
		dest := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range paths {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(root, name)
			original, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("stale"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := checkAssets(root); err == nil {
				t.Error("changed asset accepted")
			}
			if err := os.WriteFile(path, original, 0o644); err != nil {
				t.Fatal(err)
			}
		})
	}
}
