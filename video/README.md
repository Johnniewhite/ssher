# ssher Cloud announcement film

The Remotion source for the 30-second ssher Cloud announcement film. The
composition is 1920×1080 at 30fps and tells the launch story through a familiar
mobile group conversation before opening into the browser terminal experience.

The soundtrack combines the licensed cinematic track “Driving Ambition” by
Ahjay Stelino with original, deterministic interface sound design generated
from code as 44.1kHz stereo PCM. The custom layer contains distinct
incoming/outgoing message tones, keyboard detail, transitions, and a
service-restart confirmation. See `public/audio/THIRD_PARTY_ASSETS.md` for the
music source, license, and integrity hash.

## Storyboard

1. An urgent production message arrives in the team chat.
2. A teammate is away from their laptop.
3. ssher Cloud arrives in the conversation.
4. The browser terminal connects to production and restores the service.
5. The team leaves the terminal running and can resume it later.
6. The product promise resolves: “Your servers. Right where you left them.”

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
