package mixpanel

import (
	"testing"
)

const token string = "e919dea023855e3c8e2ea46a38e4032c"

func TestUpdate(t *testing.T) {
	p := &P{}
	p.Update(&P{
		"Test": "Test",
	})

	if _, ok := (*p)["Test"]; !ok {
		t.Error("Expected Test got %*v", *p)
	}

}

func TestTrack(t *testing.T) {
	mix := NewMixpanel(token)

	err := mix.Track("userId", "Plan Upgraded", &P{
		"Old Plan": "Business",
		"New Plan": "Premium",
	})

	if err != nil {
		t.Error(err)
	}
}

func TestJsonArray(t *testing.T) {
	result := jsonArray([][]byte{[]byte("{'a':'b'}")})
	if string(result) != "[{'a':'b'}]" {
		t.Error(string(result))
	}
	result = jsonArray([][]byte{[]byte("{'a':'b'}"), []byte("{'c':'d'}")})
	if string(result) != "[{'a':'b'},{'c':'d'}]" {
		t.Error(string(result))
	}
}

func TestSmoke(t *testing.T) {
	Smoke(t, NewMixpanel(token))
	Smoke(t, NewMixpanelWithConsumer(token, NewBuffConsumer(1)))
	mp := NewBuffConsumer(2)
	Smoke(t, NewMixpanelWithConsumer(token, mp))
	mp.Flush()
}

func Smoke(t *testing.T, mp *Mixpanel) {

	err := mp.PeopleSet("12345", &P{"Address": "1313 Mockingbird Lane",
		"Birthday": "1948-01-01"})
	if err != nil {
		t.Error(err)
	}

	err = mp.Alias("amy@mixpanel.com", "13793")
	if err != nil {
		t.Error(err)
	}

	// Import an older event
	err = mp.Import("12345", "Welcome Email Sent", &P{
		"time": 1392646952,
	})
	if err != nil {
		t.Error(err)
	}

	// Track that user "12345"'s credit card was declined
	err = mp.Track("12345", "Credit Card Declined", nil)
	if err != nil {
		t.Error(err)
	}

	// Properties describe the circumstances of the event,
	// or aspects of the source or user associated with the event
	err = mp.Track("12345", "Welcome Email Sent", &P{
		"Email Template":      "Pretty Pink Welcome",
		"User Sign-up Cohort": "July 2013",
	})
	if err != nil {
		t.Error(err)
	}

	err = mp.PeopleAppend("12345", &P{
		"Favorite Fruits": "Apples",
	})
	if err != nil {
		t.Error(err)
	}

	err = mp.PeopleSetOnce("12345", &P{"First Login": "2013-04-01T13:20:00"})
	if err != nil {
		t.Error(err)
	}

	err = mp.PeopleIncrement("12345", &P{"Coins Gathered": 12})
	if err != nil {
		t.Error(err)
	}

	err = mp.PeopleAppend("12345", &P{"Power Ups": "Bubble Lead"})
	if err != nil {
		t.Error(err)
	}

	err = mp.PeopleUnion("12345", &P{"Items purchased": []string{"socks", "shirts"}})
	if err != nil {
		t.Error(err)
	}

	err = mp.PeopleUnset("12345", []string{"Days Overdue"})
	if err != nil {
		t.Error(err)
	}

	//tracks a charge of $50 to user '1234'
	err = mp.PeopleTrackCharge("1234", 50, nil)
	if err != nil {
		t.Error(err)
	}

	//tracks a charge of $50 to user '1234' at a specific time
	err = mp.PeopleTrackCharge("1234", 50, &P{"$time": "2013-04-01T09:02:00"})
	if err != nil {
		t.Error(err)
	}

	err = mp.PeopleTrackCharge("12345", 9.99, nil)
	if err != nil {
		t.Error(err)
	}

	err = mp.PeopleTrackCharge("12345", 30.50, &P{
		"$time":            "2013-01-02T09:32:00",
		"Product Category": "Shoes",
	})

	if err != nil {
		t.Error(err)
	}
}
