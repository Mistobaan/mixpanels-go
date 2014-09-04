/*
The mixpanel module allows you to easily track events and
update people properties from your Go application.

The Mixpanel class is the primary class for tracking events and
sending people analytics updates.

The StdConsumer and BuffConsumer classes allow callers to
customize the IO characteristics of their tracking.
*/

package mixpanel

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type P map[string]interface{}

// Update replaces all the elements of the map
func (this *P) Update(other *P) *P {
	if other == nil {
		return this
	}
	for k, v := range *other {
		(*this)[k] = v
	}
	return this
}

type Event struct {
	Event      string `json:"event"`
	Properties *P     `json:"properties"`
}

type Consumer interface {
	Send(endpoint string, json_msg []byte) error
}

type Mixpanel struct {
	Token   string `json:token`
	verbose bool
	c       Consumer
}

const events_endpoint string = "https://api.mixpanel.com/track"
const people_endpoint string = "https://api.mixpanel.com/engage"

var import_endpoint string = "https://api.mixpanel.com/import"

func b64(payload []byte) []byte {
	var b bytes.Buffer
	encoder := base64.NewEncoder(base64.URLEncoding, &b)
	encoder.Write(payload)
	encoder.Close()
	return b.Bytes()[:b.Len()]
}

/*
NewMixpanel Creates a new Mixpanel object, which can be used for all tracking.

To use mixpanel, create a new Mixpanel object using your
token.  Takes in a user token and uses a StdConsumer
*/
func NewMixpanel(token string) *Mixpanel {
	return NewMixpanelWithConsumer(token, NewStdConsumer())
}

/*
NewMixpanelWithConsumer Creates a new Mixpanel object, which can be used for all tracking.

To use mixpanel, create a new Mixpanel object using your
token.  Takes in a user token and an optional Consumer (or
anything else with a send() method). If no consumer is
provided, Mixpanel will use the default Consumer, which
communicates one synchronous request for every message.
*/
func NewMixpanelWithConsumer(token string, c Consumer) *Mixpanel {
	return &Mixpanel{
		Token:   token,
		verbose: true,
		c:       c,
	}
}

/*
Notes that an event has occurred, along with a distinct_id
representing the source of that event (for example, a user id),
an event name describing the event and a set of properties
describing that event. Properties are provided as a Hash with
string keys and strings, numbers or booleans as values.

// Track that user "12345"'s credit card was declined
mp.Track("12345", "Credit Card Declined", nil)

// Properties describe the circumstances of the event,
// or aspects of the source or user associated with the event
mp.Track("12345", "Welcome Email Sent", &P{
  "Email Template" : "Pretty Pink Welcome",
  "User Sign-up Cohort" : "July 2013",
 })
*/
func (mp *Mixpanel) Track(distinct_id, event string, prop *P) error {
	import_endpoint += "?api_key=" + mp.Token
	return mp.sendEvent(distinct_id, event, prop, "events")
}

/*
Imports events that occurred more than 5 days in the past. Takes the
same arguments as Track and behaves in the same way.
*/
func (mp *Mixpanel) Import(distinct_id, event string, prop *P) error {
	return mp.sendEvent(distinct_id, event, prop, "import")
}

/* Internal implementation of event sending. Can be used with Track or Import. */
func (mp *Mixpanel) sendEvent(distinct_id, event string, prop *P, endpoint string) error {
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

	return mp.c.Send(endpoint, data)
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
		"alias":       alias_id,
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
func (mp *Mixpanel) PeopleUpdate(properties *P) error {
	record := &P{
		"$token": mp.Token,
		"$time":  int(time.Now().UTC().Unix()),
	}
	record.Update(properties)

	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return mp.c.Send("people", data)
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
		"$set":         properties,
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
		"$set":         properties,
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
		"$add":         properties,
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
		"$append":      properties,
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
		"$union":       properties,
	})
}

/*
PeopleUnset removes properties from a profile.

Takes a JSON list of string property names, and permanently removes the
properties and their values from a profile.
Example:
    mp.PeopleUnset("12345", ["Days Overdue"])
*/
func (mp *Mixpanel) PeopleUnset(id string, properties []string) error {
	return mp.PeopleUpdate(&P{
		"$distinct_id": id,
		"$unset":       properties,
	})
}

/*
PeopleDelete permanently deletes a profile.

Permanently delete the profile from Mixpanel, along with all of its
properties.
Example:
    mp.PeopleDelete("12345")
*/
func (mp *Mixpanel) PeopleDelete(id string) error {
	return mp.PeopleUpdate(&P{
		"$distinct_id": id,
		"$delete":      "",
	})
}

/*
PeopleTrackCharge Tracks a charge to a user.

Record that you have charged the current user a certain amount of
money. Charges recorded with track_charge will appear in the Mixpanel
revenue report.
Example:
    //tracks a charge of $50 to user '1234'
    mp.PeopleTrackCharge("1234", 50, nil)

    //tracks a charge of $50 to user '1234' at a specific time
    mp.PeopleTrackCharge("1234", 50, {"$time": "2013-04-01T09:02:00"})
*/
func (mp *Mixpanel) PeopleTrackCharge(id string, amount float64, prop *P) error {
	if prop == nil {
		prop = &P{}
	}
	prop.Update(&P{"$amount": amount})
	return mp.PeopleAppend(id, &P{
		"$transactions": prop,
	})
}

func parseJsonResponse(resp *http.Response) error {
	type jsonResponseT map[string]interface{}
	var response jsonResponseT
	var buff bytes.Buffer
	io.Copy(&buff, resp.Body)

	if err := json.Unmarshal(buff.Bytes(), &response); err == nil {
		if value, ok := response["status"]; ok {
			if value.(float64) == 1 {
				return nil
			} else {
				return errors.New(fmt.Sprintf("Mixpanel error: %s", response["error"]))
			}
		} else {
			return errors.New("Could not find field 'status' api change ?")
		}
	}
	return errors.New("Cannot interpret Mixpanel server response: " + buff.String())
}

type StdConsumer struct {
	endpoints map[string]string
}

// Creates a new StdConsumer.
// Sends one message at a time
func NewStdConsumer() *StdConsumer {
	c := new(StdConsumer)
	c.endpoints = make(map[string]string)
	c.endpoints["events"] = events_endpoint
	c.endpoints["people"] = people_endpoint
	c.endpoints["import"] = import_endpoint
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

type BuffConsumer struct {
	StdConsumer
	buffers map[string][][]byte
	maxSize int64
}

func NewBuffConsumer(maxSize int64) *BuffConsumer {
	bc := new(BuffConsumer)
	bc.StdConsumer = *NewStdConsumer()
	bc.maxSize = maxSize
	bc.buffers = make(map[string][][]byte)
	bc.buffers["people"] = make([][]byte, 0, maxSize)
	bc.buffers["events"] = make([][]byte, 0, maxSize)
	bc.buffers["import"] = make([][]byte, 0, maxSize)
	return bc
}

func (bc *BuffConsumer) Send(endpoint string, msg []byte) error {
	if _, ok := bc.buffers[endpoint]; !ok {
		return errors.New(fmt.Sprintf("No such endpoint '%s'. Valid endpoints are one of %#v", endpoint, bc.buffers))
	}
	bc.buffers[endpoint] = append(bc.buffers[endpoint], msg)
	if len(bc.buffers[endpoint]) > int(bc.maxSize) {
		bc.flushEndpoint(endpoint)
	}
	return nil
}

/*
Flush Send all remaining messages to Mixpanel. BufferedConsumers will
flush automatically when you call Send(), but you will need to call
Flush() when you are completely done using the consumer (for example,
when your application exits) to ensure there are no messages remaining
in memory.
*/
func (bc *BuffConsumer) Flush() error {
	for endpoint := range bc.buffers {
		bc.flushEndpoint(endpoint)
	}
	return nil
}

func jsonArray(a [][]byte) []byte {
	sep := ","
	if len(a) == 0 {
		return []byte("[]")
	}

	n := len(sep) * (len(a) - 1)
	for i := 0; i < len(a); i++ {
		n += len(a[i])
	}

	b := make([]byte, n+2)
	bp := copy(b, []byte{'['})
	bp += copy(b[bp:], a[0])
	for _, s := range a[1:] {
		bp += copy(b[bp:], sep)
		bp += copy(b[bp:], s)
	}
	copy(b[bp:], []byte{']'})
	return b
}

func (bc *BuffConsumer) flushEndpoint(endpoint string) error {
	msg := jsonArray(bc.buffers[endpoint])
	return bc.StdConsumer.Send(endpoint, msg)
}
