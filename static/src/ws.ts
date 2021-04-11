class WebsocketHandler {
  public ws: Websocket;
  private reconnectTimer: unknown;

  public onStatusChanged?: (status: boolean) => void
  public onMessageReceived?: (message: unknown[]) => void

  constructor(private uri: string) {
    this.connect();
  }

  connect(): void {
    this.ws = new WebSocket(this.uri);
    this.ws.onmessage = (e: MessageEvent) => this.onMessage(e);
    this.ws.onerror = (e: Event) => this.onError(e);
    this.ws.onclose = (e: Event) => this.onClose(e);
    this.ws.onopen = (e: Event) => this.onOpen(e);
  }

  onClose(e: Event): void {
    console.log("ws close", e);
    if (this.onStatusChanged) {
      this.onStatusChanged(false);
    }
    this.reconnectTimer = setTimeout(() => this.connect(), 4 * 1000); 
  }

  onOpen(e: Event): void {
    console.log("ws open", e);
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
    }

    if (this.onStatusChanged) {
      this.onStatusChanged(true);
    }
  }

  onMessage(e: MessageEvent): void {
    // console.log("ws message", e);

    let { data } = e;
    if (typeof data === 'string') {
      data = JSON.parse(data)
    }

    if (this.onMessageReceived) {
      this.onMessageReceived(data);
    }
  }

  onError(e: Event): void {
    console.error("ws error:", e);
  }
}

export default WebsocketHandler;
