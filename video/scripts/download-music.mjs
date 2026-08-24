import { createHash } from "node:crypto";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const source = "https://assets.mixkit.co/music/31/31.mp3";
const expectedHash =
  "8c89819547b42a80750fb25f37a960a1f45f6fd66bc8d784ac817b98b002897c";
const here = path.dirname(fileURLToPath(import.meta.url));
const output = path.join(
  here,
  "..",
  "public",
  "audio",
  "mixkit-dreaming-big-31.mp3",
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
