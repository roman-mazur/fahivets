const go = new Go();

const fetchMain = WebAssembly.instantiateStreaming(
  fetch("main.wasm?v=2"),
  go.importObject
);

addEventListener("DOMContentLoaded", () => {
  const container = document.getElementById("mainApp");

  const canvas = container.getElementsByClassName("display")[0];
  canvas.width = window.innerWidth;
  canvas.height = window.innerHeight;

  const graphCtx = canvas.getContext("2d");

  console.debug("document loaded, start main code")
  fetchMain.then(wasm => {
    let sharedMemory;

    window.renderDisplay = (ptr, len, w, h) => {
      if (!sharedMemory || sharedMemory.byteLength === 0) {
        sharedMemory = new Uint8ClampedArray(wasm.instance.exports.mem.buffer);
      }
      const data = sharedMemory.subarray(ptr, ptr + len);
      graphCtx.putImageData(new ImageData(data, w, h), 0, 0);
    };

    registerEventHandlers();
    go.run(wasm.instance)

    window.fahivets = makeSDK();
  });
});

const _kbEventsBuffer = [];
window.kbEventsBuffer = _kbEventsBuffer;

const pushKbEvent = data => {
  _kbEventsBuffer.push(data);
  console.debug(data);
};

function registerEventHandlers() {
  console.debug("registering keyboard events");

  const target = document.documentElement;
  target.addEventListener("keydown", event =>
    pushKbEvent({code: event.code, down: true}));
  target.addEventListener("keyup", event =>
    pushKbEvent({code: event.code, down: false}));
}

function makeSDK() {
  const sdk = {
    test: () => {
      console.log("running a test...");
      sdk.kbSequence(['F7', 'AltRight']);
    },
    kbSequence: (seq) => {
      const process = seq => {
        if (seq.length === 0) {
          return;
        }
        const key = seq.shift();
        pushKbEvent({code: key, down: true});
        setTimeout(() => {
          pushKbEvent({code: key, down: false});
          setTimeout(() => process(seq), 100);
        }, 300);
      };
      process(seq);
    },
  };
  return sdk;
}