// +build darwin

package gui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mvrilo/torresmo"
	"github.com/mvrilo/torresmo/log"
	"github.com/progrium/macdriver/cocoa"
	"github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/objc"
)

type guiMac struct {
	sync.Mutex
	t   *torresmo.Torresmo
	app cocoa.NSApplication
}

var _ GUI = &guiMac{}

func NewGUI(torresm *torresmo.Torresmo) GUI {
	log := torresm.Logger
	log.Info(fmt.Sprintf("Darwin GUI Started"))

	macapp := &guiMac{t: torresm}
	macapp.app = cocoa.NSApp_WithDidLaunch(macapp.setup)

	return macapp
}

func (g *guiMac) setup(n objc.Object) {
	obj := cocoa.NSStatusBar_System().StatusItemWithLength(cocoa.NSVariableStatusItemLength)
	obj.Retain()
	obj.Button().SetTitle("Torresmo")

	itemTorrents := cocoa.NSMenuItem_New()
	itemTorrents.SetAttributedTitle("no torrents yet")

	tcli := g.t.TorrentClient
	go func() {
		for {
			if torrents := tcli.Torrents(); len(torrents) > 0 {
				var lines []string
				for _, t := range torrents {
					lines = append(lines, t.String())
				}
				sort.Strings(lines)
				core.Dispatch(func() {
					itemTorrents.SetAttributedTitle(strings.Join(lines, "\n"))
				})
			}
			<-time.After(1 * time.Second)
		}
	}()

	// itemNew := cocoa.NSMenuItem_New()
	// itemNew.SetTitle("New torrent")
	// itemNew.SetAction(objc.Sel("newTorrent:"))

	itemQuit := cocoa.NSMenuItem_New()
	itemQuit.SetTitle("Quit")
	itemQuit.SetAction(objc.Sel("done:"))

	cocoa.DefaultDelegateClass.AddMethod("newTorrent:", func(_ objc.Object) {
		println("new torrent clicked")
	})

	cocoa.DefaultDelegateClass.AddMethod("done:", func(_ objc.Object) {
		log.Info("Shutting down Torresmo")
		if err := g.t.Shutdown(context.Background(), 10*time.Second); err != nil {
			log.Error(err)
		}
		cocoa.NSApp().Terminate()
	})

	menu := cocoa.NSMenu_New()
	menu.AddItem(itemTorrents)
	// menu.AddItem(itemNew)
	menu.AddItem(itemQuit)
	obj.SetMenu(menu)
}

func (g *guiMac) Start() {
	g.app.Run()
}

func (g *guiMac) Stop() {
	g.app.Terminate()
}
