package mixpanel

import (
	"testing"
)

const token string = "e919dea023855e3c8e2ea46a38e4032c"

func TestUpdate(t *testing.T) {
	p := &P{}
	p.Update(&P{
		"Test" : "Test",
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
