package config

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"

	"github.com/0xPolygonHermez/zkevm-node/log"
	"github.com/JiamingSuper/polygon-zkevm-bridge/bridgectrl"
	"github.com/JiamingSuper/polygon-zkevm-bridge/claimtxman"
	"github.com/JiamingSuper/polygon-zkevm-bridge/db"
	"github.com/JiamingSuper/polygon-zkevm-bridge/etherman"
	"github.com/JiamingSuper/polygon-zkevm-bridge/server"
	"github.com/JiamingSuper/polygon-zkevm-bridge/synchronizer"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Config struct
type Config struct {
	Log              log.Config
	SyncDB           db.Config
	ClaimTxManager   claimtxman.Config
	Etherman         etherman.Config
	Synchronizer     synchronizer.Config
	BridgeController bridgectrl.Config
	BridgeServer     server.Config
	NetworkConfig
}

// Load loads the configuration
func Load(configFilePath string, network string) (*Config, error) {
	var cfg Config
	viper.SetConfigType("toml")

	err := viper.ReadConfig(bytes.NewBuffer([]byte(DefaultValues)))
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&cfg, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc()))
	if err != nil {
		return nil, err
	}
	if configFilePath != "" {
		dirName, fileName := filepath.Split(configFilePath)

		fileExtension := strings.TrimPrefix(filepath.Ext(fileName), ".")
		fileNameWithoutExtension := strings.TrimSuffix(fileName, "."+fileExtension)

		viper.AddConfigPath(dirName)
		viper.SetConfigName(fileNameWithoutExtension)
		viper.SetConfigType(fileExtension)
	}
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("ZKEVM_BRIDGE")
	err = viper.ReadInConfig()
	if err != nil {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if ok {
			log.Infof("config file not found")
		} else {
			log.Infof("error reading config file: ", err)
			return nil, err
		}
	}

	err = viper.Unmarshal(&cfg, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc()))
	if err != nil {
		return nil, err
	}

	if viper.IsSet("NetworkConfig") && network != "" {
		return nil, errors.New("Network details are provided in the config file (the [NetworkConfig] section) and as a flag (the --network or -n). Configure it only once and try again please.")
	}
	if !viper.IsSet("NetworkConfig") && network == "" {
		return nil, errors.New("Network details are not provided. Please configure the [NetworkConfig] section in your config file, or provide a --network flag.")
	}
	if !viper.IsSet("NetworkConfig") && network != "" {
		cfg.loadNetworkConfig(network)
	}

	return &cfg, nil
}
