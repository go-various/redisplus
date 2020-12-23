package redisplus

import (
	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
	"testing"
	"time"
)

func TestNewNotification(t *testing.T) {
	cmd, _ := NewRedisCmd(&Config{
		Addrs:              []string{"localhost:6379"},
		Password:           "",
		KeyPrefix:          "TEST",
	})
	logger := hclog.Default()
	logger = logger.ResetNamed("test")
	logger.SetLevel(hclog.Trace)
	view := NewRedisView(cmd,"dev", nil)
	policy := []time.Duration{
		time.Second * 2,
		time.Second * 5,
		time.Second * 10,
	}
	notify, _ := NewNotification("order", view, logger,policy)
	if notify == nil{
		t.Fatal(notify)
	}
	entity := &Entity{
		count: 0,
		Key:   uuid.New().String(),
		Value: []byte("test-notify"),
	}
	notify.Subscribe(func(p *Entity, err error) PutNext {
		t.Log(p, err)
		return true
	})
	notify.PutNotification(entity)
	done := make(chan byte)
	<-done
}
