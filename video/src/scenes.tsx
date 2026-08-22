import {
  AbsoluteFill,
  useCurrentFrame,
} from "remotion";
import {
  CircularMeter,
  DataBeam,
  Eyebrow,
  FullBleedFlash,
  GlitchText,
  Pill,
  Reveal,
  SceneLabel,
  ServerNode,
  ShieldMark,
  Stage,
  TerminalWindow,
  Wordmark,
  center,
} from "./components";
import { clamp, ease, editorial, enterExit, pop } from "./motion";
import { colors, mono, sans, shadow } from "./theme";

const commandFragments = [
  "ssh -i ~/.ssh/prod.pem -p 2222 deploy@10.0.4.21",
  "scp -o ProxyCommand='ssh -W %h:%p bastion' build.tar.gz…",
  "Host production-web",
  "IdentityFile ~/.ssh/id_ed25519_prod",
  "rsync -az --progress -e 'ssh -p 2222'…",
  "export SSHPASS='••••••••••••'",
  "ssh-add ~/.ssh/company-prod",
  "ProxyJump bastion.internal",
];

export const ColdOpen: React.FC = () => {
  const frame = useCurrentFrame();
  const opacity = enterExit(frame, 110, 10, 12);
  const chaos = clamp(frame, [0, 82], [0, 1]);
  const statement = ease(frame, [55, 78]);
  const collapse = ease(frame, [84, 104]);
  return (
    <Stage glow={colors.red}>
      <AbsoluteFill style={{ opacity }}>
        {commandFragments.map((text, index) => {
          const side = index % 2 === 0 ? -1 : 1;
          const y = 125 + index * 104;
          const delay = index * 4;
          const fly = ease(frame, [delay, delay + 16]);
          const jitter = frame % (11 + index) < 2 ? side * 10 : 0;
          return (
            <div
              key={text}
              style={{
                position: "absolute",
                left: side < 0 ? 110 : 790,
                top: y,
                width: 950,
                padding: "20px 25px",
                border: "1px solid rgba(255,107,112,.15)",
                borderRadius: 12,
                background: "rgba(12,9,9,.76)",
                color: index % 3 === 0 ? colors.red : colors.gray,
                fontFamily: mono,
                fontSize: 20,
                whiteSpace: "nowrap",
                overflow: "hidden",
                opacity: fly * (1 - collapse),
                transform: `translateX(${(1 - fly) * side * 500 + jitter + collapse * side * 600}px) rotate(${side * (1 - fly) * 2.5}deg)`,
                boxShadow: "0 24px 60px rgba(0,0,0,.32)",
              }}
            >
              <span style={{ color: colors.dim, marginRight: 16 }}>
                {String(index + 1).padStart(2, "0")}
              </span>
              {text}
            </div>
          );
        })}
        <AbsoluteFill
          style={{
            ...center,
            background: `rgba(2,7,5,${statement * 0.68})`,
            backdropFilter: `blur(${statement * 8}px)`,
          }}
        >
          <div
            style={{
              textAlign: "center",
              transform: `scale(${0.88 + statement * 0.12})`,
              opacity: statement,
            }}
          >
            <Eyebrow color={colors.red}>The old way</Eyebrow>
            <GlitchText
              text="SSH SHOULDN'T"
              intensity={2}
              style={{
                fontFamily: sans,
                fontWeight: 800,
                fontSize: 118,
                letterSpacing: -7,
                lineHeight: 0.94,
                marginTop: 30,
              }}
            />
            <GlitchText
              text="FEEL LIKE THIS."
              intensity={2}
              style={{
                fontFamily: sans,
                fontWeight: 800,
                fontSize: 118,
                letterSpacing: -7,
                lineHeight: 1,
                color: colors.red,
              }}
            />
          </div>
        </AbsoluteFill>
        <div
          style={{
            position: "absolute",
            left: 0,
            top: `${(chaos * 1130) % 1080}px`,
            width: "100%",
            height: 2,
            background: "rgba(255,107,112,.35)",
            boxShadow: "0 0 16px rgba(255,107,112,.7)",
          }}
        />
      </AbsoluteFill>
      <FullBleedFlash at={103} color={colors.green} duration={7} />
    </Stage>
  );
};

export const BrandReveal: React.FC = () => {
  const frame = useCurrentFrame();
  const shield = ease(frame, [4, 50]);
  const word = pop(frame, [35, 67]);
  const tagline = ease(frame, [62, 84]);
  const orbit = frame * 1.5;
  return (
    <Stage glow={colors.green} grid={false}>
      <SceneLabel number="01">Meet your SSH companion</SceneLabel>
      <AbsoluteFill style={{ ...center }}>
        <div
          style={{
            position: "absolute",
            width: 900,
            height: 900,
            borderRadius: "50%",
            background:
              "radial-gradient(circle, rgba(76,255,139,.13), rgba(76,255,139,.025) 42%, transparent 67%)",
            transform: `scale(${0.7 + shield * 0.3})`,
          }}
        />
        {[0, 1, 2].map((index) => (
          <div
            key={index}
            style={{
              position: "absolute",
              width: 520 + index * 120,
              height: 520 + index * 120,
              borderRadius: "50%",
              border: `1px ${index === 1 ? "dashed" : "solid"} rgba(76,255,139,${0.19 - index * 0.035})`,
              transform: `rotate(${orbit * (index % 2 ? -0.45 : 0.25)}deg) scale(${shield})`,
            }}
          >
            <span
              style={{
                position: "absolute",
                left: "50%",
                top: -5,
                width: 9,
                height: 9,
                borderRadius: "50%",
                background: index === 1 ? colors.cyan : colors.green,
                boxShadow: `0 0 20px ${index === 1 ? colors.cyan : colors.green}`,
              }}
            />
          </div>
        ))}
        <div
          style={{
            transform: `translateX(${-300 + shield * 120}px) scale(${0.7 + shield * 0.3})`,
          }}
        >
          <ShieldMark size={430} progress={shield} />
        </div>
        <div
          style={{
            marginLeft: -30,
            opacity: word,
            transform: `translateX(${(1 - word) * 120}px)`,
          }}
        >
          <Wordmark scale={1.8} showVersion />
          <div
            style={{
              fontSize: 42,
              fontWeight: 600,
              letterSpacing: -1.8,
              color: colors.mint,
              marginTop: 28,
              opacity: tagline,
              transform: `translateY(${(1 - tagline) * 18}px)`,
            }}
          >
            Your SSH fleet.{" "}
            <span style={{ color: colors.green }}>One command away.</span>
          </div>
        </div>
      </AbsoluteFill>
      <FullBleedFlash at={1} duration={9} />
    </Stage>
  );
};

const secretTokens = [
  "deploy@prod",
  "10.0.4.21",
  "id_ed25519",
  "••••••••",
  "bastion",
  "production",
  "port:22",
  "web-cluster",
];

export const VaultScene: React.FC = () => {
  const frame = useCurrentFrame();
  const lock = ease(frame, [24, 82]);
  const seal = pop(frame, [82, 108]);
  return (
    <Stage glow={colors.cyan}>
      <SceneLabel number="02">Encrypted by design</SceneLabel>
      <div
        style={{
          position: "absolute",
          left: 110,
          top: 250,
          width: 620,
        }}
      >
        <Reveal from={8}>
          <Eyebrow>Local-first vault</Eyebrow>
        </Reveal>
        <Reveal from={18}>
          <div
            style={{
              fontSize: 82,
              fontWeight: 700,
              letterSpacing: -5,
              lineHeight: 1.02,
              marginTop: 30,
            }}
          >
            Secrets in.
            <br />
            <span style={{ color: colors.green }}>Ciphertext out.</span>
          </div>
        </Reveal>
        <Reveal from={36}>
          <p
            style={{
              fontSize: 25,
              color: colors.gray,
              lineHeight: 1.6,
              marginTop: 32,
              width: 530,
            }}
          >
            AES-256-GCM. Argon2id. Your credentials stay encrypted on your
            machine.
          </p>
        </Reveal>
        <Reveal from={54}>
          <div style={{ display: "flex", gap: 12, marginTop: 32 }}>
            <Pill>AES-256-GCM</Pill>
            <Pill color={colors.cyan}>ARGON2ID</Pill>
            <Pill>0600</Pill>
          </div>
        </Reveal>
      </div>
      <div
        style={{
          position: "absolute",
          right: 80,
          top: 138,
          width: 900,
          height: 800,
          ...center,
        }}
      >
        {secretTokens.map((token, index) => {
          const angle = (index / secretTokens.length) * Math.PI * 2 + frame / 55;
          const radius = 325 - lock * 240;
          const x = Math.cos(angle) * radius;
          const y = Math.sin(angle) * radius * 0.66;
          const fade = 1 - clamp(lock, [0.65, 1], [0, 1]);
          return (
            <div
              key={token}
              style={{
                position: "absolute",
                left: 450 + x - 70,
                top: 400 + y - 22,
                padding: "10px 15px",
                border: "1px solid rgba(93,231,255,.25)",
                borderRadius: 7,
                background: "rgba(6,20,18,.78)",
                color: index % 2 ? colors.gray : colors.cyan,
                fontFamily: mono,
                fontSize: 14,
                opacity: fade,
                filter: `blur(${lock * 3}px)`,
              }}
            >
              {token}
            </div>
          );
        })}
        {[0, 1, 2, 3].map((index) => (
          <div
            key={index}
            style={{
              position: "absolute",
              width: 500 + index * 90,
              height: 500 + index * 90,
              borderRadius: "50%",
              border: `1px ${index % 2 ? "dashed" : "solid"} rgba(93,231,255,${0.15 - index * 0.025})`,
              transform: `rotate(${frame * (index % 2 ? -0.3 : 0.22)}deg) scale(${0.7 + lock * 0.3})`,
              opacity: 0.6,
            }}
          />
        ))}
        <div
          style={{
            transform: `scale(${0.72 + lock * 0.28})`,
            filter: `drop-shadow(0 0 ${seal * 55}px rgba(76,255,139,.35))`,
          }}
        >
          <ShieldMark size={450} progress={lock} />
        </div>
        <div
          style={{
            position: "absolute",
            bottom: 64,
            display: "flex",
            gap: 18,
            opacity: seal,
            transform: `translateY(${(1 - seal) * 20}px)`,
          }}
        >
          <Pill>✓ VAULT SEALED</Pill>
          <Pill color={colors.cyan}>LOCAL ONLY</Pill>
        </div>
      </div>
    </Stage>
  );
};

export const ConnectScene: React.FC = () => {
  const frame = useCurrentFrame();
  const panel = ease(frame, [8, 36]);
  const tunnel = ease(frame, [48, 94]);
  const pulse = 0.5 + Math.sin(frame / 5) * 0.5;
  return (
    <Stage>
      <SceneLabel number="03">Connect naturally</SceneLabel>
      <div
        style={{
          position: "absolute",
          left: 100,
          top: 185,
          opacity: panel,
          transform: `translateY(${(1 - panel) * 50}px) rotateX(${(1 - panel) * 7}deg)`,
        }}
      >
        <TerminalWindow
          width={1120}
          command="ssher prod"
          typeFrom={25}
          lines={[
            { text: "[i] matched alias → production-web", color: colors.gray },
            { text: "[ok] connected to production-web", color: colors.green },
            { text: "deploy@web1:~$ _", color: colors.mint },
          ]}
        />
      </div>
      <div
        style={{
          position: "absolute",
          right: 82,
          top: 205,
          width: 620,
          height: 640,
        }}
      >
        <div
          style={{
            position: "absolute",
            left: 30,
            top: 270,
            width: 112,
            height: 112,
            borderRadius: 28,
            background: "rgba(76,255,139,.1)",
            border: "1px solid rgba(76,255,139,.45)",
            boxShadow: `0 0 ${20 + pulse * 28}px rgba(76,255,139,.2)`,
            color: colors.green,
            fontFamily: mono,
            fontSize: 25,
            ...center,
          }}
        >
          &gt;_
        </div>
        <DataBeam
          x1={142}
          y1={326}
          x2={445}
          y2={326}
          progress={tunnel}
        />
        {Array.from({ length: 6 }, (_, index) => {
          const packet = (tunnel * 1.8 + index / 6) % 1;
          return (
            <div
              key={index}
              style={{
                position: "absolute",
                left: 132 + packet * 325,
                top: 318,
                width: 15,
                height: 15,
                borderRadius: 4,
                background: index % 2 ? colors.green : colors.cyan,
                boxShadow: `0 0 18px ${index % 2 ? colors.green : colors.cyan}`,
                opacity: tunnel,
                transform: `rotate(${frame * 6 + index * 40}deg)`,
              }}
            />
          );
        })}
        <div
          style={{
            position: "absolute",
            left: 445,
            top: 230,
            width: 160,
            height: 195,
            borderRadius: 18,
            background: "linear-gradient(145deg, #10281A, #07100B)",
            border: "1px solid rgba(76,255,139,.4)",
            boxShadow: shadow,
            opacity: tunnel,
            transform: `scale(${0.8 + tunnel * 0.2})`,
            ...center,
            flexDirection: "column",
          }}
        >
          {[0, 1, 2].map((index) => (
            <div
              key={index}
              style={{
                width: 105,
                height: 28,
                border: "1px solid rgba(76,255,139,.2)",
                borderRadius: 5,
                margin: 7,
                background: "rgba(76,255,139,.04)",
                display: "flex",
                alignItems: "center",
                padding: "0 10px",
              }}
            >
              <span
                style={{
                  width: 5,
                  height: 5,
                  borderRadius: "50%",
                  background: colors.green,
                  boxShadow: `0 0 9px ${colors.green}`,
                }}
              />
            </div>
          ))}
          <span
            style={{
              fontFamily: mono,
              color: colors.green,
              fontSize: 12,
              marginTop: 8,
            }}
          >
            PRODUCTION
          </span>
        </div>
        <div
          style={{
            position: "absolute",
            left: 155,
            top: 400,
            opacity: tunnel,
            color: colors.gray,
            fontFamily: mono,
            fontSize: 14,
            letterSpacing: 1.5,
          }}
        >
          HOST KEY VERIFIED · 218ms
        </div>
      </div>
    </Stage>
  );
};

const fleetNodes = [
  { name: "web-01", meta: "uptime 42d · 0.12", x: 1110, y: 185 },
  { name: "web-02", meta: "uptime 37d · 0.08", x: 1440, y: 315 },
  { name: "api-01", meta: "uptime 29d · 0.21", x: 1270, y: 510 },
  { name: "db-01", meta: "uptime 81d · 0.05", x: 930, y: 660 },
  { name: "cache", meta: "uptime 64d · 0.03", x: 690, y: 435 },
];

export const FleetScene: React.FC = () => {
  const frame = useCurrentFrame();
  const fire = ease(frame, [43, 88]);
  const result = ease(frame, [78, 118]);
  return (
    <Stage glow={colors.greenBright}>
      <SceneLabel number="04">Parallel fleet execution</SceneLabel>
      <div style={{ position: "absolute", left: 105, top: 230, width: 700 }}>
        <Reveal from={6}>
          <Eyebrow>One command</Eyebrow>
        </Reveal>
        <Reveal from={16}>
          <div
            style={{
              fontSize: 90,
              fontWeight: 700,
              lineHeight: 1,
              letterSpacing: -5.5,
              marginTop: 26,
            }}
          >
            Every server.
            <br />
            <span style={{ color: colors.green }}>At once.</span>
          </div>
        </Reveal>
        <Reveal from={32}>
          <div
            style={{
              marginTop: 46,
              border: "1px solid rgba(76,255,139,.2)",
              borderRadius: 13,
              background: "rgba(4,12,8,.9)",
              padding: "25px 28px",
              fontFamily: mono,
              fontSize: 21,
              color: colors.mint,
              boxShadow: shadow,
            }}
          >
            <span style={{ color: colors.green }}>$ </span>
            ssher exec "uptime" --all
          </div>
        </Reveal>
        <Reveal from={82}>
          <div style={{ display: "flex", gap: 14, marginTop: 28 }}>
            <Pill>5/5 ONLINE</Pill>
            <Pill color={colors.cyan}>218ms AVG</Pill>
          </div>
        </Reveal>
      </div>
      <div
        style={{
          position: "absolute",
          left: 0,
          top: 0,
          width: 1920,
          height: 1080,
        }}
      >
        <div
          style={{
            position: "absolute",
            left: 936,
            top: 440,
            width: 150,
            height: 150,
            borderRadius: "50%",
            background: "rgba(76,255,139,.11)",
            border: "1px solid rgba(76,255,139,.5)",
            boxShadow: `0 0 ${fire * 90}px rgba(76,255,139,.25)`,
            color: colors.green,
            fontFamily: mono,
            fontWeight: 600,
            fontSize: 24,
            transform: `scale(${0.7 + fire * 0.3})`,
            ...center,
          }}
        >
          EXEC
        </div>
        {fleetNodes.map((node, index) => {
          const nodeProgress = Math.max(
            0,
            Math.min(1, (fire - index * 0.08) / (1 - index * 0.08)),
          );
          return (
            <div key={node.name}>
              <DataBeam
                x1={1011}
                y1={515}
                x2={node.x + 122}
                y2={node.y + 56}
                progress={nodeProgress}
                color={index % 2 ? colors.cyan : colors.green}
              />
              <ServerNode
                {...node}
                progress={result}
                delay={index * 0.08}
              />
            </div>
          );
        })}
      </div>
      <FullBleedFlash at={43} duration={5} />
    </Stage>
  );
};

export const TransferScene: React.FC = () => {
  const frame = useCurrentFrame();
  const progress = ease(frame, [18, 96]);
  const complete = pop(frame, [86, 112]);
  const curve = (value: number) => ({
    x: 505 + value * 900,
    y: 585 - Math.sin(value * Math.PI) * 270,
  });
  return (
    <Stage glow={colors.cyan}>
      <SceneLabel number="05">Native SFTP</SceneLabel>
      <AbsoluteFill>
        <div
          style={{
            position: "absolute",
            left: 110,
            top: 250,
            width: 470,
          }}
        >
          <Reveal from={5}>
            <Eyebrow color={colors.cyan}>Files, in flight</Eyebrow>
          </Reveal>
          <Reveal from={16}>
            <div
              style={{
                fontSize: 82,
                fontWeight: 700,
                letterSpacing: -5,
                lineHeight: 1,
                marginTop: 28,
              }}
            >
              Move data.
              <br />
              <span style={{ color: colors.cyan }}>Stay native.</span>
            </div>
          </Reveal>
          <Reveal from={38}>
            <p
              style={{
                color: colors.gray,
                fontSize: 24,
                lineHeight: 1.6,
                marginTop: 30,
              }}
            >
              Upload and download over the same trusted connection. No extra
              toolchain.
            </p>
          </Reveal>
        </div>
        <div
          style={{
            position: "absolute",
            left: 430,
            top: 100,
            width: 1050,
            height: 700,
          }}
        >
          <svg
            width="1050"
            height="700"
            viewBox="0 0 1050 700"
            style={{ position: "absolute" }}
          >
            <path
              d="M75 485 C340 90 700 90 975 485"
              fill="none"
              stroke="rgba(93,231,255,.14)"
              strokeWidth="3"
              strokeDasharray="10 12"
            />
            <path
              d="M75 485 C340 90 700 90 975 485"
              fill="none"
              stroke={colors.cyan}
              strokeWidth="5"
              strokeLinecap="round"
              pathLength="1"
              strokeDasharray="1"
              strokeDashoffset={1 - progress}
              style={{ filter: "drop-shadow(0 0 9px rgba(93,231,255,.65))" }}
            />
          </svg>
          {[0, 1, 2, 3, 4, 5].map((index) => {
            const value = Math.max(0, Math.min(1, progress - index * 0.08));
            const point = curve(value);
            return (
              <div
                key={index}
                style={{
                  position: "absolute",
                  left: point.x - 430,
                  top: point.y - 100,
                  width: 38,
                  height: 38,
                  borderRadius: 7,
                  border: "1px solid rgba(93,231,255,.75)",
                  background: "rgba(93,231,255,.16)",
                  boxShadow: "0 0 28px rgba(93,231,255,.32)",
                  transform: `rotate(${frame * 4 + index * 35}deg) scale(${value > 0 ? 1 : 0})`,
                  color: colors.cyan,
                  fontFamily: mono,
                  fontSize: 10,
                  ...center,
                }}
              >
                {index.toString(16).toUpperCase()}
              </div>
            );
          })}
        </div>
        {[
          { left: 400, label: "LOCAL", icon: "◫" },
          { left: 1400, label: "REMOTE", icon: "▦" },
        ].map((node) => (
          <div
            key={node.label}
            style={{
              position: "absolute",
              left: node.left,
              top: 645,
              width: 215,
              height: 175,
              borderRadius: 22,
              border: "1px solid rgba(93,231,255,.32)",
              background: "linear-gradient(145deg, #0C211A, #06100C)",
              boxShadow: shadow,
              ...center,
              flexDirection: "column",
            }}
          >
            <span
              style={{
                color: colors.cyan,
                fontSize: 52,
                filter: "drop-shadow(0 0 18px rgba(93,231,255,.4))",
              }}
            >
              {node.icon}
            </span>
            <span
              style={{
                marginTop: 18,
                color: colors.gray,
                fontFamily: mono,
                letterSpacing: 2.2,
                fontSize: 13,
              }}
            >
              {node.label}
            </span>
          </div>
        ))}
        <div
          style={{
            position: "absolute",
            left: 755,
            top: 750,
            width: 520,
            opacity: complete,
            transform: `translateY(${(1 - complete) * 20}px)`,
          }}
        >
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              fontFamily: mono,
              color: colors.mint,
              fontSize: 15,
            }}
          >
            <span>build.tar.gz</span>
            <span style={{ color: colors.cyan }}>42.8 MB · 100%</span>
          </div>
          <div
            style={{
              height: 8,
              marginTop: 13,
              borderRadius: 99,
              background: "rgba(93,231,255,.12)",
              overflow: "hidden",
            }}
          >
            <div
              style={{
                width: `${progress * 100}%`,
                height: "100%",
                background: colors.cyan,
                boxShadow: `0 0 16px ${colors.cyan}`,
              }}
            />
          </div>
        </div>
      </AbsoluteFill>
    </Stage>
  );
};

const toolbox = [
  {
    command: "ssher prod --record",
    label: "SESSION RECORDING",
    accent: colors.green,
    glyph: "●",
  },
  {
    command: "ssher wrap -e ssh user@host",
    label: "SSHPASS WRAPPER",
    accent: colors.amber,
    glyph: "⌁",
  },
  {
    command: "ssher prod -L 8080:localhost:80",
    label: "PORT FORWARDING",
    accent: colors.cyan,
    glyph: "⇄",
  },
  {
    command: "ssher completion zsh",
    label: "SHELL COMPLETION",
    accent: colors.greenBright,
    glyph: "⌘",
  },
];

export const ToolboxScene: React.FC = () => {
  const frame = useCurrentFrame();
  return (
    <Stage glow={colors.amber}>
      <SceneLabel number="06">More than connections</SceneLabel>
      <div
        style={{
          position: "absolute",
          left: 110,
          top: 210,
          width: 650,
        }}
      >
        <Reveal from={4}>
          <Eyebrow color={colors.amber}>Power tools included</Eyebrow>
        </Reveal>
        <Reveal from={14}>
          <div
            style={{
              fontSize: 84,
              fontWeight: 700,
              lineHeight: 1,
              letterSpacing: -5,
              marginTop: 28,
            }}
          >
            One binary.
            <br />
            <span style={{ color: colors.amber }}>Deep toolkit.</span>
          </div>
        </Reveal>
      </div>
      <div
        style={{
          position: "absolute",
          left: 760,
          top: 135,
          width: 1060,
          height: 790,
          perspective: 1200,
        }}
      >
        {toolbox.map((item, index) => {
          const entry = pop(frame, [18 + index * 12, 44 + index * 12]);
          const y = index * 170;
          const wave = Math.sin(frame / 28 + index) * 7;
          return (
            <div
              key={item.command}
              style={{
                position: "absolute",
                left: index % 2 ? 120 : 20,
                top: 35 + y,
                width: 900,
                minHeight: 138,
                border: `1px solid ${item.accent}38`,
                borderRadius: 17,
                background:
                  "linear-gradient(100deg, rgba(7,19,13,.97), rgba(12,30,21,.94))",
                boxShadow: shadow,
                padding: "25px 30px",
                opacity: entry,
                transform: `translateX(${(1 - entry) * 280}px) translateY(${wave}px) rotateY(${-4 + index * 1.1}deg)`,
                display: "grid",
                gridTemplateColumns: "72px 1fr auto",
                gap: 22,
                alignItems: "center",
              }}
            >
              <div
                style={{
                  width: 64,
                  height: 64,
                  borderRadius: 16,
                  border: `1px solid ${item.accent}66`,
                  background: `${item.accent}12`,
                  color: item.accent,
                  fontSize: 26,
                  boxShadow: `0 0 28px ${item.accent}15`,
                  ...center,
                }}
              >
                {item.glyph}
              </div>
              <div>
                <div
                  style={{
                    color: item.accent,
                    fontFamily: mono,
                    fontSize: 13,
                    letterSpacing: 2,
                  }}
                >
                  {item.label}
                </div>
                <div
                  style={{
                    color: colors.white,
                    fontFamily: mono,
                    fontSize: 18,
                    marginTop: 12,
                  }}
                >
                  <span style={{ color: colors.green }}>$ </span>
                  {item.command}
                </div>
              </div>
              <div
                style={{
                  color: colors.dim,
                  fontFamily: mono,
                  fontSize: 14,
                }}
              >
                0{index + 1}
              </div>
            </div>
          );
        })}
      </div>
      <div
        style={{
          position: "absolute",
          left: 110,
          bottom: 180,
          display: "grid",
          gridTemplateColumns: "repeat(2, 200px)",
          gap: 18,
        }}
      >
        <CircularMeter
          progress={editorial(frame, [44, 105])}
          value="∞"
          label="PROFILES"
        />
        <CircularMeter
          progress={editorial(frame, [55, 116])}
          value="1"
          label="BINARY"
        />
      </div>
    </Stage>
  );
};

const codeLines = [
  "package main",
  'import "github.com/johnniewhite/ssher"',
  "",
  "vault := encrypt.Argon2id(credentials)",
  "client := ssh.Dial(target, knownHosts)",
  'fleet.Exec("uptime", parallel.All)',
  "",
  "// yours to inspect. yours to improve.",
];

export const OpenSourceScene: React.FC = () => {
  const frame = useCurrentFrame();
  const codeProgress = clamp(frame, [8, 76], [0, codeLines.join("\n").length]);
  const glow = ease(frame, [55, 94]);
  return (
    <Stage glow={colors.cyan}>
      <SceneLabel number="07">Built in public</SceneLabel>
      <div
        style={{
          position: "absolute",
          left: 105,
          top: 190,
          width: 880,
          height: 650,
          border: "1px solid rgba(200,255,218,.16)",
          borderRadius: 18,
          background: "rgba(3,11,7,.92)",
          overflow: "hidden",
          boxShadow: shadow,
          transform: `perspective(1200px) rotateY(${3 - glow * 3}deg)`,
        }}
      >
        <div
          style={{
            height: 60,
            borderBottom: "1px solid rgba(200,255,218,.1)",
            padding: "0 24px",
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            color: colors.gray,
            fontFamily: mono,
            fontSize: 13,
          }}
        >
          <span>github.com/Johnniewhite/ssher</span>
          <span style={{ color: colors.green }}>PUBLIC · MAIN</span>
        </div>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "72px 1fr",
            padding: "30px 0",
            fontFamily: mono,
            fontSize: 20,
            lineHeight: 2,
          }}
        >
          <div style={{ color: colors.dim, textAlign: "right", paddingRight: 22 }}>
            {codeLines.map((_, index) => (
              <div key={index}>{index + 1}</div>
            ))}
          </div>
          <pre
            style={{
              margin: 0,
              color: colors.mint,
              whiteSpace: "pre-wrap",
            }}
          >
            {codeLines.join("\n").slice(0, Math.floor(codeProgress))}
            <span style={{ color: colors.green }}>▋</span>
          </pre>
        </div>
      </div>
      <div
        style={{
          position: "absolute",
          right: 110,
          top: 230,
          width: 700,
        }}
      >
        <Reveal from={20}>
          <Eyebrow>Open source · MIT</Eyebrow>
        </Reveal>
        <Reveal from={34}>
          <div
            style={{
              fontSize: 80,
              fontWeight: 700,
              lineHeight: 1.02,
              letterSpacing: -5,
              marginTop: 28,
            }}
          >
            Audit it.
            <br />
            Fork it.
            <br />
            <span style={{ color: colors.green }}>Make it yours.</span>
          </div>
        </Reveal>
        <Reveal from={62}>
          <div
            style={{
              display: "flex",
              gap: 13,
              marginTop: 38,
              flexWrap: "wrap",
            }}
          >
            <Pill>GO</Pill>
            <Pill color={colors.cyan}>MIT LICENSE</Pill>
            <Pill>MACOS + LINUX</Pill>
          </div>
        </Reveal>
        <Reveal from={76}>
          <div
            style={{
              marginTop: 42,
              color: colors.gray,
              fontFamily: mono,
              fontSize: 18,
              letterSpacing: 1,
            }}
          >
            github.com/Johnniewhite/ssher
          </div>
        </Reveal>
      </div>
    </Stage>
  );
};

const finaleFeatures = [
  "ENCRYPTED VAULT",
  "NATIVE SSH",
  "PARALLEL EXEC",
  "NATIVE SFTP",
  "SESSION RECORDING",
  "PORT FORWARDING",
  "OPEN SOURCE",
];

export const FinaleScene: React.FC = () => {
  const frame = useCurrentFrame();
  const duration = 175;
  const featureBurst = clamp(frame, [0, 74], [0, finaleFeatures.length]);
  const logo = pop(frame, [60, 92]);
  const copy = ease(frame, [76, 104]);
  const install = pop(frame, [100, 132]);
  const finalHit = pop(frame, [139, 160]);
  const out = ease(duration - frame, [0, 12]);
  return (
    <Stage glow={colors.greenBright} grid={false}>
      <AbsoluteFill style={{ opacity: out }}>
        {finaleFeatures.map((feature, index) => {
          const angle = (index / finaleFeatures.length) * Math.PI * 2 - 0.7;
          const visible = ease(featureBurst, [index, index + 0.7]);
          const radius = 540 + index * 8;
          const x = 960 + Math.cos(angle) * radius;
          const y = 500 + Math.sin(angle) * radius * 0.62;
          return (
            <div
              key={feature}
              style={{
                position: "absolute",
                left: x - 120,
                top: y - 22,
                width: 240,
                textAlign: "center",
                color: index % 3 === 1 ? colors.cyan : colors.green,
                fontFamily: mono,
                fontSize: 14,
                letterSpacing: 2.1,
                opacity: visible * (1 - clamp(frame, [62, 88], [0, 0.72])),
                transform: `scale(${0.7 + visible * 0.3}) translateY(${Math.sin(frame / 20 + index) * 7}px)`,
              }}
            >
              {feature}
            </div>
          );
        })}
        <div
          style={{
            position: "absolute",
            left: "50%",
            top: "46%",
            transform: `translate(-50%, -50%) scale(${0.65 + logo * 0.35})`,
            opacity: logo,
            textAlign: "center",
          }}
        >
          <div
            style={{
              position: "absolute",
              width: 900,
              height: 900,
              left: "50%",
              top: "50%",
              transform: "translate(-50%, -50%)",
              borderRadius: "50%",
              background:
                "radial-gradient(circle, rgba(76,255,139,.16), rgba(76,255,139,.03) 45%, transparent 70%)",
            }}
          />
          <div style={{ position: "relative", ...center }}>
            <ShieldMark size={280} progress={logo} />
            <div style={{ marginLeft: -22 }}>
              <Wordmark scale={2.15} showVersion />
            </div>
          </div>
          <div
            style={{
              position: "relative",
              fontSize: 52,
              fontWeight: 600,
              letterSpacing: -2.5,
              marginTop: 2,
              opacity: copy,
              transform: `translateY(${(1 - copy) * 20}px)`,
            }}
          >
            The SSH companion that{" "}
            <span style={{ color: colors.green }}>remembers.</span>
          </div>
          <div
            style={{
              position: "relative",
              margin: "48px auto 0",
              width: 720,
              minHeight: 82,
              border: "1px solid rgba(76,255,139,.34)",
              borderRadius: 13,
              background: "rgba(5,16,10,.92)",
              boxShadow: `0 25px 70px rgba(0,0,0,.5), 0 0 ${install * 45}px rgba(76,255,139,.12)`,
              color: colors.mint,
              fontFamily: mono,
              fontSize: 22,
              opacity: install,
              transform: `translateY(${(1 - install) * 28}px) scale(${0.92 + install * 0.08})`,
              ...center,
            }}
          >
            <span style={{ color: colors.green, marginRight: 14 }}>$</span>
            brew install ssher
          </div>
          <div
            style={{
              position: "relative",
              marginTop: 34,
              color: colors.greenBright,
              fontFamily: mono,
              fontWeight: 600,
              fontSize: 27,
              letterSpacing: 3.4,
              opacity: finalHit,
              transform: `scale(${0.85 + finalHit * 0.15})`,
              textShadow: "0 0 25px rgba(76,255,139,.38)",
            }}
          >
            GETSSHER.COM
          </div>
        </div>
        <FullBleedFlash at={139} duration={11} />
      </AbsoluteFill>
    </Stage>
  );
};
