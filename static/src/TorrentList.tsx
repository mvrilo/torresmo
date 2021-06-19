import { FC } from "preact";
import React from "preact/compat";

import TorrentEntry from "./TorrentEntry";

type TorrentListProps = {
  torrents: any;
  label: string;
};

const TorrentList: FC<unknown> = ({ torrents, label }: TorrentListProps) => {
  const count = Object.keys(torrents).length;
  const list = [<h2 style={{ paddingTop: "20px", color: "#ccc" }}>{label} {count > 0 ? `(${count})` : ""}</h2>];

  if (count === 0) {
    list.push(<p>None yet</p>);
    return list;
  }

  for (const i in torrents) {
    const torrent = torrents[i];

    list.push(<TorrentEntry torrent={torrent} />);
  }

  return list;
};


export default TorrentList;
