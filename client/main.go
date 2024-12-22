package main

import (
	"fmt"
	"unsafe"

	"github.com/mandiant/gocat/v6"
	"github.com/mandiant/gocat/v6/hcargp"
)

const wordlistExample = "/usr/share/seclists/Passwords/darkc0de.txt"
const fileToCrack = "/home/angelo/tools/hcxtools/gotests/output.hashcat"
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
	/*
		rl.InitWindow(1200, 800, "raylib [core] example - basic window")
		defer rl.CloseWindow()

		rl.SetTargetFPS(60)

		for !rl.WindowShouldClose() {
			rl.BeginDrawing()

			rl.ClearBackground(rl.RayWhite)
			rl.DrawText("Congrats! You created your first window!", 190, 200, 20, rl.LightGray)

			rl.EndDrawing()
		}
	*/

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

	// potfile remembers cracked hashcat
	err = hashcat.RunJob("-a", "3", "-m", "22000", "--potfile-disable", "--logfile-disable", fileToCrack, "test12?d?d")

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(crackedHashes)

}
