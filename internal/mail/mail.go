// Package mail provides a simple interface to send emails.
package mail

type Message struct {
	From    string
	To      []string
	Subject string
	Body    []byte
	IsHTML  bool
}

func NewMessage(from string, to []string, subject string, body []byte, isHTML bool) Message {
	return Message{
		From:    from,
		To:      to,
		Subject: subject,
		Body:    body,
		IsHTML:  isHTML,
	}
}

type Mailer interface {
	Send(msg Message) error
}

type Config struct {
	Server      string
	Port        int
	Username    string
	Password    string
	DefaultFrom string
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) WithServer(server string) *Config {
	c.Server = server
	return c
}

func (c *Config) WithPort(port int) *Config {
	c.Port = port
	return c
}

func (c *Config) WithUsername(username string) *Config {
	c.Username = username
	return c
}

func (c *Config) WithPassword(password string) *Config {
	c.Password = password
	return c
}

func (c *Config) WithDefaultFrom(defaultFrom string) *Config {
	c.DefaultFrom = defaultFrom
	return c
}
