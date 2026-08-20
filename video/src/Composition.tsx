import { Audio } from "@remotion/media";
import { AbsoluteFill, Composition, interpolate, staticFile } from "remotion";
import { SsherCloudAnnouncement } from "./cloud-announcement";

export const SsherCloudLaunchFilm: React.FC = () => (
  <AbsoluteFill>
    <Audio
      src={staticFile("audio/ssher-cloud-score.wav")}
      volume={(frame) =>
        interpolate(frame, [0, 20, 840, 899], [0, 0.92, 0.92, 0], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
        })
      }
    />
    <SsherCloudAnnouncement />
  </AbsoluteFill>
);

export const SsherCompositions: React.FC = () => (
  <Composition
    id="SsherCloudLaunch"
    component={SsherCloudLaunchFilm}
    durationInFrames={900}
    fps={30}
    width={1920}
    height={1080}
  />
);
