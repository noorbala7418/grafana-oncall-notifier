package models

type Config struct {
	LogLevel        string                   `yaml:"log_level"`
	EmailCredential EmailCredential          `yaml:"email_server"`
	Grafana         GrafanaOnCallCredentials `yaml:"grafana_oncall"`
	IgnoreShifts    []string                 `yaml:"ignore_oncalls"`
}

type Email struct {
	Sender   string
	Receiver string
	Subject  string
	Body     string
}

type EmailCredential struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type GrafanaOnCallCredentials struct {
	Url        string `yaml:"url"`
	AdminToken string `yaml:"admin_token"`
}

type Shift struct {
	ShiftID  string `json:"id"`
	Name     string `json:"name"`
	TeamID   string `json:"team_id"`
	TimeZone string `json:"time_zone"`
}

type OnCallShifts struct {
	Count  int     `json:"count"`
	Shifts []Shift `json:"results"`
}

type OnCalls struct {
	Count  int      `json:"count"`
	OnCall []OnCall `json:"results"`
}

type OnCall struct {
	Email      string `json:"user_email"`
	Username   string `json:"user_username"`
	ShiftStart string `json:"shift_start"`
	ShiftEnd   string `json:"shift_end"`
}
