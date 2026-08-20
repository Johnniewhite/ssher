import { createHash } from "node:crypto";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const source = "https://assets.mixkit.co/music/32/32.mp3";
const expectedHash =
  "e3c88488e65b8c87a6f06120983ce2cb12ea3aeba99f8cadb7ee5d6d284ef2c6";
const here = path.dirname(fileURLToPath(import.meta.url));
const output = path.join(
  here,
  "..",
  "public",
  "audio",
  "mixkit-driving-ambition-32.mp3",
);

const sha256 = (contents) =>
  createHash("sha256").update(contents).digest("hex");

if (fs.existsSync(output)) {
  const existing = fs.readFileSync(output);
  if (sha256(existing) === expectedHash) {
    console.log(`Using verified ${output}`);
    process.exit(0);
  }
}

const response = await fetch(source);
if (!response.ok) {
  throw new Error(
    `Music download failed: ${response.status} ${response.statusText}`,
  );
}

const contents = Buffer.from(await response.arrayBuffer());
const actualHash = sha256(contents);
if (actualHash !== expectedHash) {
  throw new Error(
    `Music integrity check failed: expected ${expectedHash}, received ${actualHash}`,
  );
}

fs.mkdirSync(path.dirname(output), { recursive: true });
fs.writeFileSync(output, contents);
console.log(`Downloaded and verified ${output}`);
