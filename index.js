import SimplePeer from "simple-peer";

if (window.location.pathname !== "/") {
  initRoom(window.location.pathname.substring(1));
}

function initRoom(name) {
  const url = `ws://localhost:8080/rooms/${encodeURIComponent(name)}`;
  const ws = new WebSocket(url);

  const peers = {};

  ws.onmessage = (event) => {
    const { EXISTING_PEER, SIGNAL_RECEIVED } = JSON.parse(event.data);

    // Bootstrap a connection to an existing peer
    if (EXISTING_PEER) {
      const peer = newPeer({
        initiator: true,
        peerID: EXISTING_PEER.peerID,
        ws,
      });
      peers[EXISTING_PEER.peerID] = peer;
    }

    // Accept a signal (offer/answer) from another user
    if (SIGNAL_RECEIVED) {
      const peer =
        peers[SIGNAL_RECEIVED.peerID] ||
        newPeer({
          initiator: false,
          peerID: SIGNAL_RECEIVED.peerID,
          ws,
        });
      peer.signal(SIGNAL_RECEIVED.signal);
      peers[SIGNAL_RECEIVED.peerID] = peer;
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
        SIGNAL: { peerID, signal: data },
      })
    );
  });

  peer.on("connect", () => {
    console.log("CONNECT");
    peer.write("Hello, World!");
  });

  peer.on("data", (data) => {
    console.log("DATA", data.toString());
  });

  return peer;
}
