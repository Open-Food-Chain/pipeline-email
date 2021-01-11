package pipeline

import (
	"fmt"
	"github.com/jmoiron/jsonq"
	"github.com/pkg/errors"
	"github.com/unchain/pipeline/pkg/actions/fileparser_action"
	"github.com/unchain/pipeline/pkg/actions/http_action"
	"github.com/unchain/pipeline/pkg/actions/imap_action"
	"github.com/unchain/pipeline/pkg/actions/templater_action"
	"github.com/unchain/pipeline/pkg/domain"
	"github.com/unchain/pipeline/pkg/triggers/cron_trigger"
)

func (p *Pipeline) Start() error {
	// Initialize trigger
	trigger := &cron_trigger.Trigger{}
	err := trigger.Init(p.log, []byte(p.cfg.Trigger.Config))
	if err != nil {
		return errors.Wrap(err, "could not init trigger")
	}

	p.log.Debugf("Initialized pipeline trigger")

	p.start(trigger)

	return nil
}

func (p *Pipeline) start(trigger domain.Trigger) {
	// start infinite loop to process messages
	for {
		tag, _, err := trigger.NextMessage()
		if err != nil {
			p.handleError(trigger, tag, err, 0)
		}
		p.log.Debugf("Next message with tag %v", tag)

		// check email
		imapOutput, err := imap_action.Invoke(p.log, map[string]interface{}{
			imap_action.ConfigInput: p.cfg.Actions.ImapAction.Config,
			imap_action.Function:    "GetNewMessageAttachments",
		})
		if err != nil {
			p.handleError(trigger, tag, err, 0)
		}
		// if no new messages, continue
		if imapOutput == nil {
			err = trigger.Respond(tag, nil, err)
			if err != nil {
				p.handleError(trigger, tag, err, 0)
			}
			continue
		}
		messages, ok := imapOutput["messages"].(map[uint32]interface{})
		if !ok {
			p.handleError(trigger, tag, errors.New("could not cast messages output from imap action"), 0)
		}

		if len(messages) < 1 {
			continue
		}
		p.log.Debugf("New messages: # %v", len(messages))

		// handle email messages
		seqNum, err := p.handleMessages(messages)
		if err != nil {
			p.handleError(trigger, tag, err, seqNum)
		}

		// call respond to finish processing
		err = trigger.Respond(tag, nil, err)
		if err != nil {
			p.handleError(trigger, tag, err, 0)
		}
	}
}

// handle email messages in loop
func (p *Pipeline) handleMessages(messages map[uint32]interface{}) (uint32, error) {
	for seqNum, message := range messages {
		// file parsing
		fileparserOutput, err := fileparser_action.Invoke(p.log, map[string]interface{}{
			fileparser_action.FileType:  p.cfg.Actions.FileparserAction.Filetype,
			fileparser_action.File:      message,
			fileparser_action.Header:    p.cfg.Actions.FileparserAction.Header,
			fileparser_action.Delimiter: ';',
		})
		if err != nil {
			return seqNum, errors.Wrap(err, fmt.Sprintf( "could not parse file in email with seqNum %v\n", seqNum))
		}
		p.log.Debugf("Parsed file, output: %v", fileparserOutput)
		// data transformation
		records, ok := fileparserOutput["messages"].([]map[string]interface{})
		if !ok {
			return seqNum, errors.Errorf("could not cast fileparser messages output for mail with seqNum: %v\n", seqNum)
		}

		err = p.handleRecords(records)
		if err != nil {
			return seqNum, errors.Wrap(err, fmt.Sprintf("error in email with seqNum %v - record handling stopped at index prior to error index\n", seqNum))
		}

		// mark message as read
		_, err = imap_action.Invoke(p.log, map[string]interface{}{
			imap_action.ConfigInput: p.cfg.Actions.ImapAction.Config,
			imap_action.Function:    "MarkMessageAsRead",
			imap_action.Params: map[string]interface{}{
				"seqNum": int(seqNum),
			},
		})
		if err != nil {
			return seqNum, errors.Wrap(err, fmt.Sprintf("error in email with seqNum %v\n", seqNum))
		}
	}

	return 0, nil
}

// handle product batch records in loop
func (p *Pipeline) handleRecords(records []map[string]interface{}) error {
	for index, record := range records {
		inputVariables := GetInputVariables(jsonq.NewQuery(record), p.cfg.Actions.TemplaterAction.Variables)
		templaterOutput, err := templater_action.Invoke(p.log, map[string]interface{}{
			templater_action.InputTemplate:  p.cfg.Actions.TemplaterAction.Template,
			templater_action.InputVariables: inputVariables,
		})
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not transform data for record with index %v", index))
		}

		// call import-api
		httpOutput, err := http_action.Invoke(p.log, map[string]interface{}{
			http_action.RequestBody: []byte(fmt.Sprintf("%s", templaterOutput[templater_action.TemplateResult])),
			http_action.Url:         p.cfg.Actions.HttpAction.Url,
			http_action.ContentType: "application/json",
			http_action.Method:      p.cfg.Actions.HttpAction.Method,
		})
		p.log.Debugf("invoked http action with result: %v \n and error: %v \n", httpOutput, err)
		if err != nil {
			return err
		}

		statusCode := httpOutput[http_action.ResponseStatusCode].(int)
		if statusCode != 200 {
			return errors.New(fmt.Sprintf("failed to call import-api for record with ID %v \n HTTP response: %v", index, httpOutput))
		}
	}
	return nil
}
