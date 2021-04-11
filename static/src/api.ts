const addTorrent = async (uri: string): Promise<unknown> =>
  await fetch("/api/torrents/", {
    method: "POST",
    body: JSON.stringify({ uri })
  }).then((res) => res.json());

const listTorrents = async (): Promise<unknown> =>
  await fetch("/api/torrents/").then((res) => res.json());

export { addTorrent, listTorrents };
