package gui

import "github.com/mvrilo/torresmo"

type GUI interface {
	Register(t *torresmo.Torresmo)
	Start()
	Stop()
}

var App GUI
