package main

import (
	"strconv"
	"strings"
	"log"
	"os"

	mixpanel "github.com/mixpanel/mixpanel-go"
)

func check(err error){
	if err != nil {
		log.Fatal(err)
	}
}

func extractProperties(cmds []string) *mixpanel.P {
	props := &mixpanel.P{}
	for _, element := range cmds[2:] {
		idx := strings.Index(element, "=")
		if idx != -1 {
			(*props)[element[:idx]] = element[idx+1:]
		} else {
			log.Fatalf("Invalid argument %s", element)
		}
	}
	return props
}

// export MIXPANEL_TOKEN=
// track id event_name a=b c=d d=e 
// track 
// 
func main() {
	token := os.Getenv("MIXPANEL_TOKEN")
	if len(token) == 0 {
		log.Fatal("Please Set MIXPANEL_TOKEN env variable")
	} 

	mp := mixpanel.NewMixpanel(token)
	if len(os.Args) < 2 {
		log.Fatal("not enough arguments")
	}
	cmds := os.Args[1:]

	switch cmds[0] {
	case "track":
		if len(cmds) < 3 {
			log.Fatal("not enough arguments for track")
		} else if len(cmds) == 3 {
			check(mp.Track(cmds[1], cmds[2], nil))
		} else {
			check(mp.Track(cmds[1], cmds[2], extractProperties(cmds[2:])))
		}
	case "alias":
		if len(cmds) < 2 {
			log.Fatal("not enough arguments for alias")
		} else {
			check(mp.Alias(cmds[1], cmds[2]))
		} 
	case "set":
		if len(cmds) < 2 {
			log.Fatal("not enough arguments for set")
		} else {
			check(mp.PeopleSet(cmds[1], extractProperties(cmds[1:])))
		} 
	case "set_once":
		if len(cmds) < 2 {
			log.Fatal("not enough arguments for set_once")
		} else {
			check(mp.PeopleSetOnce(cmds[1], extractProperties(cmds[1:])))
		}
	case "add":
		if len(cmds) < 2 {
			log.Fatal("not enough arguments for add [id] [key=value]*")
		} else {
			check(mp.PeopleIncrement(cmds[1], extractProperties(cmds[1:])))
		}
	case "append":
		if len(cmds) < 2 {
			log.Fatal("not enough arguments for append [id] [key=value]*")
		} else {
			check(mp.PeopleAppend(cmds[1], extractProperties(cmds[1:])))
		}
	case "union":
		if len(cmds) < 2 {
			log.Fatal("not enough arguments for union [id] [key=value]*")
		} else {
			check(mp.PeopleUnion(cmds[1], extractProperties(cmds[1:])))
		}
	case "unset":
		if len(cmds) < 2 {
			log.Fatal("not enough arguments for unset [id] [*values]")
		} else {
			check(mp.PeopleUnset(cmds[1], cmds[1:]))
		}
	case "delete":
		if len(cmds) < 2 {
			log.Fatal("not enough arguments for delete <id>")
		} else {
			check(mp.PeopleDelete(cmds[1]))
		}
	case "charge":
		if len(cmds) < 3 {
			log.Fatal("not enough arguments for delete <id>")
		} else {
			amount, err := strconv.ParseFloat(cmds[2], 64)
			check(err)
			check(mp.PeopleTrackCharge(cmds[1], amount, extractProperties(cmds[2:])))
		}
	case "help":
		log.Fatal("You are on your own")
	default:
		log.Fatalf("Unknown command %s", cmds[0])
	}

}
