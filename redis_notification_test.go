package redisplus

import (
	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
	"testing"
	"time"
)

func TestNewNotification(t *testing.T) {
	cfg := &Config{
		Addrs:              []string{"localhost:6379"},
		Password:           "",
		KeyPrefix:          "TEST",
	}

	view, _ := NewRedisCli(cfg,"dev")
	policies := []time.Duration{
		time.Second * 10,
		time.Second * 15,
		time.Second * 30,
	}

	t.Log(policies)
	notify, err := NewNotification("order", view, hclog.Default(), policies)
	if err != nil {
		t.Fatal(err)
		return
	}

	if notify == nil{
		t.Fatal(notify)
	}
	entity := &Entity{
		count: 0,
		Key:   uuid.New().String(),
		Value: []byte("test-notify"),
	}

	notify.Subscribe(func(p *Entity, err error) PutNext {
		t.Log(err, p.Key, p.Value, p.valueKey(), p.notifyKey())
		return p.count == 1
	})

	notify.PutNotification(entity)
	done := make(chan byte)
	<-done
}