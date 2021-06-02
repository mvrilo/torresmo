// +build darwin
// +build !arm64

package gui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	torresm "github.com/mvrilo/torresmo"
	"github.com/mvrilo/torresmo/log"

	"github.com/progrium/macdriver/cocoa"
	"github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/objc"
	"github.com/progrium/macdriver/webkit"
)

func init() {
	App = &GuiMac{}
}

type GuiMac struct {
	sync.Mutex
	t   *torresm.Torresmo
	app cocoa.NSApplication
}

var _ GUI = (*GuiMac)(nil)

func (g *GuiMac) Register(torresm *torresm.Torresmo) {
	g.t = torresm
	log := g.t.Logger

	config := webkit.WKWebViewConfiguration_New()
	config.Preferences().SetValueForKey(core.True, core.String("developerExtrasEnabled"))

	url := core.URL(fmt.Sprintf("http://%s", g.t.HTTPServer.Addr))
	req := core.NSURLRequest_Init(url)

	g.app = cocoa.NSApp_WithDidLaunch(func(n objc.Object) {
		g.setup(n, req, config)
	})

	log.Info("Darwin GUI Started")
}

func (g *GuiMac) setup(n objc.Object, req core.NSURLRequest, config webkit.WKWebViewConfiguration) {
	obj := cocoa.NSStatusBar_System().StatusItemWithLength(cocoa.NSVariableStatusItemLength)
	obj.Retain()
	obj.Button().SetTitle("Torresmo")

	wv := webkit.WKWebView_Init(cocoa.NSScreen_Main().Frame(), config)
	wv.Retain()
	// wv.SetOpaque(false)
	// wv.SetBackgroundColor(cocoa.NSColor_Clear())
	// wv.SetValueForKey(core.False, core.String("drawsBackground"))
	wv.LoadRequest(req)

	win := cocoa.NSWindow_Init(
		cocoa.NSScreen_Main().Frame(),
		cocoa.NSClosableWindowMask|cocoa.NSBorderlessWindowMask,
		cocoa.NSBackingStoreBuffered,
		false,
	)
	win.SetContentView(wv)
	// win.SetBackgroundColor(cocoa.NSColor_Clear())
	// win.SetOpaque(false)
	// win.SetTitleVisibility(cocoa.NSWindowTitleHidden)
	win.SetTitlebarAppearsTransparent(true)
	win.SetIgnoresMouseEvents(true)
	win.SetLevel(cocoa.NSMainMenuWindowLevel + 2)
	win.MakeKeyAndOrderFront(win)
	win.SetCollectionBehavior(cocoa.NSWindowCollectionBehaviorCanJoinAllSpaces)
	win.Send("setHasShadow:", false)

	openWindow := cocoa.NSMenuItem_New()
	openWindow.Retain()
	openWindow.SetTitle("Open")
	openWindow.SetAction(objc.Sel("open:"))

	cocoa.DefaultDelegateClass.AddMethod("open:", func(_ objc.Object) {
		if win.IgnoresMouseEvents() {
			win.SetLevel(cocoa.NSMainMenuWindowLevel - 1)
			openWindow.SetState(1)
		} else {
			win.SetLevel(cocoa.NSMainMenuWindowLevel + 2)
			openWindow.SetState(0)
		}
	})

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
	menu.AddItem(openWindow)
	menu.AddItem(itemTorrents)
	menu.AddItem(itemQuit)
	obj.SetMenu(menu)
}

func (g *GuiMac) Start() {
	g.app.Run()
}

func (g *GuiMac) Stop() {
	g.app.Terminate()
}
