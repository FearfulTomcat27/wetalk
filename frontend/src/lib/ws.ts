type EventHandler = (payload: unknown) => void;

interface WsMessage {
  type: string;
  [key: string]: unknown;
}

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

function deriveWsUrl(apiUrl: string): string {
  return apiUrl.replace(/^http/, "ws") + "/ws";
}

const WS_URL =
  typeof window !== "undefined"
    ? process.env.NEXT_PUBLIC_WS_URL || deriveWsUrl(API_URL)
    : "";

const RECONNECT_BASE_MS = 1000;
const RECONNECT_MAX_MS = 30000;
const HEARTBEAT_INTERVAL_MS = 25000;

class WSClient {
  private ws: WebSocket | null = null;
  private handlers: Map<string, Set<EventHandler>> = new Map();
  private retryCount = 0;
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private _connected = false;
  private intentionalClose = false;

  get connected(): boolean {
    return this._connected;
  }

  connect(): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      return;
    }

    const token =
      typeof window !== "undefined" ? localStorage.getItem("token") : null;
    if (!token) {
      return;
    }

    this.intentionalClose = false;
    this.ws = new WebSocket(WS_URL);

    this.ws.onopen = () => {
      this.retryCount = 0;
      this.send("auth", { token });
      this.startHeartbeat();
    };

    this.ws.onmessage = (event) => {
      try {
        const msg: WsMessage = JSON.parse(event.data);
        this.dispatch(msg.type, msg);
      } catch {
        // ignore malformed messages
      }
    };

    this.ws.onclose = () => {
      this._connected = false;
      this.stopHeartbeat();
      this.dispatch("disconnect", {});
      if (!this.intentionalClose) {
        this.scheduleReconnect();
      }
    };

    this.ws.onerror = () => {
      // onclose will fire after onerror, reconnect handled there
    };
  }

  disconnect(): void {
    this.intentionalClose = true;
    this.stopHeartbeat();
    this.clearReconnect();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this._connected = false;
  }

  send(type: string, payload: Record<string, unknown> = {}): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, ...payload }));
    }
  }

  on(type: string, handler: EventHandler): () => void {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler);
    return () => {
      this.handlers.get(type)?.delete(handler);
    };
  }

  private dispatch(type: string, payload: unknown): void {
    if (type === "auth.ok") {
      this._connected = true;
    }
    if (type === "auth.error") {
      this._connected = false;
      this.intentionalClose = true;
      this.disconnect();
      return;
    }
    const handlers = this.handlers.get(type);
    if (handlers) {
      for (const h of handlers) {
        h(payload);
      }
    }
  }

  private startHeartbeat(): void {
    this.stopHeartbeat();
    this.heartbeatTimer = setInterval(() => {
      this.send("ping");
    }, HEARTBEAT_INTERVAL_MS);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  private scheduleReconnect(): void {
    this.clearReconnect();
    const delay = Math.min(
      RECONNECT_BASE_MS * Math.pow(2, this.retryCount),
      RECONNECT_MAX_MS,
    );
    const jitter = Math.random() * delay * 0.3;
    this.reconnectTimer = setTimeout(() => {
      this.retryCount++;
      this.connect();
    }, delay + jitter);
  }

  private clearReconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }
}

export const wsClient = new WSClient();