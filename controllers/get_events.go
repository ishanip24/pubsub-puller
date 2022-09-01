package controllers

import (
	"context"
	"reflect"
	"sync"

	ctrl "sigs.k8s.io/controller-runtime"

	"cloud.google.com/go/pubsub"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type UserEvents struct {
	ctx          context.Context
	log          logr.Logger
	client       client.Client
	subscription *pubsub.Subscription
	lock         sync.RWMutex
	users        chan<- event.GenericEvent
}

type Event struct {
	UserName string `json:"userName,omitempty"`
	Project  string `json:"project,omitempty"`
}

func CreateUserEvents(client client.Client, subscription *pubsub.Subscription, users chan<- event.GenericEvent) UserEvents {
	log := ctrl.Log.
		WithName("source").
		WithName(reflect.TypeOf(UserEvents{}).Name())
	return UserEvents{
		ctx:          context.Background(),
		log:          log,
		client:       client,
		subscription: subscription,
		lock:         sync.RWMutex{},
		users:        users,
	}
}

// keep recieving pubsub messages until there are none left
func (t *UserEvents) Run() {
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		t.checkSubscription()
	}
}

// stores event in users prop in UserEvents
func (t *UserEvents) checkSubscription() {
	if t.subscription.String() != "pull-test-results" {
		t.log.Info("Subscription names are equal")
	} else {
		t.log.Info("Subscription names are different")
	}
}
