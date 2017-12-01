package disk

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const defaultDataDir = "/var/lib/cni/networks"

// Store is a simple disk-backed store that creates one file per network namespace
// container ID is a given filename. The contents of the file are the interface name for ovs.
type Store struct {
	dataDir string
}

func New(network, dataDir string) (*Store, error) {
	if dataDir == "" {
		dataDir = defaultDataDir
	}
	dir := filepath.Join(dataDir, network)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	return &Store{dir}, nil
}

func (s *Store) Reserve(id, ovsIfaceName string) (bool, error) {
	fname := strings.TrimSpace(id)
	fpath := filepath.Join(s.dataDir, fname)

	f, err := os.OpenFile(fpath, os.O_RDWR|os.O_EXCL|os.O_CREATE, 0644)
	if os.IsExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if _, err := f.WriteString(ovsIfaceName); err != nil {
		f.Close()
		os.Remove(f.Name())
		return false, err
	}
	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		return false, err
	}
	return true, nil
}

func (s *Store) ReleaseByID(id string) (string, error) {
	fname := strings.TrimSpace(id)
	fpath := filepath.Join(s.dataDir, fname)

	data, err := ioutil.ReadFile(fpath)
	if err != nil {
		return "", err
	}

	// log.Infof("delete port from ovs %s", ovsIface)
	// err = br.delPort(ovsIface)
	// if err != nil {
	// 	log.Fatalf("failed to delPort from switch %v", err)
	// 	return "", err
	// }

	if err := os.Remove(fpath); err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
