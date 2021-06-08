// +build darwin
// +build !arm64

package gui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/progrium/macdriver/cocoa"
	"github.com/progrium/macdriver/core"
	"github.com/progrium/macdriver/objc"
	"github.com/progrium/macdriver/webkit"
	"github.com/skratchdot/open-golang/open"

	torresm "github.com/mvrilo/torresmo"
	"github.com/mvrilo/torresmo/log"
	"github.com/mvrilo/torresmo/stream"
)

func init() {
	App = &GuiMac{}
}

type GuiMac struct {
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

var _ GUI = (*GuiMac)(nil)

func (g *GuiMac) Register(torresm *torresm.Torresmo) {
	cocoa.TerminateAfterWindowsClose = false

	g.t = torresm
	log := g.t.Logger

	config := webkit.WKWebViewConfiguration_New()
	config.Preferences().SetValueForKey(core.True, core.String("developerExtrasEnabled"))

	addr := g.t.HTTPServer.Addr
	if addr[0] == ':' {
		addr = "127.0.0.1" + addr[0:]
	}

	url := core.URL(fmt.Sprintf("http://%s", addr))
	req := core.NSURLRequest_Init(url)
	g.app = cocoa.NSApp_WithDidLaunch(g.setup(req, config, addr))

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

func (g *GuiMac) newWebViewWindow(n objc.Object, frame core.NSRect, req core.NSURLRequest, config webkit.WKWebViewConfiguration) (cocoa.NSWindow, webkit.WKWebView) {
	wv := webkit.WKWebView_Init(frame, config)
	wv.Retain()
	wv.SetOpaque(false)
	wv.SetValueForKey(core.False, core.String("drawsBackground"))
	wv.LoadRequest(req)

	win := cocoa.NSWindow_Init(
		frame,
		cocoa.NSClosableWindowMask|cocoa.NSBorderlessWindowMask|cocoa.NSResizableWindowMask,
		cocoa.NSBackingStoreBuffered,
		false,
	)
	win.Retain()
	// win.Center()
	win.SetContentView(wv)
	win.SetOpaque(false)
	win.SetTitleVisibility(cocoa.NSWindowTitleHidden)
	win.SetTitlebarAppearsTransparent(true)
	win.SetLevel(cocoa.NSMainMenuWindowLevel + 2)
	win.MakeKeyAndOrderFront(win)
	win.SetCollectionBehavior(cocoa.NSWindowCollectionBehaviorCanJoinAllSpaces)
	win.Send("setHasShadow:", false)
	win.OrderOut(win)

	return win, wv
}

func wsWatch(ctx context.Context, addr string) (res chan stream.Response, err error) {
	uri := fmt.Sprintf("ws://%s/api/events/", addr)

	conn, _, _, err := ws.Dial(ctx, uri)
	if err != nil {
		return nil, err
	}

	res = make(chan stream.Response)
	go func() {
		for {
			msg, op, err := wsutil.ReadServerData(conn)
			if err != nil && err == io.EOF {
				continue
			}

			if op == ws.OpContinuation {
				continue
			}

			var payload stream.Response
			if err = json.Unmarshal(msg, &payload); err != nil {
				continue
			}

			res <- payload
		}
	}()

	return
}

func (g *GuiMac) setup(req core.NSURLRequest, config webkit.WKWebViewConfiguration, addr string) func(n objc.Object) {
	return func(n objc.Object) {
		win, _ := g.newWebViewWindow(n, core.NSMakeRect(440, 320, 920, 500), req, config)

		obj := cocoa.NSStatusBar_System().StatusItemWithLength(cocoa.NSVariableStatusItemLength)
		obj.Retain()
		obj.Button().SetTitle("Torresmo")

		itemTorrents := cocoa.NSMenuItem_New()
		itemTorrents.SetEnabled(false)
		itemTorrents.SetAttributedTitle("no torrents yet")

		openBrowser := cocoa.NSMenuItem_New()
		openBrowser.Retain()
		openBrowser.SetTitle("Open Browser")
		openBrowser.SetAction(objc.Sel("browser:"))
		cocoa.DefaultDelegateClass.AddMethod("browser:", func(_ objc.Object) {
			if err := open.Run(fmt.Sprintf("http://%s", addr)); err != nil {
				log.Error("error opening browser: ", err)
			}
		})

		showWindow := cocoa.NSMenuItem_New()
		showWindow.SetTitle("Toggle Window")
		showWindow.SetState(1)
		showWindow.SetAction(objc.Sel("visible:"))
		cocoa.DefaultDelegateClass.AddMethod("visible:", func(_ objc.Object) {
			if win.IsVisible() {
				showWindow.SetState(0)
				win.OrderOut(win)
			} else {
				showWindow.SetState(1)
				win.OrderFront(win)
			}
		})

		itemQuit := cocoa.NSMenuItem_New()
		itemQuit.Retain()
		itemQuit.SetTitle("Quit")
		itemQuit.SetAction(objc.Sel("done:"))
		cocoa.DefaultDelegateClass.AddMethod("done:", func(_ objc.Object) {
			log.Info("Shutting down Torresmo")
			if err := g.t.Shutdown(context.Background(), 10*time.Second); err != nil {
				log.Error(err)
			}
			cocoa.NSApp().Terminate()
		})

		menu := cocoa.NSMenu_New()
		menu.SetAutoenablesItems(false)
		menu.AddItem(showWindow)
		menu.AddItem(openBrowser)
		menu.AddItem(cocoa.NSMenuItem_Separator())
		menu.AddItem(itemTorrents)
		menu.AddItem(cocoa.NSMenuItem_Separator())
		menu.AddItem(itemQuit)
		obj.SetMenu(menu)

		tcli := g.t.TorrentClient
		go func() {
			var lastPaste string

			for {
				gp := cocoa.NSPasteboard_GeneralPasteboard()
				paste := gp.StringForType(cocoa.NSPasteboardTypeString)
				if paste != lastPaste && strings.Contains(paste, "magnet:") {
					if _, err := tcli.AddURI(paste); err != nil {
						log.Error(err)
					} else {
						log.Info("Found magnet uri in clipboard")
					}
					lastPaste = paste
				}

				<-time.After(1 * time.Second)
			}
		}()

		torrents := tcli.Torrents()
		downloading := make(map[string]interface{})
		var completed int

		for _, t := range torrents {
			if t.Completed() {
				completed++
				continue
			}
			downloading[t.Name()] = nil
		}

		events, err := wsWatch(context.Background(), addr)
		if err != nil {
			log.Error(err)
		}

		go func() {
			for event := range events {
				data, ok := event.Data.(map[string]interface{})
				if !ok {
					continue
				}

				name := data["name"].(string)

				switch event.Topic {
				case stream.TopicDownloading.String():
					if _, ok := downloading[name]; !ok {
						downloading[name] = nil
					}
				case stream.TopicCompleted.String():
					notifyCompleted(name)
					delete(downloading, name)
					completed++
				default:
				}

				core.Dispatch(func() {
					lines := []string{
						fmt.Sprintf("Downloading: %d", len(downloading)-completed),
						fmt.Sprintf("Completed: %d", completed),
					}
					itemTorrents.SetAttributedTitle(strings.Join(lines, "\n"))
				})
			}
		}()
	}
}

func (g *GuiMac) Start() {
	// g.app.SetActivationPolicy(cocoa.NSApplicationActivationPolicyRegular)
	g.app.ActivateIgnoringOtherApps(true)
	g.app.Run()
}

func (g *GuiMac) Stop() {
	g.app.Terminate()
}
