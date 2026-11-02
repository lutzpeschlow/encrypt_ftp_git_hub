package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"

	"fernet"

	"github.com/fernet/fernet-go"
)

type Control struct {
	Operation  string `json:"operation"`
	InputFile  string `json:"inputfile"`
	OutputFile string `json:"outputfile"`
}

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

func main() {

	osName := runtime.GOOS
	current_dir, _ := os.Getwd()
	fmt.Println(osName, current_dir)

	var ctrl Control
	if err := loadConfig("control.json", &ctrl); err != nil {
		fmt.Fprintln(os.Stderr, "Config Error:", err)
		os.Exit(1)
	}

	// variables in ctrl
	fmt.Printf("Operation: %s\n", ctrl.Operation)
	fmt.Printf("Input:     %s\n", ctrl.InputFile)
	fmt.Printf("Output:    %s\n", ctrl.OutputFile)

	key := fernet.Key.Generate()
	print(key)
	//
	// // Datei lesen
	// plaintext, err := os.ReadFile(inputFile)
	// if err != nil {
	// 	fmt.Printf("Datei %q konnte nicht gelesen werden: %v\n", inputFile, err)
	// 	os.Exit(1)
	// }
	//
	// // Datei verschlüsseln
	// token, err := fernet.EncryptAndSign(plaintext, selectedKey)
	// if err != nil {
	// 	fmt.Println("Fehler beim Verschlüsseln:", err)
	// 	os.Exit(1)
	// }
	//
	// // Verschlüsselte Datei schreiben
	// if err := os.WriteFile(outputFile, token, 0600); err != nil {
	// 	fmt.Println("Fehler beim Schreiben der verschlüsselten Datei:", err)
	// 	os.Exit(1)
	// }
	//
	// fmt.Printf("\nDatei erfolgreich verschlüsselt: %s -> %s\n", inputFile, outputFile)
	// fmt.Println("Verwendeter Schlüssel:")
	// fmt.Println(keyTexts[1-1])
	//
	// fmt.Println("elapsed:", time.Since(start))

	s1 := strings.ToUpper("hallo")
	fmt.Println(s1)
}

// key, err := fernet.DecodeKey("HIER_DEIN_GESPEICHERTER_SCHLUESSEL")
// if err != nil {
// 	panic(err)
// }
//
// token, err := os.ReadFile("name.txt.fernet")
// if err != nil {
// 	panic(err)
// }
//
// plaintext := fernet.VerifyAndDecrypt(token, 0, []*fernet.Key{key})
// if plaintext == nil {
// 	panic("Entschlüsselung fehlgeschlagen")
// }
//
// err = os.WriteFile("name.txt.decrypted", plaintext, 0600)
// if err != nil {
// 	panic(err)
// }
