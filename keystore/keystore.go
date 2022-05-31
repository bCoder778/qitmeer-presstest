package keystore

import (
	"io/ioutil"
	"strings"
	"sync"
)

type Keystore struct {
	keys  map[string]string
	dir   string
	mutex sync.RWMutex
}

func LoadKey(dir string) (*Keystore, error) {
	keyStore := &Keystore{keys: make(map[string]string, 0), dir: dir}
	fileInfoList, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, fileInfo := range fileInfoList {
		if !fileInfo.IsDir() && strings.Contains(fileInfo.Name(), ".key") {
			addressList := strings.Split(fileInfo.Name(), ".")
			if len(addressList) == 2 {
				bytes, err := ioutil.ReadFile(dir + "/" + fileInfo.Name())
				if err != nil {
					return nil, err
				}

				address := strings.Split(fileInfo.Name(), ".")[0]
				keyStore.keys[address] = string(bytes)
			}
		}
	}
	return keyStore, nil
}

func (k *Keystore) AddressList() []string {
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	addrs := []string{}
	for address, _ := range k.keys {
		addrs = append(addrs, address)
	}
	return addrs
}

func (k *Keystore) Key(address string) string {
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	return k.keys[address]
}

func (k *Keystore) RandAddress() string {
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	for address, _ := range k.keys {
		return address
	}
	return ""
}

func (k *Keystore) AddKey(address, key string) error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	err := ioutil.WriteFile(k.dir+"/"+address+".key", []byte(key), 0666)
	if err != nil {
		return err
	}
	k.keys[address] = key
	return nil
}
