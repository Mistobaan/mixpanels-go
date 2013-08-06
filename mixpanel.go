package mixpanel

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type P map[string]string

type Event struct {
	Event      string `json:event`
	Properties *P     `json:properties`
}

type Mixpanel struct {
	Token string `json:token`
}

const track_endpoint string = "https://api.mixpanel.com/track"
const people_endpoint string = "https://api.mixpanel.com/engage"

func b64(payload []byte) []byte {
	var b bytes.Buffer
	encoder := base64.NewEncoder(base64.URLEncoding, &b)
	encoder.Write(payload)
	encoder.Close()
	return b.Bytes()[:b.Len()]
}

func NewMixpanel(token string) *Mixpanel {
	return &Mixpanel{
		Token: token,
	}
}

func (this *P) Update(other *P) *P {
	for k, v := range *other {
		(*this)[k] = v
	}
	return this
}

// Track 
func (mix *Mixpanel) Track(distinct_id, event string, prop *P) error {
	track_url, err := url.Parse(track_endpoint)
	if err != nil {
		return err
	}

	data, err := json.Marshal(&Event{
		Event: event,
		Properties: &P{
			"token":        mix.Token,
			"distinct_id":  distinct_id,
			"time":         time.Now().UTC().Format(time.UnixDate),
			"mp_lib":       "go",
			"$lib_version": "1.0",
		},
	})
	if err != nil {
		return err
	}

	track_url.Query().Add("data", string(b64(data)))
	track_url.Query().Add("verbose", "1")

	resp, err := http.Get(track_url.String())

	if err != nil {
		return err
	}

	io.Copy(os.Stdout, resp.Body)
	return nil
}

func (mp *Mixpanel) Alias(new_internal_id, original_anonymous_id string) error {
	return nil
}
