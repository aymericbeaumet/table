import ReconnectingWebSocket from "reconnecting-websocket";
import SimplePeer from "simple-peer";
import getVideoId from "get-video-id";

const WEBSOCKET_BASE_URL =
  process.env.WEBSOCKET_BASE_URL || "ws://localhost:8080";

navigator.mediaDevices.enumerateDevices().then((devices) => {
  for (const d of devices) {
    const select = document.getElementById(d.kind);
    const option = document.createElement("option");
    option.innerText = d.label;
    select.appendChild(option);
  }
});

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
      newVideoDOM(stream, { muted: true });
      initRoom(window.location.pathname.substring(1), stream);
    })
    .catch((error) => {
      throw error;
    });
}

function initRoom(name, stream) {
  const url = `${WEBSOCKET_BASE_URL}/rooms/${encodeURIComponent(name)}`;
  const ws = new ReconnectingWebSocket(url);

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

    const cancel = newVideoDOM(stream);
    peer.once("close", cancel);
  });

  return peer;
}

function newVideoDOM(stream, opts = {}) {
  const container = document.createElement("div");
  container.classList.add("container");

  const video = document.createElement("video");
  video.muted = !!opts.muted;
  container.appendChild(video);

  if ("srcObject" in video) {
    video.srcObject = stream;
  } else {
    video.src = window.URL.createObjectURL(stream); // for older browsers
  }

  video.play();

  document.body.appendChild(container);
  return () => document.body.removeChild(container);
}

function newTextDOM(text) {
  const container = document.createElement("div");
  container.classList.add("container");

  const p = document.createElement("p");
  container.appendChild(p);

  p.innerText = text;

  document.body.appendChild(container);
  return () => document.body.removeChild(container);
}

function newIframe(src) {
  const container = document.createElement("div");
  container.classList.add("container");

  const iframe = document.createElement("iframe");
  iframe.width = "640";
  iframe.height = "360";
  iframe.src = src;
  iframe.frameborder = "0";
  iframe.allow = "autoplay; encrypted-media";
  iframe.allowfullscreen = true;
  container.appendChild(iframe);

  document.body.appendChild(container);
  return () => document.body.removeChild(container);
}

/* Modal */

const modalElement = document.getElementById("modal");
const formElement = modalElement.querySelector("form");
const inputTextElement = formElement.querySelector('input[type="text"]');

formElement.addEventListener("submit", function (event) {
  event.preventDefault();

  newTextDOM(inputTextElement.value);

  const { service, id } = getVideoId(inputTextElement.value);
  if (service === "youtube") {
    newIframe(`https://www.youtube-nocookie.com/embed/${id}`);
  } else if (service === "vimeo") {
    newIframe(`https://player.vimeo.com/video/${id}`);
  }

  modalElement.classList.remove("visible");
});

/* Context menu */

const contextMenuElement = document.getElementById("context-menu");

document.body.addEventListener(
  "contextmenu",
  function (event) {
    event.preventDefault();

    let x = event.clientX;
    const clientXDelta =
      x + contextMenuElement.offsetWidth - document.body.offsetWidth;
    if (clientXDelta > 0) {
      x -= clientXDelta;
    }

    let y = event.clientY;
    const clientYDelta =
      y + contextMenuElement.offsetHeight - document.body.offsetHeight;
    if (clientYDelta > 0) {
      y -= clientYDelta;
    }

    contextMenuElement.style.top = `${y}px`;
    contextMenuElement.style.left = `${x}px`;
    contextMenuElement.classList.add("visible");
  },
  false
);

document.body.addEventListener("mousedown", function (event) {
  if (contextMenuElement.classList.contains("visible")) {
    event.preventDefault();
  }

  if (event.target.offsetParent === contextMenuElement) {
    switch (event.target.getAttribute("data-action")) {
      case "pinnote":
        modal.classList.add("visible");
        inputTextElement.value = "";
        inputTextElement.focus();
        break;
      case "copyroom":
        navigator.clipboard.writeText(window.location.toString());
        break;
    }
  }

  contextMenuElement.classList.remove("visible");
});
