package mixpanel

import (
	"fmt"
	"errors"
	"io"
	"strconv"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

type P map[string]interface{}

type Event struct {
	Event      string `json:"event"`
	Properties *P     `json:"properties"`
}

type Consumer interface {
	Send(endpoint string, json_msg []byte) error
}

type Mixpanel struct {
	Token string `json:token`
	verbose bool 
	c Consumer
}

const events_endpoint string = "https://api.mixpanel.com/track"
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
		verbose: true,
		c: NewStdConsumer(),
	}
}

func (this *P) Update(other *P) *P {
	for k, v := range *other {
		(*this)[k] = v
	}
	return this
}

// Track
func (mp *Mixpanel) Track(distinct_id, event string, prop *P) error {
	properties := &P{
		"token":        mp.Token,
		"distinct_id":  distinct_id,
		"time":         strconv.FormatInt(time.Now().UTC().Unix(), 10),
		"mp_lib":       "go",
		"$lib_version": "0.1",
	}
	if prop == nil {
		prop = &P{}
	}

	properties.Update(prop)

	data, err := json.Marshal(&Event{
		Event:      event,
		Properties: properties,
	})
	if err != nil {
		return err
	}

	return mp.c.Send("events", data)
}

/*
Alias gives custom alias to a people record.

Alias sends an update to our servers linking an existing distinct_id
with a new id, so that events and profile updates associated with the
new id will be associated with the existing user's profile and behavior.
Example:
    mp.Alias("amy@mixpanel.com", "13793")
*/        
func (mp *Mixpanel) Alias(alias_id, original_id string) error {
	return mp.Track(original_id, "$create_alias", &P{
            "distinct_id": original_id,
            "alias": alias_id,
        })
}

/*
PeopleUpdate sends a generic update to Mixpanel people analytics.
Caller is responsible for formatting the update message, as
documented in the Mixpanel HTTP specification, and passing
the message as a dict to update. This
method might be useful if you want to use very new
or experimental features of people analytics from python
The Mixpanel HTTP tracking API is documented at
https://mixpanel.com/help/reference/http
*/
func (mp *Mixpanel) PeopleUpdate(alias_id string, properties *P) error {
	return nil
}

/*
PeopleSet set properties of a people record.

PeopleSet sets properties of a people record given in JSON object. If the profile
does not exist, creates new profile with these properties.
Example:
    mp.PeopleSet("12345", &P{"Address": "1313 Mockingbird Lane",
                            "Birthday": "1948-01-01"})
*/
func (mp *Mixpanel) PeopleSet(id string, properties *P) error {
	return mp.PeopleUpdate(&P{
		"$distinct_id": id,
		"$set" : properties,
		})
}

/*
PeopleSetOnce sets immutable properties of a people record.

PeopleSetOnce sets properties of a people record given in JSON object. If the profile
does not exist, creates new profile with these properties. Does not
overwrite existing property values.
Example:
    mp.PeopleSetOnce("12345", &P{"First Login": "2013-04-01T13:20:00"})
*/
func (mp *Mixpanel) PeopleSetOnce(id string, properties *P) error {
	return mp.PeopleUpdate(&P{
		"$distinct_id": id,
		"$set" : properties,
	})
}

/*
PeopleIncrement Increments/decrements numerical properties of people record.

Takes in JSON object with keys and numerical values. Adds numerical
values to current property of profile. If property doesn't exist adds
value to zero. Takes in negative values for subtraction.
Example:
    mp.PeopleIncrement("12345", &P{"Coins Gathered": 12})
*/
func (mp *Mixpanel) PeopleIncrement(id string, properties *P) error {
	return mp.PeopleUpdate(&P{
		"$distinct_id": id,
		"$add" : properties,
	})
}

/*
PeopleAppend appends to the list associated with a property.

Takes a JSON object containing keys and values, and appends each to a
list associated with the corresponding property name. $appending to a
property that doesn't exist will result in assigning a list with one
element to that property.
Example:
    mp.PeopleAppend("12345", &P{ "Power Ups": "Bubble Lead" })
*/
func (mp *Mixpanel) PeopleAppend(id string, properties *P) error {
	return mp.PeopleUpdate(&P{
		"$distinct_id": id,
		"$append" : properties,
	})
}

/*
PeopleUnion Merges the values for a list associated with a property.

Takes a JSON object containing keys and list values. The list values in
the request are merged with the existing list on the user profile,
ignoring duplicate list values.
Example:
    mp.PeopleUnion("12345", &P{ "Items purchased": ["socks", "shirts"] } )
*/
func (mp *Mixpanel) PeopleUnion(id string, properties *P) error {
	return mp.PeopleUpdate(&P{
		"$distinct_id": id,
		"$union" : properties,
	})
}



func parseJsonResponse(resp *http.Response) error {
	type jsonResponseT map[string]interface{}
	var response jsonResponseT
	var buff bytes.Buffer
	io.Copy(&buff, resp.Body)

	if err := json.Unmarshal(buff.Bytes(), &response); err == nil{
		if value, ok := response["status"]; ok {
			if value.(float64) == 1 {
				return nil
			} else {
				return errors.New( fmt.Sprintf("Mixpanel error: %s", response["error"]))
			}
		} else {
			return errors.New("Could not find field 'status' api change ?")
		}
	}
	return errors.New("Cannot interpret Mixpanel server response: "+buff.String())
}

type StdConsumer struct {
	endpoints map[string]string
}

func NewStdConsumer() Consumer {
	c := new(StdConsumer)
	c.endpoints = make(map[string]string)
	c.endpoints["events"] = events_endpoint
	c.endpoints["people"] = people_endpoint
	return c
}

func (c *StdConsumer) Send(endpoint string, msg []byte) error {

	if url, ok := c.endpoints[endpoint]; !ok {
		return errors.New(fmt.Sprintf("No such endpoint '%s'. Valid endpoints are one of %#v", endpoint, c.endpoints))
	} else {
		return c.write(url, msg)
	}
}

func (c *StdConsumer) write(endpoint string, msg []byte) error {
	track_url, err := url.Parse(endpoint)
	if err != nil {
		return err
	}

	q := track_url.Query()
	q.Add("data", string(b64(msg)))
	q.Add("verbose", "1")

	track_url.RawQuery = q.Encode()

	resp, err := http.Get(track_url.String())

	if err != nil {
		return err
	}

	return parseJsonResponse(resp)
}
