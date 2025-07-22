const go = new Go();

const fetchMain = WebAssembly.instantiateStreaming(
  fetch("main.wasm?v=4"),
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

    go.run(wasm.instance)
  });
});
