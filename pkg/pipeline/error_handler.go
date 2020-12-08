package pipeline

import (
	"fmt"
	"github.com/unchain/pipeline/pkg/actions/imap_action"
	"github.com/unchain/pipeline/pkg/actions/smtp_action"
	"github.com/unchain/pipeline/pkg/domain"
)

func (p *Pipeline) handleError(trigger domain.Trigger, tag string, err error, seqNum uint32) {
	var messageString string
	if seqNum != 0 {
		messageString = fmt.Sprintf("Email account: %s \n Email with seqNum %v is moved to mailbox 'Failed', message processing resulted in error:\n %v", p.cfg.Organization, seqNum, err)
		// move mail to failed email box
		_, err = imap_action.Invoke(p.log, map[string]interface{}{
			imap_action.ConfigInput: p.cfg.Actions.ImapAction.Config,
			imap_action.Function:    "MoveFailedMessage",
			imap_action.Params: map[string]interface{}{
				"seqNum": int(seqNum),
			},
		})
		if err != nil {
			p.log.Errorf("error while ")
		}
	} else {
		messageString = fmt.Sprintf("Email account: %s \n Message processing resulted in error:\n %v", p.cfg.Organization, err)
	}

	// send alert email
	_, err = smtp_action.Invoke(p.log, map[string]interface{}{
		"username":   p.cfg.Actions.SmtpAction.Username,
		"password":   p.cfg.Actions.SmtpAction.Password,
		"hostname":   p.cfg.Actions.SmtpAction.Hostname,
		"port":       p.cfg.Actions.SmtpAction.Port,
		"from":       p.cfg.Actions.SmtpAction.From,
		"recipients": p.cfg.Actions.SmtpAction.Recipients,
		"message":    []byte(messageString),
	})
	if err != nil {
		p.log.Errorf("Could not handle error, msg: %v", err)
	}
}
