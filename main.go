package main

// libraries
import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fernet/fernet-go"
)

// control object
//   - operation:
//     ENCODE or DECODE
//   - input file
//   - output file
type Control struct {
	Operation  string `json:"operation"`
	InputFile  string `json:"inputfile"`
	OutputFile string `json:"outputfile"`
}

// read control file and store data in control object
// obj now can be used to get actions and data
func loadConfig(path string, obj *Control) error {
	// (1) read json control file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read control file: %w", err)
	}
	// (2) json parsen
	if err := json.Unmarshal(data, obj); err != nil {
		return fmt.Errorf("parse config json %s: %w", path, err)
	}
	return nil
}

// zip and encrypt function
// - input file
// - output file
// - used encryption key
func encryptFile(inputFile, outputFile string, key *fernet.Key) error {
	// read input file
	plaintext, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}
	// zip in memory
	var zipBuffer bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuffer)
	entry, err := zipWriter.Create(filepath.Base(inputFile))
	if err != nil {
		return fmt.Errorf("create zip entry: %w", err)
	}
	// write zip file
	if _, err := entry.Write(plaintext); err != nil {
		return fmt.Errorf("write into zip: %w", err)
	}
	// close zip
	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("close zip: %w", err)
	}
	// encode with encryption key
	token, err := fernet.EncryptAndSign(zipBuffer.Bytes(), key)
	if err != nil {
		return fmt.Errorf("encode zip file: %w", err)
	}
	// write encrypted file
	if err := os.WriteFile(outputFile, token, 0600); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	// return value
	return nil
}

// decrypt and unzip file
// - input file (encrypted file)
// - output file (orig file)
// - key (that was used also to encrypt)
func decryptFile(inputFile, outputFile string, key *fernet.Key) error {
	// read encrypted file
	token, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("read encrypted file: %w", err)
	}
	// decode file
	zipBytes := fernet.VerifyAndDecrypt(token, 0, []*fernet.Key{key})
	if zipBytes == nil {
		return errors.New("error in decryption")
	}
	// unzip
	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	// no file
	if len(zipReader.File) == 0 {
		return errors.New("no file in zip")
	}
	// unzip first file in zip
	zipEntry := zipReader.File[0]
	src, err := zipEntry.Open()
	if err != nil {
		return fmt.Errorf("open zip entry: %w", err)
	}
	defer src.Close()
	// write output file
	dst, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer dst.Close()
	// unpack file
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("unpack file: %w", err)
	}
	// return value
	return nil
}

// read the Fernet key from the first line of key.txt
// this key will be used for encode or decode
func loadKey(path string) (*fernet.Key, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read key file %s: %w", path, err)
	}
	// use only the first line
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return nil, errors.New("key file is empty")
	}
	keyString := strings.TrimSpace(lines[0])
	if keyString == "" {
		return nil, errors.New("first line of key file is empty")
	}
	keys, err := fernet.DecodeKeys(keyString)
	if err != nil {
		return nil, fmt.Errorf("decode Fernet key: %w", err)
	}
	return keys[0], nil
}

// main function
//   - common settings
//   - read control from json
//   - get key
//   - depending on control request:
//     encode
//     decode
func main() {
	// common settings
	osName := runtime.GOOS
	current_dir, _ := os.Getwd()
	fmt.Println(osName, current_dir)
	// control object
	var ctrl Control
	if err := loadConfig("control.json", &ctrl); err != nil {
		fmt.Fprintln(os.Stderr, "Config Error:", err)
		os.Exit(1)
	}
	fmt.Printf("Operation: %s\n", ctrl.Operation)
	fmt.Printf("Input:     %s\n", ctrl.InputFile)
	fmt.Printf("Output:    %s\n", ctrl.OutputFile)
	// key for encryption
	key, err := loadKey("c:/tmp/key.txt")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Key Error:", err)
		os.Exit(1)
	}

	// (1)
	// ENCODE
	switch strings.ToUpper(strings.TrimSpace(ctrl.Operation)) {
	case "ENCODE":
		if err := encryptFile(ctrl.InputFile, ctrl.OutputFile, key); err != nil {
			fmt.Fprintln(os.Stderr, "Encryption Error:", err)
			os.Exit(1)
		}
		fmt.Printf("file encrypted: %s -> %s\n", ctrl.InputFile, ctrl.OutputFile)
		// (2)
	// DECODE
	case "DECODE":
		if err := decryptFile(ctrl.InputFile, ctrl.OutputFile, key); err != nil {
			fmt.Fprintln(os.Stderr, "Decryption Error:", err)
			os.Exit(1)
		}
		fmt.Printf("file decrypted: %s -> %s\n", ctrl.InputFile, ctrl.OutputFile)
		// (3)
	// unknown operation - no encode or decode found
	default:
		fmt.Fprintf(os.Stderr, "unknown operation: %q (valid: ENCODE, DECODE)\n", ctrl.Operation)
		os.Exit(1)
	}

}
