// +build darwin
// +build !arm64

package gui

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	torresm "github.com/mvrilo/torresmo"
	"github.com/mvrilo/torresmo/log"
	"github.com/progrium/macdriver/cocoa"
	"github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/objc"
)

func init() {
	App = &guiMac{}
}

type guiMac struct {
	sync.Mutex
	t   *torresm.Torresmo
	app cocoa.NSApplication
}

type NSUserNotification struct {
	objc.Object
}

var NSUserNotification_ = objc.Get("NSUserNotification")

type NSUserNotificationCenter struct {
	objc.Object
}

var NSUserNotificationCenter_ = objc.Get("NSUserNotificationCenter")

var _ GUI = (*guiMac)(nil)

func (g *guiMac) Register(torresm *torresm.Torresmo) {
	g.t = torresm
	log := g.t.Logger
	g.app = cocoa.NSApp_WithDidLaunch(g.setup)

	nsbundle := cocoa.NSBundle_Main().Class()
	nsbundle.AddMethod("__bundleIdentifier", func(_ objc.Object) objc.Object {
		return core.String("co.murilo.torresmo")
	})
	nsbundle.Swizzle("bundleIdentifier", "__bundleIdentifier")

	log.Info("Darwin GUI Started")
}

func notifyCompleted(value string) {
	notification := NSUserNotification{NSUserNotification_.Alloc().Init()}
	notification.Set("title:", core.String("Torrent Downloaded"))
	notification.Set("informativeText:", core.String(value))

	center := NSUserNotificationCenter{NSUserNotificationCenter_.Send("defaultUserNotificationCenter")}
	center.Send("deliverNotification:", notification)
	notification.Release()
}

func (g *guiMac) setup(n objc.Object) {
	obj := cocoa.NSStatusBar_System().StatusItemWithLength(cocoa.NSVariableStatusItemLength)
	obj.Retain()
	obj.Button().SetTitle("Torresmo")

	itemTorrents := cocoa.NSMenuItem_New()
	itemTorrents.SetAttributedTitle("no torrents yet")

	tcli := g.t.TorrentClient
	go func() {
		completed := make(map[string]struct{})
		for _, t := range tcli.Torrents() {
			if !t.Completed() {
				continue
			}
			completed[t.Name()] = struct{}{}
		}

		for {
			if torrents := tcli.Torrents(); len(torrents) > 0 {
				var lines []string
				for _, t := range torrents {
					if t.Name() == "" {
						continue
					}

					lines = append(lines, t.String())

					if _, ok := completed[t.Name()]; t.Completed() && !ok {
						completed[t.Name()] = struct{}{}
						notifyCompleted(t.Name())
					}
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
