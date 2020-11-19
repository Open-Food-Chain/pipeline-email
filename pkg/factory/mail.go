package factory

import (
	"fmt"
	imapclient "github.com/emersion/go-imap/client"
	"github.com/unchainio/interfaces/logger"
	"github.com/unchainio/pkg/errors"
	"gopkg.in/gomail.v2"
)

type ImapConfig struct {
	Port     string
	Domain   string
	Username string
	Password string
}

type ImapClient struct {
	logger logger.Logger
	cfg    *ImapConfig
	Client *client
}

type client struct {
	*imapclient.Client
}

func NewImapClient(logger logger.Logger, cfg *ImapConfig) (*ImapClient, error) {
	c, err := imapclient.DialTLS(fmt.Sprintf("%s%s", cfg.Domain, cfg.Port), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial imap server")
	}

	client := &ImapClient{
		logger: logger,
		cfg:    cfg,
		Client: &client{
			Client: c,
		},
	}

	err = c.Login(cfg.Username, cfg.Password)
	if err != nil {
		return nil, errors.Wrap(err, "failed to login")
	}

	go func() {
		loggedOut := <-client.Client.LoggedOut()
		logger.Warnf("logged out: %v", loggedOut)
	}()

	mbox, err := client.Client.Select("INBOX", false)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Started client - mailbox contains %v messages", mbox.Messages)

	return client, nil
}

func SendAustriaJuiceMail(username, password, host, from string, port int, tos []string, attachmentFilePath string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", tos...)
	m.SetHeader("Subject", "passphrase")
	m.Attach(attachmentFilePath)

	d := gomail.NewDialer(host, port, username, password)

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
