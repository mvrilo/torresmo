import { FC } from "preact";
import React from "preact/compat";

type ProgressProps = {
  progress: number;
};

const ProgressBar: FC<unkown> = ({ progress }: ProgressProps) => {
  return (
    <div className="progress-bar">
      <div className={"progress-bar-filled"} style={{ width: `${progress}%` }}></div>
    </div>
  );
};

export default ProgressBar;
