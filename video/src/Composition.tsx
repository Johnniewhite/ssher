import { Audio } from "@remotion/media";
import { AbsoluteFill, Composition, interpolate, staticFile } from "remotion";
import { SsherCloudLaunchV2 } from "./cloud-launch-v2";

export const SsherCloudLaunchFilm: React.FC = () => (
  <AbsoluteFill>
    <Audio
      src={staticFile("audio/mixkit-dreaming-big-31.mp3")}
      trimBefore={58 * 30}
      volume={(frame) =>
        interpolate(frame, [0, 30, 1260, 1349], [0, 0.62, 0.62, 0], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
        })
      }
    />
    <Audio
      src={staticFile("audio/ssher-cloud-sfx.wav")}
      volume={(frame) =>
        interpolate(frame, [0, 12, 1305, 1349], [0, 0.74, 0.74, 0], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
        })
      }
    />
    <SsherCloudLaunchV2 />
  </AbsoluteFill>
);

export const SsherCompositions: React.FC = () => (
  <Composition
    id="SsherCloudLaunch"
    component={SsherCloudLaunchFilm}
    durationInFrames={1350}
    fps={30}
    width={1920}
    height={1080}
  />
);
