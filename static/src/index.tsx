import { FC, render } from "preact";
import React from "preact/compat";
import { useRef, useState, useCallback, useEffect } from "preact/hooks";
import "terminal.css";

import WebsocketHandler from "./ws";
import { addTorrent, listTorrents } from "./api";

import Header from "./Header";
import TorrentList from "./TorrentList";

const WSURI = "ws://localhost:8000/api/events/";

const filterTorrents = (torrents = [], completed = false) =>
  Object.keys(torrents).filter((torrent) =>
    torrents[torrent].completed === completed).map((torrent) =>
      torrents[torrent]);

const Torresmo: FC<unknown> = () => {
  const ws = useRef(null);
  const [status, setStatus] = useState(false);
  const [torrents, setTorrents] = useState({});
  const [onlineCount, setOnlineCount] = useState(0);

  const onlineCallback = useCallback((count) => setOnlineCount(count));
  const torrentsCallback = useCallback((torrent) => {
    let speed = 0;
    const { name, bytesCompleted } = torrent;
    if (bytesCompleted && torrents[name]) {
      speed = bytesCompleted - torrents[name].bytesCompleted;
    }

    setTorrents({ ...torrents, [name]: { ...torrent, speed } });
  });

  const onMessageCallback = (topic, data) => {
    // console.log("received message from topic:", topic, data);
    setStatus(true);

    if (topic === "online") {
      onlineCallback(data);
      return;
    }

    torrentsCallback(data);
  };

  useEffect(() => {
    ws.conn = new WebsocketHandler(WSURI);
    ws.conn.onStatusChanged = (s: boolean) => setStatus(s);

    document.body.addEventListener("paste", (e: ClipboardEvent) => {
      e.preventDefault();
      const data = e.clipboardData.getData("text");
      if (data.indexOf("magnet:") === 0) {
        console.log("magnet detected, adding it:", data);
        addTorrent(data);
      }
    });

    // TODO: support dropping torrent files
    document.body.addEventListener("drop", (e) => {
      e.preventDefault();
      console.log("drop", e);
      const { items } = e.dataTransfer;

      for (let i = 0; i < items.length; i++) {
        const item = items[i];
        if (item.kind !== "file") {
          continue;
        }

        // const file = item.getAsFile();
      }
    });

    document.body.addEventListener("dragover", (e) => e.preventDefault());

    (async () => {
      const ts = { ...torrents };
      const res = await listTorrents();
      if (res && Object.keys(res).length > 0) {
        res.forEach((t) => { ts[t.name] = { ...t }; });
        setTorrents(ts);
      }
      main.style.opacity = 1;
    })();

    return () => {
      ws.conn.close();
    };
  }, []);

  useEffect(() => {
    ws.conn.onMessageReceived = ({ topic, data }) => {
      onMessageCallback(topic, data);
    };
  }, [onMessageCallback]);

  return (
    <div>
      <Header connected={status} onlineCount={onlineCount} />
      <p style={{ color: "#888" }}>paste a magnet uri to start downloading</p>

      <TorrentList label="Downloading" torrents={filterTorrents(torrents, false)} />
      <TorrentList label="Completed" torrents={filterTorrents(torrents, true)} />
    </div>
  );
};

render(<Torresmo />, document.getElementById("main"));
