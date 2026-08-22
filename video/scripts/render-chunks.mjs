import fs from "node:fs";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const here = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(here, "..");
const out = path.join(root, "out");
const chunks = path.join(out, "chunks");
const remotion = path.join(
  root,
  "node_modules",
  ".bin",
  process.platform === "win32" ? "remotion.cmd" : "remotion",
);

fs.mkdirSync(chunks, { recursive: true });

const run = (command, args) => {
  const result = spawnSync(command, args, {
    cwd: root,
    stdio: "inherit",
    env: process.env,
  });
  if (result.status !== 0) {
    process.exit(result.status ?? 1);
  }
};

const chunkSize = 120;
const totalFrames = 1080;
const files = [];

for (let start = 0; start < totalFrames; start += chunkSize) {
  const end = Math.min(totalFrames - 1, start + chunkSize - 1);
  const index = String(files.length).padStart(2, "0");
  const file = path.join(chunks, `part-${index}.mp4`);
  files.push(file);
  run(remotion, [
    "render",
    "SsherLaunch",
    file,
    `--frames=${start}-${end}`,
    "--muted",
    "--codec=h264",
    "--crf=16",
    "--jpeg-quality=60",
    "--concurrency=2",
  ]);
}

const concatFile = path.join(chunks, "concat.txt");
fs.writeFileSync(
  concatFile,
  files.map((file) => `file '${file.replaceAll("'", "'\\''")}'`).join("\n"),
);

const silent = path.join(out, "ssher-v0.1.2-silent.mp4");
const final = path.join(out, "ssher-v0.1.2-launch.mp4");
const score = path.join(root, "public", "audio", "ssher-score.wav");

run("ffmpeg", [
  "-y",
  "-f",
  "concat",
  "-safe",
  "0",
  "-i",
  concatFile,
  "-c",
  "copy",
  silent,
]);

run("ffmpeg", [
  "-y",
  "-i",
  silent,
  "-i",
  score,
  "-c:v",
  "copy",
  "-c:a",
  "aac",
  "-b:a",
  "256k",
  "-shortest",
  "-movflags",
  "+faststart",
  final,
]);

console.log(`Rendered ${final}`);
