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
	"os"
	"testing"
	"time"
)

func skipCI(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}
}

func TestAustriaJuiceEndToEndSuccess(t *testing.T) {
	
	// TEST SETUP
	skipCI(t)

	// send to austria-juice-staging@tnf-mail.unchain.io
	go factory.SendAustriaJuiceMail(  	
		"ajpipelinetest@gmail.com",
        	"pleasechangeme",
        	"smtp.gmail.com",
        	"ajpipelinetest@gmail.com",
        	465,
        	[]string{"ajpipelinetest@gmail.com"},
		"./example.csv")
	//require.NoError(t, err)
	log.Printf("send email")
 
	requestChannel := make(chan []byte)
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		resBody, err := ioutil.ReadAll(request.Body)
		require.NoError(t, err, "request body could not be read")
		request.Body.Close()
		//require.NoError(t, err, "request body could not be closed")
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

	go p.Start()
	//require.NoError(t, err)

	// TEST ASSERTIONS
	// check http call rec
	resBody := <-requestChannel

	//log.Println("chris is cool")

	time.Sleep(5 * time.Second)

	expectedBody := `{
  "anfp": "11021034",
  "dfp": n/a,
  "bnfp": "B1253854",
  "pds": 1970-01-01,
  "pde": 2020-01-06,
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

	// check for no failed messages
	errorCount := countErrorEmails(log, t)
	require.Equal(t, float64(0), errorCount, "error alert email expected to be 0")

	// check no emails with status unread remain
	unreadCount := countUnreadEmails(log, t)
	require.Equal(t, 0, len(unreadCount), "unread email count expected to be 0")

	p.Stop()
}

func TestAustriaJuiceEndToEndFailover(t *testing.T) {
	skipCI(t)

	// TEST SETUP

	// send to austria-juice-staging@tnf-mail.unchain.io
	err := factory.SendAustriaJuiceMail(
		"test@tnf-mail.unchain.io",
		"%m9BkE&3EVdT",
		"tnf-mail.unchain.io",
		"test@tnf-mail.unchain.io",
		465,
		[]string{"austria-juice-staging@tnf-mail.unchain.io"},
		"./bad_example.csv")
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
	// check http call rec
	time.Sleep(5 * time.Second)

	// check for one failed messages
	errorCount := countErrorEmails(log, t)
	require.Equal(t, float64(1), errorCount, "error alert email expected to be 1")

	// check no emails with status unread remain
	failedCount := countFailedEmails(log, t)
	require.Equal(t, 1, len(failedCount), "unread email count expected to be 0")

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

func countFailedEmails(l logger.Logger, t *testing.T) []uint32 {
	c, err := factory.NewImapClient(l, &factory.ImapConfig{
		Port:     ":993",
		Username: "austria-juice-staging@tnf-mail.unchain.io",
		Password: "QF9*!e52aoFr",
		Domain:   "tnf-mail.unchain.io",
	})
	require.NoError(t, err)

	_, err = c.Client.Select("Failed", false)
	require.NoError(t, err, "cannot select failed inbox")

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
