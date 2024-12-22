package deamon

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

type TestEnvironment struct {
	HandshakeDirectory string
	files              []*pcap.Handle
}

type ProdEnvironment struct {
	HandshakeDirectory string
	files              []*pcap.Handle
}

func ChooseEnvironment() (Environment, error) {
	userDir, _ := os.UserHomeDir()
	pwd, _ := os.Getwd()

	switch constants.Test {
	case true:
		return &TestEnvironment{
			HandshakeDirectory: filepath.Join(pwd, "handshakes"),
		}, nil
	default:
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
	}
}

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

func (d *TestEnvironment) LoadEnvironment() (map[string]*pcap.Handle, error) {
	return getPaths(d.HandshakeDirectory, constants.PCAPExtension)
}

func (d *ProdEnvironment) LoadEnvironment() (map[string]*pcap.Handle, error) {
	return getPaths(d.HandshakeDirectory, constants.PCAPExtension)
}

func (d *TestEnvironment) Close() {
	closeFiles(d.files)
}

func (d *ProdEnvironment) Close() {
	closeFiles(d.files)
}
