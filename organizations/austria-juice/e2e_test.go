package austria_juice_test

import (
	"encoding/json"
	"github.com/The-New-Fork/email-pipeline/pkg/factory"
	"github.com/The-New-Fork/email-pipeline/pkg/pipeline"
	"github.com/emersion/go-imap"
	"github.com/go-chi/render"
	"github.com/stretchr/testify/require"
	"github.com/unchainio/interfaces/logger"
	"github.com/unchainio/pkg/xconfig"
	"github.com/unchainio/pkg/xlogger"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestAustriaJuiceEndToEndSuccess(t *testing.T) {
	// TEST SETUP

	// send to austria-juice-staging@tnf-mail.unchain.io
	err := factory.SendAustriaJuiceMail(
		"test@tnf-mail.unchain.io",
		"%m9BkE&3EVdT",
		"tnf-mail.unchain.io",
		"test@tnf-mail.unchain.io",
		465,
		[]string{"austria-juice-staging@tnf-mail.unchain.io"},
		"./example.csv")
	require.NoError(t, err)
	log.Printf("send email")

	requestChannel := make(chan []byte)
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		resBody, err := ioutil.ReadAll(request.Body)
		require.NoError(t, err, "request body could not be read")
		err = request.Body.Close()
		require.NoError(t, err, "request body could not be closed")
		requestChannel <- resBody
		render.JSON(writer, request, map[string]interface{}{
			"key": "value",
		})
	})
	go http.ListenAndServe(":80", nil)

	log.Println("test setup complete")

	// TEST EXECUTION

	// load config & logger
	cfg := loadConfig()
	log, _ := xlogger.New(cfg.Logger)

	// create and start pipeline
	p := pipeline.New(cfg, log)

	err = p.Start()
	require.NoError(t, err)

	// TEST ASSERTIONS
	time.Sleep(5 * time.Second)

	// check for no failed messages
	errorCount := countErrorEmails(log, t)
	require.Equal(t, float64(0), errorCount, "error alert email expected to be 0")

	// check no emails with status unread remain
	unreadCount := countUnreadEmails(log, t)
	require.Equal(t, 0, len(unreadCount), "unread email count expected to be 0")

	// check http call rec
	resBody := <-requestChannel
	log.Println("received response on localhost:80/")

	expectedBody := `{
  "anfp": "11021034",
  "dfp": n/a,
  "bnfp": "B1253854",
  "pds": 1970-01-01,
  "pde": 2020-11-19,
  "jds": 0,
  "jde": 0,
  "bbd": 1970-01-01,
  "pc": "PL",
  "pl": "Bialobrzegi",
  "rmn": "n/a",
  "pon": "104/4500878994",
  "pop": "10",
}
`
	require.Equal(t, expectedBody, string(resBody), "http request body received from pipeline on import api does not match expected body")

	p.Stop()
}

func loadConfig() *pipeline.Config {
	// load config
	cfg := new(pipeline.Config)
	info := new(xconfig.Info)

	errs := xconfig.Load(
		cfg,
		xconfig.FromPathFlag("cfg", "./config.toml"),
		xconfig.FromEnv(),
		xconfig.GetInfo(info),
	)
	if errs != nil {
		log.Fatal(errs)
	}
	return cfg
}

func countUnreadEmails(l logger.Logger, t *testing.T) []uint32 {
	c, err := factory.NewImapClient(l, &factory.ImapConfig{
		Port:     ":993",
		Username: "austria-juice-staging@tnf-mail.unchain.io",
		Password: "QF9*!e52aoFr",
		Domain:   "tnf-mail.unchain.io",
	})
	require.NoError(t, err)

	criteria := &imap.SearchCriteria{
		WithoutFlags: []string{imap.SeenFlag},
	}

	seqNums, err := c.Client.Search(criteria)
	require.NoError(t, err)

	return seqNums
}

func countErrorEmails(l logger.Logger, t *testing.T) float64 {
	resp, err := http.Get("http://localhost:8025/api/v2/messages")
	require.NoError(t, err, "could not get message")
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err, "could not read smtp server API response body")

	var objectMap map[string]interface{}
	err = json.Unmarshal(bytes, &objectMap)
	require.NoError(t, err, "could not unmarshal local smtp server API response")

	return objectMap["count"].(float64)
}