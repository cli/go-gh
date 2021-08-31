package config

type Config interface {
	Get(key string) (string, error)
	GetForHost(host string, key string) (string, error)
}

type config struct {
}

//Read in config from file
func Read() (Config, error) {
	return nil, nil
}
