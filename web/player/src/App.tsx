import HLS from "hls.js";
import "./app.css";
import { createSignal } from "solid-js";

const STREAM_SOURCE = "http://localhost:8080/stream";

const App = () => {
  let videoRef: HTMLVideoElement | undefined;
  let hls: HLS | undefined;

  const [isPlay, setIsPlay] = createSignal(false);

  const togglePlayback = () => {
    if (!isPlay() && HLS.isSupported()) {
      hls = new HLS();
      hls.loadSource(STREAM_SOURCE);
      hls.attachMedia(videoRef as unknown as HTMLMediaElement);
    }

    if (isPlay()) {
      setIsPlay(false);
      videoRef?.pause();
      hls?.destroy();
    } else {
      setIsPlay(true);
      videoRef?.play();
    }
  };

  return (
    <>
      <video id="video" ref={videoRef}></video>
      <div class="container">
        <div
          class={`box center${isPlay() ? " pause" : ""}`}
          onClick={togglePlayback}
        ></div>
      </div>
    </>
  );
};

export default App;
