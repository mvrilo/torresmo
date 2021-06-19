import { FC } from "preact";
import React from "preact/compat";

type HeaderProps = {
  connected: boolean;
};

const Header: FC<unknown> = ({ connected }: HeaderProps) => {
  return (
    <header style={{ paddingTop: 20 }}>
      <h1 style={{ fontSize: "large", color: "#fff" }}>
        Torresmo
        <small style={{ fontSize: "x-small", color: "lightgreen" }}>{ connected ? " connected" : "" }</small>
      </h1>
    </header>
  );
};

export default Header;
