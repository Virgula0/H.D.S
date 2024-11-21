package main

import (
	"fmt"
	"unsafe"

	gocat "github.com/mandiant/gocat/v6"
	"github.com/mandiant/gocat/v6/hcargp"
)

const wordlistExample = "/usr/share/seclists/Passwords/darkc0de.txt"
const DebugTest = true

func callbackForTests(resultsmap map[string]*string) gocat.EventCallback {
	return func(hc unsafe.Pointer, payload interface{}) {
		switch pl := payload.(type) {
		case gocat.LogPayload:
			if DebugTest {
				fmt.Printf("LOG [%s] %s\n", pl.Level, pl.Message)
			}
		case gocat.ActionPayload:
			if DebugTest {
				fmt.Printf("ACTION [%d] %s\n", pl.HashcatEvent, pl.Message)
			}
		case gocat.CrackedPayload:
			if DebugTest {
				fmt.Printf("CRACKED %s -> %s\n", pl.Hash, pl.Value)
			}
			if resultsmap != nil {
				resultsmap[pl.Hash] = hcargp.GetStringPtr(pl.Value)
			}
		case gocat.FinalStatusPayload:
			if DebugTest {
				fmt.Printf("FINAL STATUS -> %v\n", pl.Status)
			}
		case gocat.TaskInformationPayload:
			if DebugTest {
				fmt.Printf("TASK INFO -> %v\n", pl)
			}
		}
	}
}

func main() {
	crackedHashes := map[string]*string{}

	tt := gocat.Options{
		ExecutablePath:    "/usr/local/bin",
		SharedPath:        "/tmp",
		PatchEventContext: true,
	}

	hashcat, err := gocat.New(tt, callbackForTests(crackedHashes))
	defer hashcat.Free()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 5d41402abc4b2a76b9719d911017c592:hello

	err = hashcat.RunJob("-O", "-a", "0", "-m", "0", "--potfile-disable", "--logfile-disable", "5d41402abc4b2a76b9719d911017c592", wordlistExample)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(crackedHashes)
}
