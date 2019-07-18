// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package gcp

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	cloudfunctions "google.golang.org/api/cloudfunctions/v1"

	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/x-pack/functionbeat/manager/executor"
)

type opDeleteFunction struct {
	log      *logp.Logger
	location string
	name     string
	tokenSrc oauth2.TokenSource
}

func newOpDeleteFunction(
	log *logp.Logger,
	location string,
	name string,
	tokenSrc oauth2.TokenSource,
) *opDeleteFunction {
	return &opDeleteFunction{
		log:      log,
		location: location,
		name:     name,
		tokenSrc: tokenSrc,
	}
}

// Execute creates a function from the zip uploaded.
func (o *opDeleteFunction) Execute(ctx executor.Context) error {
	c, ok := ctx.(*functionContext)
	if !ok {
		return errWrongContext
	}

	client := oauth2.NewClient(context.TODO(), o.tokenSrc)
	svc, err := cloudfunctions.New(client)
	if err != nil {
		return fmt.Errorf("error while creating cloud functions service: %+v", err)
	}

	functionSvc := cloudfunctions.NewProjectsLocationsFunctionsService(svc)
	operation, err := functionSvc.Delete(o.name).Context(context.TODO()).Do()
	if err != nil {
		return fmt.Errorf("error while removing function %s: %+v", o.name, err)
	}

	c.name = &operation.Name

	if operation.Done {
		o.log.Debugf("Function %s removed successfully", o.name)
	}

	return nil
}

// Rollback
func (o *opDeleteFunction) Rollback(_ executor.Context) error {
	return nil
}
