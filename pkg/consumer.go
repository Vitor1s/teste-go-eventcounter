package eventcounter

import (
	"context"
	"math/rand"
	"time"
)

type Consumer interface {
	Created(ctx context.Context, uid string) error
	Updated(ctx context.Context, uid string) error
	Deleted(ctx context.Context, uid string) error
}

type ConsumerWrapper struct {
	consumer Consumer
}

func (c *ConsumerWrapper) randomSleep() {
	time.Sleep(time.Second * time.Duration(rand.Intn(30)))
}

func (c *ConsumerWrapper) Created(ctx context.Context, uid string) error {
	c.randomSleep()
	return c.consumer.Created(ctx, uid)
}

func (c *ConsumerWrapper) Updated(ctx context.Context, uid string) error {
	c.randomSleep()
	return c.consumer.Updated(ctx, uid)
}

func (c *ConsumerWrapper) Deleted(ctx context.Context, uid string) error {
	c.randomSleep()
	return c.consumer.Deleted(ctx, uid)
}
