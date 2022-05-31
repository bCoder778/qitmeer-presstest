package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/bCoder778/log"
	"os"
	"sync"
)

var (
	ConfigFile = "config.toml"
)

var Setting *Config
var once sync.Once

func LoadConfig() error {
	if Exist(ConfigFile) {
		if _, err := toml.DecodeFile(ConfigFile, &Setting); err != nil {
			return err
		}
		return nil
	} else {
		return fmt.Errorf("%s not is exist", ConfigFile)
	}
}

type Config struct {
	Api   *Api   `toml:"api"`
	Rpc   []*Rpc `toml:"rpc"`
	DB    *DB    `toml:"db"`
	Log   *Log   `toml:"log"`
	Email *EMail `toml:"email"`
	Auth  *Auth  `toml:"auth"`
	Key   *Key   `toml:"key"`
	Cpu   *Cpu   `toml:"cpu"`
	Evm   *EVM   `toml:"evm"`
}

type Cpu struct {
	Number int `toml:"number"`
}

type Key struct {
	Dir string `dir`
}

type Api struct {
	Listen string `toml:"listen"`
}

type Rpc struct {
	Host     string `toml:"host"`
	Admin    string `toml:"admin"`
	Password string `toml:"password"`
	Net      string `toml:"net"`
}

type EVM struct {
	Rpc      []string `toml:"rpc"`
	Token    string   `toml:"token"`
	Mnemonic string   `toml:"mnemonic"`
	Address  string   `toml:"address"`
	Account  int      `toml:"account"`
}

type DB struct {
	DBType   string `toml:"dbtype"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Address  string `toml:"address"`
	DBName   string `toml:"dbname"`
}

type Auth struct {
	Jwt            bool   `toml:"jwt"`
	ExpirationTime int64  `toml:"expirationTime"`
	Issuer         string `toml:"issuer"`
	SecretKey      string `toml:"secretKey"`
}

type Log struct {
	Mode  log.Mode  `toml:"mode"`
	Level log.Level `toml:"level"`
	Path  string    `toml:"path"`
}

type EMail struct {
	Title string   `toml:"title"`
	User  string   `toml:"user"`
	Pass  string   `toml:"pass"`
	Host  string   `toml:"host"`
	Port  string   `toml:"port"`
	To    []string `toml:"to"`
}

func Exist(fileName string) bool {
	_, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
