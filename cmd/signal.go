package cmd

import (
	"context"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
)

func cancelOnSignal(ctx context.Context, signs ...os.Signal) context.Context {
	recv := make(chan os.Signal, 1)
	signal.Notify(recv, signs...)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()
		logrus.Warnf("receive signal: %+v", <-recv)
	}()
	return ctx
}
