package config

const (
	version  = "develop"
	asciiArt = `
                      __             __
    ____  ____ ______/ /_____  _____/ /__________  ____
   / __ \/ __ '/ ___/ //_/ _ \/ ___/ __/ ___/ __ \/ __ \
  / /_/ / /_/ / /  / ,< /  __/ /  / /_/ /  / /_/ / / / /
 / .___/\__,_/_/  /_/|_|\___/_/   \__/_/   \____/_/ /_/
/_/
`
)

type Parkertron struct {
	Services []string `json:"services,omitempty"`
	Log      `json:"log,omitempty"`
	Database `json:"database,omitempty"`
}

type Log struct {
	Level     string `json:"level,omitempty"`
	Directory string `json:"directory,omitempty"`
	Rotate    bool   `json:"rotate,omitempty"`
	Interval  int    `json:"interval,omitempty"`
}

type Database struct {
	Hostname string `json:"hoatname,omitempty"`
	Port     int    `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

func (new Parkertron) New() {
	new.Services = []string{"discord"}

	new.Log.Level = "info"
	new.Log.Directory = "logs/"
	new.Log.Rotate = true
	new.Log.Interval = 24 // time is in hours

	new.Database.Hostname = ""
	new.Database.Port = 5432
	new.Database.Username = "parkertron"
	new.Database.Password = "SuperSafePa$$w0rd"

	return
}
