import SimplePeer from "simple-peer";

if (window.location.pathname !== "/") {
  initRoom(window.location.pathname.substring(1));
}

function initRoom(name) {
  const url = `ws://localhost:8080/rooms/${encodeURIComponent(name)}`;
  const ws = new WebSocket(url);

  const peers = {};

  ws.onmessage = (event) => {
    const { HANDSHAKE, PEER } = JSON.parse(event.data);

    if (HANDSHAKE) {
      const peer =
        peers[HANDSHAKE.peerID] ||
        newPeer({
          initiator: false,
          peerID: HANDSHAKE.peerID,
          ws,
        });
      peer.signal(HANDSHAKE.signal);
      peers[HANDSHAKE.peerID] = peer;
    }

    if (PEER) {
      const peer =
        peers[PEER.peerID] ||
        newPeer({
          initiator: true,
          peerID: PEER.peerID,
          ws,
        });
      peers[PEER.peerID] = peer;
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
    peer.send("Hello, World!");
  });

  peer.on("data", (data) => {
    console.log("DATA", data.toString());
  });

  return peer;
}
