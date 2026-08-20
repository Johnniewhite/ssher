import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const here = path.dirname(fileURLToPath(import.meta.url));
const outputDir = path.join(here, "..", "public", "audio");
fs.mkdirSync(outputDir, { recursive: true });

const sampleRate = 44100;
const duration = 30;
const length = sampleRate * duration;
const left = new Float32Array(length);
const right = new Float32Array(length);
let noiseState = 0x5f3759df;

const random = () => {
  noiseState = (1664525 * noiseState + 1013904223) >>> 0;
  return noiseState / 0xffffffff;
};

const add = (buffer, index, value) => {
  if (index >= 0 && index < buffer.length) buffer[index] += value;
};

const panGains = (pan) => [
  Math.cos(((pan + 1) * Math.PI) / 4),
  Math.sin(((pan + 1) * Math.PI) / 4),
];

const envelope = (t, dur, attack = 0.02, release = 0.2) => {
  const inGain = Math.min(1, t / Math.max(attack, 0.001));
  const outGain = Math.min(1, (dur - t) / Math.max(release, 0.001));
  return Math.max(0, Math.min(inGain, outGain));
};

const sine = ({
  start,
  dur,
  freq,
  endFreq = freq,
  amp,
  pan = 0,
  attack = 0.01,
  release = 0.12,
  harmonics = 0,
}) => {
  const begin = Math.floor(start * sampleRate);
  const count = Math.floor(dur * sampleRate);
  const [gainL, gainR] = panGains(pan);
  let phase = 0;
  for (let i = 0; i < count; i++) {
    const t = i / sampleRate;
    const progress = t / dur;
    const currentFreq = freq * Math.pow(endFreq / freq, progress);
    phase += (Math.PI * 2 * currentFreq) / sampleRate;
    const fundamental = Math.sin(phase);
    const color =
      fundamental +
      harmonics * 0.42 * Math.sin(phase * 2) +
      harmonics * 0.18 * Math.sin(phase * 3);
    const value = color * amp * envelope(t, dur, attack, release);
    add(left, begin + i, value * gainL);
    add(right, begin + i, value * gainR);
  }
};

const noise = ({
  start,
  dur,
  amp,
  pan = 0,
  attack = 0.001,
  release = 0.08,
  lowpass = 0.35,
}) => {
  const begin = Math.floor(start * sampleRate);
  const count = Math.floor(dur * sampleRate);
  const [gainL, gainR] = panGains(pan);
  let smooth = 0;
  for (let i = 0; i < count; i++) {
    const t = i / sampleRate;
    const raw = random() * 2 - 1;
    smooth += (raw - smooth) * lowpass;
    const value = smooth * amp * envelope(t, dur, attack, release);
    add(left, begin + i, value * gainL);
    add(right, begin + i, value * gainR);
  }
};

const impact = (start, amp = 0.55) => {
  sine({
    start,
    dur: 1.2,
    freq: 82,
    endFreq: 32,
    amp,
    release: 1.1,
    harmonics: 0.3,
  });
  noise({
    start,
    dur: 0.5,
    amp: amp * 0.42,
    release: 0.42,
    lowpass: 0.18,
  });
};

const whoosh = (start, dur = 0.65, amp = 0.18, pan = 0) => {
  const begin = Math.floor(start * sampleRate);
  const count = Math.floor(dur * sampleRate);
  const [gainL, gainR] = panGains(pan);
  let smooth = 0;
  for (let i = 0; i < count; i++) {
    const t = i / sampleRate;
    const p = t / dur;
    const raw = random() * 2 - 1;
    smooth += (raw - smooth) * (0.015 + p * 0.7);
    const value = smooth * amp * Math.sin(Math.PI * p) ** 1.7;
    add(left, begin + i, value * gainL);
    add(right, begin + i, value * gainR);
  }
};

const blip = (start, freq = 880, amp = 0.12, pan = 0) => {
  sine({
    start,
    dur: 0.08,
    freq,
    endFreq: freq * 1.04,
    amp,
    pan,
    release: 0.06,
  });
};

// Original interface sound design layered over the licensed music bed. Keeping
// these cues deterministic makes message, terminal, and transition timing exact.
noise({
  start: 0,
  dur: 9.6,
  amp: 0.018,
  attack: 0.2,
  release: 0.5,
  lowpass: 0.06,
});
const messageTimes = [1.27, 3.4, 5.53, 7.53, 23.67, 25.07, 26.17];
messageTimes.forEach((time, index) => {
  const incoming = index !== 1 && index !== 4;
  sine({
    start: time,
    dur: 0.23,
    freq: incoming ? 720 : 930,
    endFreq: incoming ? 1010 : 1220,
    amp: 0.15,
    pan: incoming ? -0.34 : 0.34,
    release: 0.17,
    harmonics: 0.35,
  });
  sine({
    start: time + 0.08,
    dur: 0.2,
    freq: incoming ? 1080 : 1390,
    amp: 0.08,
    pan: incoming ? -0.2 : 0.2,
    release: 0.18,
  });
});

whoosh(9.35, 0.85, 0.26, 0.25);
impact(9.78, 0.62);
whoosh(22.65, 0.75, 0.2, -0.25);
impact(23.0, 0.38);
whoosh(27.05, 0.8, 0.25, 0);
impact(27.48, 0.74);

// Keyboard taps and a rising confirmation tone for the service restart.
for (let t = 14.15, index = 0; t < 16.72; t += 0.105, index++) {
  noise({
    start: t,
    dur: 0.025,
    amp: 0.045,
    pan: Math.sin(index * 1.7) * 0.25,
    release: 0.02,
    lowpass: 0.82,
  });
}
sine({
  start: 18.02,
  dur: 0.4,
  freq: 620,
  endFreq: 980,
  amp: 0.13,
  pan: 0.15,
  release: 0.3,
  harmonics: 0.28,
});
sine({
  start: 18.14,
  dur: 0.38,
  freq: 930,
  endFreq: 1320,
  amp: 0.08,
  pan: -0.1,
  release: 0.3,
});

for (let t = 27.45, i = 0; t < 29.45; t += 0.25, i++) {
  blip(
    t,
    293.66 * Math.pow(2, (i % 8) / 12),
    0.035 + i * 0.003,
    Math.sin(i) * 0.45,
  );
}
sine({ start: 27.48, dur: 2.25, freq: 293.66, amp: 0.07, release: 2.1 });
sine({
  start: 27.48,
  dur: 2.25,
  freq: 440,
  amp: 0.045,
  attack: 0.01,
  release: 2.1,
  pan: 0.25,
});

let peak = 0;
for (let i = 0; i < length; i++) {
  const time = i / sampleRate;
  const gain =
    Math.min(1, time / 0.3) * Math.max(0, Math.min(1, (duration - time) / 0.7));
  left[i] *= gain;
  right[i] *= gain;
  peak = Math.max(peak, Math.abs(left[i]), Math.abs(right[i]));
}
const master = 0.92 / Math.max(peak, 0.001);
for (let i = 0; i < length; i++) {
  left[i] *= master;
  right[i] *= master;
}

const dataSize = length * 4;
const header = Buffer.alloc(44);
header.write("RIFF", 0);
header.writeUInt32LE(36 + dataSize, 4);
header.write("WAVE", 8);
header.write("fmt ", 12);
header.writeUInt32LE(16, 16);
header.writeUInt16LE(1, 20);
header.writeUInt16LE(2, 22);
header.writeUInt32LE(sampleRate, 24);
header.writeUInt32LE(sampleRate * 4, 28);
header.writeUInt16LE(4, 32);
header.writeUInt16LE(16, 34);
header.write("data", 36);
header.writeUInt32LE(dataSize, 40);

const pcm = Buffer.alloc(dataSize);
for (let i = 0; i < length; i++) {
  pcm.writeInt16LE(
    Math.round(Math.max(-1, Math.min(1, left[i])) * 32767),
    i * 4,
  );
  pcm.writeInt16LE(
    Math.round(Math.max(-1, Math.min(1, right[i])) * 32767),
    i * 4 + 2,
  );
}

const output = path.join(outputDir, "ssher-cloud-sfx.wav");
fs.writeFileSync(output, Buffer.concat([header, pcm]));
console.log(`Generated ${output} (${duration}s, ${sampleRate}Hz stereo)`);
