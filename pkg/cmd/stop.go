/*
* Copyright 2020-present Arpabet Inc. All rights reserved.
 */


package cmd

import (
	"github.com/arpabet/sprint/pkg/app"
	"github.com/arpabet/sprint/pkg/client"
	"github.com/arpabet/sprint/pkg/util"
)

type stopCommand struct {
}

func (t *stopCommand) Desc() string {
	return "stop server"
}

func (t *stopCommand) Run(args []string) error {
	app.ParseFlags(args)

	if status, err := client.RequestStop(); err == nil {
		println(status)
		return nil
	} else {
		return util.KillServer()
	}
}
