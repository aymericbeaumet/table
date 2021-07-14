import SimplePeer from "simple-peer";

if (window.location.pathname !== "/") {
  initRoom(window.location.pathname.substring(1));
}

function initRoom(name) {
  const url = `ws://localhost:8080/rooms/${encodeURIComponent(name)}`;
  const ws = new WebSocket(url);

  const peers = {};

  ws.onmessage = (event) => {
    const { type, payload } = JSON.parse(event.data);

    if (type === "PEER") {
      const peer = newPeer({
        initiator: true,
        peerID: payload.peerID,
        ws,
      });
      peers[payload.peerID] = peer;
    }

    if (type === "HANDSHAKE") {
      const peer =
        peers[payload.peerID] ||
        newPeer({
          initiator: false,
          peerID: payload.peerID,
          ws,
        });
      peer.signal(payload.signal);
      peers[payload.peerID] = peer;
    }
  };
}

function newPeer({ initiator, peerID, ws }) {
  const peer = new SimplePeer({ initiator, trickle: false });

  peer.on("error", (err) => {
    console.error(err);
  });

  peer.once("signal", (data) => {
    ws.send(
      JSON.stringify({
        HANDSHAKE: { peerID, signal: data },
      })
    );
  });

  peer.on("connect", () => {
    console.log("CONNECT");
  });

  peer.on("data", (data) => {
    console.log("DATA", data);
  });

  return peer;
}
