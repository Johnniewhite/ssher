# ssher Cloud launch film

The Remotion source for the 45-second ssher Cloud launch film. The composition
is 1920×1080 at 30fps and presents the real product promise through cinematic
product footage: cross-device access, encrypted workspace sync, browser SSH,
AI-assisted incident recovery, persistent sessions, and team control.

The soundtrack uses the late orchestral build of the licensed piano-and-strings
film score “Dreaming Big” by Ahjay Stelino, layered with original, deterministic
interface sound design generated from code as 44.1kHz stereo PCM. See
`public/audio/THIRD_PARTY_ASSETS.md` for the source, license, and integrity hash.

## Storyboard

1. The old constraint breaks: your servers do not live at your desk.
2. One encrypted session appears across laptop, tablet, and phone.
3. The monochrome workspace brings server and team access into one view.
4. A browser terminal and AI operator investigate and recover a live incident.
5. The session detaches from one device and resumes on another.
6. Zero-knowledge encryption, team control, and activity history resolve into
   the launch promise: “Every server. Any screen. Exactly where you left it.”

## Develop

```sh
npm install
npm run audio
npm run dev
```

All visual motion is driven by Remotion frames, springs, and interpolation.
The production components intentionally avoid CSS transitions and animations.

## Render and validate

```sh
npm run lint
npm run render
npx remotion ffprobe out/ssher-cloud-launch.mp4
```

The H.264/AAC master is written to `out/ssher-cloud-launch.mp4`.

The ssher source and original sound-design code are MIT licensed. The music is
used under the Mixkit Stock Music Free License and is not relicensed under MIT.
Remotion itself has separate licensing requirements for some commercial teams;
consult Remotion's current license before commercial production use.
