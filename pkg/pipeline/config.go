package pipeline

import (
	"github.com/unchainio/pkg/xlogger"
)

type Config struct {
	Organization string
	ID           string
	Logger       *xlogger.Config
	Trigger      TriggerConfig
	Actions      ActionsConfig
}

type TriggerConfig struct {
	Config string
}

type ActionsConfig struct {
	ImapAction       *ImapActionConfig       `mapstructure:"imap_action"`
	FileparserAction *FileparserActionConfig `mapstructure:"fileparser_action"`
	TemplaterAction  *TemplaterActionConfig  `mapstructure:"templater_action"`
	HttpAction       *HttpActionConfig       `mapstructure:"http_action"`
	SmtpAction       *SmtpActionConfig       `mapstructure:"smtp_action"`
}

type ImapActionConfig struct {
	Config string
}

type FileparserActionConfig struct {
	Filetype string
	Header   bool
}

type TemplaterActionConfig struct {
	Template  string
	Variables map[string]interface{}
}

type HttpActionConfig struct {
	Url    string
	Method string
}

type SmtpActionConfig struct {
	Username   string
	Password   string
	Hostname   string
	Port       string
	From       string
	Recipients []string
}
