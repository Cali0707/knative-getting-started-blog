/*
Copyright 2019 The Knative Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
		http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package adapter implements a sample receive adapter that generates events
// at a regular interval.
package adapter

import (
	"bytes"
	"context"
	"text/template"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"go.uber.org/zap"

	"knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"
)

type envConfig struct {
	// Include the standard adapter.EnvConfig used by all adapters.
	adapter.EnvConfig

	// Interval between events, for example "5s", "100ms"
	Interval time.Duration `envconfig:"INTERVAL" required:"true"`

	MessageTemplate string `envconfig:"MESSAGE_TEMPLATE" required:"true"`

	ConfigVars map[string]string `envconfig:"CONFIG_VARS"`
}

func NewEnv() adapter.EnvConfigAccessor { return &envConfig{} }

// Adapter generates events at a regular interval.
type Adapter struct {
	client          cloudevents.Client
	interval        time.Duration
	messageTemplate template.Template
	logger          *zap.SugaredLogger
	configVars      map[string]string

	nextID int
}

type dataExample struct {
	Sequence  int    `json:"sequence"`
	Heartbeat string `json:"heartbeat"`
	Message   string `json:"message"`
}

func (a *Adapter) newEvent() cloudevents.Event {
	event := cloudevents.NewEvent()
	event.SetType("dev.knative.sample")
	event.SetSource("sample.knative.dev/heartbeat-source")

	a.logger.Warn("interval", zap.Duration("interval", a.interval))

	var message bytes.Buffer
	if err := a.messageTemplate.Execute(&message, a.configVars); err != nil {
		a.logger.Errorw("Failed to execute message template", err)
	}

	if err := event.SetData(cloudevents.ApplicationJSON, &dataExample{
		Sequence:  a.nextID,
		Heartbeat: a.interval.String(),
		Message:   message.String(),
	}); err != nil {
		a.logger.Errorw("failed to set data")
	}
	a.nextID++
	return event
}

// Start runs the adapter.
// Returns if ctx is cancelled or Send() returns an error.
func (a *Adapter) Start(ctx context.Context) error {
	a.logger.Infow("Starting heartbeat", zap.String("interval", a.interval.String()))
	for {
		select {
		case <-time.After(a.interval):
			event := a.newEvent()
			a.logger.Infow("Sending new event", zap.String("event", event.String()))
			if result := a.client.Send(context.Background(), event); !cloudevents.IsACK(result) {
				a.logger.Infow("failed to send event", zap.String("event", event.String()), zap.Error(result))
				// We got an error but it could be transient, try again next interval.
				continue
			}
		case <-ctx.Done():
			a.logger.Info("Shutting down...")
			return nil
		}
	}
}

func NewAdapter(ctx context.Context, aEnv adapter.EnvConfigAccessor, ceClient cloudevents.Client) adapter.Adapter {
	env := aEnv.(*envConfig) // Will always be our own envConfig type
	logger := logging.FromContext(ctx)
	logger.Infow("Heartbeat example", zap.Duration("interval", env.Interval))
	logger.Warn("interval", zap.Duration("interval", env.Interval))
	messageTemplate, _ := template.New("samplesource.message.template").Parse(env.MessageTemplate)
	return &Adapter{
		interval:        env.Interval,
		messageTemplate: *messageTemplate,
		client:          ceClient,
		logger:          logger,
		configVars:      env.ConfigVars,
	}
}
