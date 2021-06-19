import { FC } from "preact";
import React from "preact/compat";

import ProgressBar from "./ProgressBar";

const sizes = ['b', 'kb', 'mb', 'gb', 'tb', 'pb', 'eb', 'zb', 'yb'];
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

type TorrentEntryProps = {
  torrent: any;
};

const TorrentEntry: FC<unknown> = ({ torrent }: TorrentEntryProps) => {
  const { name, speed, infoHash, files, totalLength, bytesCompleted } = torrent;
  const percentage = parseFloat(bytesCompleted / totalLength * 100.0).toFixed(2);
  const downloaded = humanBytes(bytesCompleted);
  const total = humanBytes(totalLength);

  return (
    <div style={{ marginBottom: "20px" }}>
      <span>
        {name}<br/>
        {infoHash}<br/>
        {files.length} files {percentage}% {downloaded}/{total} {humanBytes(speed)}/s
      </span>
      <ProgressBar progress={percentage} />
    </div>
  );
};

export default TorrentEntry;
