import { render } from "preact";
import React from "preact/compat";
import { useState, useEffect } from "preact/hooks";

import styles from "./styles";
import WebsocketHandler from "./ws";
// import parseTorrent from "parse-torrent";
import { addTorrent, listTorrents } from "./api";

import "terminal.css";

const sizes = ['b', 'kb', 'mb', 'gb', 'tb', 'pb', 'eb', 'zb', 'yb'];
const WSURI = "ws://localhost:8000/api/events/?rooms=downloading,completed,started";

const ws = new WebsocketHandler(WSURI);

const humanBytes = (bytes: number) => {
   if (bytes === 0) {
     return '0b';
   }

  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  if (isNaN(i)) {
    return '0b';
  }

  return parseFloat((bytes / Math.pow(1024, i)).toFixed(2)) + sizes[i];
};

const List = ({ torrents, header }) => {
  const count = Object.keys(torrents).length;
  const list = [<h2>{header} {count > 0 ? `(${count})` : ""}</h2>];

  if (count === 0) {
    list.push(<p>None yet</p>);
    return list;
  }

  for (const i in torrents) {
    const torrent = torrents[i];
    const { name, speed, infoHash, files, totalLength, bytesCompleted } = torrent;
    const percentage = parseFloat(bytesCompleted / totalLength * 100.0).toFixed(2);
    const downloaded = humanBytes(bytesCompleted);
    const total = humanBytes(totalLength);

    list.push(
      <div style={styles.listItem}>
        <div style={{ ...styles.listItemBackground, width: `${percentage}%` }}></div>
        <p style={styles.listItemLeft}>
          {name}<br/>
          {infoHash} (files: {files.length})
        </p>
        <p style={styles.listItemRight}>
          {downloaded}/{total}<br/>
          {percentage}% {humanBytes(speed)}/s
        </p>
      </div>
    );
  }

  return list;
};

const Header = ({ connected }) => {
  const color = connected ? "lightgreen" : "red";
  const status = connected ? "connected" : "disconnected";
  return (
    <nav>
      <h1>
        Torresmo 
        <small style={{color}}> {status} </small>
      </h1>
      <p style={styles.info}>paste a magnet uri to start downloading</p>
    </nav>
  );
};

const filterTorrents = (torrents = [], completed = false) =>
  Object.keys(torrents).filter((torrent) => torrents[torrent].completed === completed).map((torrent) => torrents[torrent]);

const Torresmo = () => {
  const [ready, setReady] = useState(false);
  const [status, setStatus] = useState(false);
  const [torrents, setTorrents] = useState({});

  useEffect(() => {
    if (ready) {
      return;
    }

    document.body.addEventListener("paste", (e: ClipboardEvent) => {
      e.preventDefault();
      // console.log("paste", e);
      const data = e.clipboardData.getData("text");
      if (data.indexOf("magnet:") === 0) {
        console.log("magnet detected, adding it:", data);
        addTorrent(data);
      }
    });

    document.body.addEventListener("dragover", (e) => e.preventDefault());
    document.body.addEventListener("drop", (e) => {
      e.preventDefault();
      console.log("drop", e);
      const { items } = e.dataTransfer;

      for (let i = 0; i < items.length; i++) {
        const item = items[i];
        if (item.kind !== "file") {
          continue;
        }

        const file = item.getAsFile();
        // const torrent = parseTorrent(file);
        // console.log(file.name);
      }
    });

    (async () => {
      main.style.opacity = 1;
      const ts = { ...torrents };
      const res = await listTorrents();
      if (res && Object.keys(res).length > 0) {
        res.forEach((t) => { ts[t.name] = { ...t }; });
        setTorrents(ts);
      }
      setReady(true);
    })();
  }, [ready]);

  useEffect(() => {
    ws.onStatusChanged = (s: boolean) => setStatus(s);
    ws.onMessageReceived = ({ room, torrent }) => {
      console.log("received message from room:", room);

      let speed = 0;
      const { name, bytesCompleted } = torrent;
      const newTorrents = { ...torrents };
      if (bytesCompleted && newTorrents[name] && newTorrents[name].bytesCompleted) {
        speed = bytesCompleted - newTorrents[name].bytesCompleted;
      }
      newTorrents[name] = { ...torrent, speed };
      setTorrents({ ...newTorrents });
      setStatus(true);
    };
  }, [torrents, status]);

  return (
    <div>
      <Header connected={status} />
      <List header="Downloading" torrents={filterTorrents(torrents, false)} />
      <List header="Completed" torrents={filterTorrents(torrents, true)} />
    </div>
  );
};

render(<Torresmo />, document.getElementById("main"));
