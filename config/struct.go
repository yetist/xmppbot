package config

type Config struct {
	AppName    string
	AppVersion string
	AppConfig  string
	Account    struct {
		Username string
		Password string
		Resource string
		Server   string
		Port     int
		NoTLS    bool
		SelfSign bool `toml:"self_signed"`
		Session  bool
	}
	Setup struct {
		Admin         []string
		Debug         bool
		AutoSubscribe bool   `toml:"auto_subscribe"`
		CmdPrefix     string `toml:"cmd_prefix"`
		Status        string
		StatusMessage string `toml:"status_message"`
		WebHost       string `toml:"web_host"`
		WebPort       int    `toml:"web_port"`
		Rooms         []map[string]interface{}
	}
	Plugin map[string]map[string]interface{}
}
