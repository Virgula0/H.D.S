package daemon

import (
	"github.com/Virgula0/progetto-dp/raspberrypi/internal/constants"
	"github.com/Virgula0/progetto-dp/raspberrypi/internal/utils"
	"github.com/google/gopacket/pcap"
	"os"
	"path/filepath"
)

type Environment interface {
	LoadEnvironment() (map[string]*pcap.Handle, error)
	Close()
}

type Env struct {
	HandshakeDirectory string
	files              []*pcap.Handle
}

type ProdEnvironment struct {
	HandshakeDirectory string
	files              []*pcap.Handle
}

/*
ChooseEnvironment

Handy function for returning a different environment based on the presence of Bettercap environment variables
Because if it's a test, we will send test.pcap within hs directory, otherwise we get real handshakes from ~/handshakes
*/
func ChooseEnvironment() (Environment, error) {
	userDir, _ := os.UserHomeDir()
	pwd, _ := os.Getwd()

	switch constants.Bettercap {
	case true:
		dd := filepath.Join(userDir, "handshakes")
		if exists, _ := utils.DirExists(dd); !exists {
			err := os.MkdirAll(dd, os.ModePerm)
			if err != nil {
				return nil, err
			}
		}
		return &ProdEnvironment{
			HandshakeDirectory: filepath.Join(userDir, "hs"),
		}, nil
	default:
		return &Env{
			HandshakeDirectory: filepath.Join(pwd, "handshakes"),
		}, nil
	}
}

/*
getPaths

reads pcap using gopacket/pcap
*/
func getPaths(root, extension string) (map[string]*pcap.Handle, error) {
	files, err := utils.ReadFileNamesByExtension(root, extension)

	if err != nil {
		return nil, err
	}

	var pcaps = make(map[string]*pcap.Handle)
	for _, file := range files {
		handle, err := pcap.OpenOffline(file)
		if err != nil {
			return nil, err
		}
		pcaps[file] = handle
	}

	return pcaps, nil
}

func closeFiles(d []*pcap.Handle) {
	for _, handle := range d {
		handle.Close()
	}
}

func (d *Env) LoadEnvironment() (map[string]*pcap.Handle, error) {
	return getPaths(d.HandshakeDirectory, constants.PCAPExtension)
}

func (d *ProdEnvironment) LoadEnvironment() (map[string]*pcap.Handle, error) {
	return getPaths(d.HandshakeDirectory, constants.PCAPExtension)
}

func (d *Env) Close() {
	closeFiles(d.files)
}

func (d *ProdEnvironment) Close() {
	closeFiles(d.files)
}
