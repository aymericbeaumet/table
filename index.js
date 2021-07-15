import SimplePeer from "simple-peer";

const WEBSOCKET_BASE_URL =
  process.env.WEBSOCKET_BASE_URL || "ws://localhost:8080";

if (window.location.pathname !== "/") {
  navigator.mediaDevices
    .getUserMedia({
      audio: true,
      video: {
        width: { ideal: 360, max: 360 },
        height: { ideal: 360, max: 360 },
        facingMode: { ideal: "user" },
        frameRate: { ideal: 30, max: 30 },
      },
    })
    .then((stream) => {
      newVideoDOM("me", stream);
      initRoom(window.location.pathname.substring(1), stream);
    })
    .catch((error) => {
      throw error;
    });
}

function initRoom(name, stream) {
  const url = `${WEBSOCKET_BASE_URL}/rooms/${encodeURIComponent(name)}`;
  const ws = new WebSocket(url);

  const peers = {};

  ws.onmessage = (event) => {
    const { PEER_STATUS, PEER_SIGNAL } = JSON.parse(event.data);

    // Update the peer connection following a status update
    if (PEER_STATUS) {
      const { peerID, status } = PEER_STATUS;
      switch (status) {
        case "online": {
          const peer =
            peers[peerID] || newPeer(ws, peerID, { initiator: true, stream });
          peers[peerID] = peer;
          break;
        }
        case "offline": {
          const peer = peers[peerID];
          if (peer) {
            peer.destroy();
            delete peers[peerID];
          }
          break;
        }
      }
    }

    // Accept a signal (offer/answer) from another user
    if (PEER_SIGNAL) {
      const { peerID, signal } = PEER_SIGNAL;
      const peer =
        peers[peerID] || newPeer(ws, peerID, { initiator: false, stream });
      peer.signal(signal);
      peers[peerID] = peer;
    }
  };
}

function newPeer(ws, peerID, { initiator, stream }) {
  console.log(`[${peerID}] New peer`);
  const peer = new SimplePeer({
    trickle: false,
    initiator,
    stream,
  });

  peer.on("error", (err) => {
    console.error(`[${peerID}] error:`, err);
  });

  peer.once("signal", (data) => {
    ws.send(
      JSON.stringify({
        SIGNAL: { peerID, signal: data },
      })
    );
  });

  peer.on("stream", (stream) => {
    console.log(`[${peerID}] Incoming stream`);

    const cancel = newVideoDOM(peerID, stream);
    peer.once("close", cancel);
  });

  return peer;
}

function newVideoDOM(name, stream) {
  const container = document.createElement("div");
  container.classList.add("container");
  document.body.appendChild(container);

  const video = document.createElement("video");
  container.appendChild(video);

  const legend = document.createElement("legend");
  legend.innerText = name;
  container.appendChild(legend);

  if ("srcObject" in video) {
    video.srcObject = stream;
  } else {
    video.src = window.URL.createObjectURL(stream); // for older browsers
  }

  video.play();

  return () => {
    document.body.removeChild(container);
  };
}
