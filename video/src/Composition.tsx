import { Audio } from "@remotion/media";
import { AbsoluteFill, Composition, interpolate, staticFile } from "remotion";
import { SsherCloudAnnouncement } from "./cloud-announcement";

export const SsherCloudLaunchFilm: React.FC = () => (
  <AbsoluteFill>
    <Audio
      src={staticFile("audio/mixkit-driving-ambition-32.mp3")}
      volume={(frame) =>
        interpolate(frame, [0, 36, 780, 899], [0, 0.46, 0.46, 0], {
          extrapolateLeft: "clamp",
          extrapolateRight: "clamp",
        })
      }
    />
    <Audio
      src={staticFile("audio/ssher-cloud-sfx.wav")}
      volume={(frame) =>
        interpolate(frame, [0, 12, 870, 899], [0, 0.78, 0.78, 0], {
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
