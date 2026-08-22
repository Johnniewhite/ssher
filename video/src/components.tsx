import type { CSSProperties, ReactNode } from "react";
import {
  AbsoluteFill,
  interpolate,
  useCurrentFrame,
} from "remotion";
import { clamp, ease, seeded } from "./motion";
import { colors, mono, sans, shadow } from "./theme";

export const center: CSSProperties = {
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
};

export const Stage: React.FC<{
  children: ReactNode;
  glow?: string;
  grid?: boolean;
}> = ({ children, glow = colors.green, grid = true }) => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill
      style={{
        background: colors.black,
        color: colors.white,
        fontFamily: sans,
        overflow: "hidden",
      }}
    >
      <AbsoluteFill
        style={{
          background: `
            radial-gradient(circle at ${68 + Math.sin(frame / 80) * 5}% ${32 + Math.cos(frame / 90) * 5}%, ${glow}22 0, transparent 34%),
            radial-gradient(circle at 14% 84%, #5DE7FF0C 0, transparent 32%),
            linear-gradient(135deg, #020705 0%, #07110C 50%, #020705 100%)
          `,
        }}
      />
      {grid ? <PerspectiveGrid /> : null}
      <ParticleField />
      {children}
      <HudFrame />
      <AbsoluteFill
        style={{
          pointerEvents: "none",
          background:
            "linear-gradient(90deg, rgba(2,7,5,.38), transparent 8%, transparent 92%, rgba(2,7,5,.38)), linear-gradient(rgba(2,7,5,.25), transparent 10%, transparent 90%, rgba(2,7,5,.35))",
        }}
      />
    </AbsoluteFill>
  );
};

const PerspectiveGrid: React.FC = () => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill style={{ opacity: 0.23, overflow: "hidden" }}>
      <div
        style={{
          position: "absolute",
          width: 2600,
          height: 1150,
          left: -340,
          top: 390,
          transform: `perspective(700px) rotateX(64deg) translateY(${(frame * 1.25) % 80}px)`,
          transformOrigin: "center top",
          backgroundImage:
            "linear-gradient(rgba(76,255,139,.13) 1px, transparent 1px), linear-gradient(90deg, rgba(76,255,139,.13) 1px, transparent 1px)",
          backgroundSize: "80px 80px",
          maskImage: "linear-gradient(to bottom, transparent, black 22%, black 72%, transparent)",
        }}
      />
      <div
        style={{
          position: "absolute",
          inset: 0,
          backgroundImage:
            "linear-gradient(90deg, transparent 49.9%, rgba(76,255,139,.05) 50%, transparent 50.1%)",
          backgroundSize: "240px 100%",
          transform: `translateX(${(frame * 0.3) % 240}px)`,
        }}
      />
    </AbsoluteFill>
  );
};

const ParticleField: React.FC = () => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill style={{ opacity: 0.48 }}>
      {Array.from({ length: 42 }, (_, index) => {
        const x = seeded(index + 4) * 1920;
        const y = (seeded(index + 47) * 1200 - frame * (0.22 + seeded(index) * 0.38) + 1200) % 1200;
        const size = 1 + seeded(index + 80) * 2.2;
        const alpha = 0.16 + seeded(index + 10) * 0.45;
        return (
          <div
            key={index}
            style={{
              position: "absolute",
              width: size,
              height: size,
              left: x,
              top: y - 60,
              borderRadius: "50%",
              background: index % 7 === 0 ? colors.cyan : colors.green,
              boxShadow: `0 0 ${size * 6}px currentColor`,
              opacity: alpha,
            }}
          />
        );
      })}
    </AbsoluteFill>
  );
};

const HudFrame: React.FC = () => {
  const frame = useCurrentFrame();
  const second = Math.floor(frame / 30);
  return (
    <AbsoluteFill
      style={{
        pointerEvents: "none",
        padding: 38,
        color: colors.dim,
        fontFamily: mono,
        fontSize: 14,
        letterSpacing: 2,
      }}
    >
      {[
        { left: 38, top: 38, borderLeftWidth: 1, borderTopWidth: 1 },
        { right: 38, top: 38, borderRightWidth: 1, borderTopWidth: 1 },
        { left: 38, bottom: 38, borderLeftWidth: 1, borderBottomWidth: 1 },
        { right: 38, bottom: 38, borderRightWidth: 1, borderBottomWidth: 1 },
      ].map((style, index) => (
        <div
          key={index}
          style={{
            position: "absolute",
            width: 34,
            height: 34,
            borderColor: "rgba(76,255,139,.24)",
            borderStyle: "solid",
            borderWidth: 0,
            ...style,
          }}
        />
      ))}
      <div style={{ position: "absolute", left: 58, top: 53 }}>
        SSHER / SIGNAL_0{(second % 9) + 1}
      </div>
      <div style={{ position: "absolute", right: 58, top: 53 }}>
        {String(Math.floor(frame / 30)).padStart(2, "0")}:
        {String(frame % 30).padStart(2, "0")}
      </div>
      <div
        style={{
          position: "absolute",
          left: 58,
          bottom: 53,
          color: "rgba(76,255,139,.5)",
        }}
      >
        ● SECURE CHANNEL
      </div>
      <div style={{ position: "absolute", right: 58, bottom: 53 }}>
        GETSSHER.COM
      </div>
    </AbsoluteFill>
  );
};

export const Wordmark: React.FC<{
  scale?: number;
  showVersion?: boolean;
}> = ({ scale = 1, showVersion = false }) => (
  <div style={{ display: "flex", alignItems: "center", gap: 18 * scale }}>
    <div
      style={{
        width: 58 * scale,
        height: 58 * scale,
        border: `${Math.max(1, 2 * scale)}px solid rgba(76,255,139,.55)`,
        borderRadius: 14 * scale,
        background: "rgba(76,255,139,.08)",
        color: colors.green,
        fontFamily: mono,
        fontWeight: 600,
        fontSize: 22 * scale,
        boxShadow: "inset 0 0 30px rgba(76,255,139,.08), 0 0 35px rgba(76,255,139,.08)",
        ...center,
      }}
    >
      &gt;_
    </div>
    <div style={{ display: "flex", alignItems: "baseline", gap: 12 * scale }}>
      <span
        style={{
          fontWeight: 800,
          fontSize: 46 * scale,
          letterSpacing: -2.8 * scale,
        }}
      >
        ssher
      </span>
      {showVersion ? (
        <span
          style={{
            color: colors.green,
            fontFamily: mono,
            fontSize: 13 * scale,
            letterSpacing: 1.5 * scale,
          }}
        >
          v0.1.2
        </span>
      ) : null}
    </div>
  </div>
);

export const ShieldMark: React.FC<{
  size?: number;
  progress?: number;
}> = ({ size = 360, progress = 1 }) => {
  const dash = 740;
  return (
    <div
      style={{
        width: size,
        height: size,
        position: "relative",
        ...center,
      }}
    >
      <div
        style={{
          position: "absolute",
          inset: 0,
          borderRadius: "50%",
          border: "1px dashed rgba(76,255,139,.25)",
          transform: `rotate(${progress * 70}deg) scale(${0.84 + progress * 0.16})`,
          opacity: progress,
          boxShadow: "0 0 100px rgba(76,255,139,.08)",
        }}
      />
      <div
        style={{
          position: "absolute",
          inset: size * 0.11,
          borderRadius: "50%",
          border: "1px solid rgba(93,231,255,.12)",
          transform: `rotate(${-progress * 50}deg)`,
        }}
      />
      <svg width={size * 0.62} height={size * 0.72} viewBox="0 0 240 280">
        <defs>
          <linearGradient id="shield-gradient" x1="0" x2="1" y1="0" y2="1">
            <stop offset="0" stopColor={colors.greenBright} />
            <stop offset="1" stopColor={colors.greenDark} />
          </linearGradient>
          <filter id="shield-glow">
            <feGaussianBlur stdDeviation="5" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
        </defs>
        <path
          d="M120 12 28 47v63c0 77 45 137 92 157 47-20 92-80 92-157V47L120 12Z"
          fill="rgba(76,255,139,.035)"
          stroke="url(#shield-gradient)"
          strokeWidth="6"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeDasharray={dash}
          strokeDashoffset={dash * (1 - progress)}
          filter="url(#shield-glow)"
        />
        <circle
          cx="120"
          cy="116"
          r="25"
          fill="none"
          stroke={colors.green}
          strokeWidth="7"
          opacity={progress}
        />
        <path
          d="M110 138v58h20v-58"
          fill="none"
          stroke={colors.green}
          strokeWidth="7"
          strokeLinecap="round"
          opacity={progress}
        />
        <path
          d="M145 90h26m-13-13v26M67 112h22m-11-11v22"
          stroke={colors.cyan}
          strokeWidth="4"
          strokeLinecap="round"
          opacity={progress * 0.75}
        />
      </svg>
    </div>
  );
};

export const Eyebrow: React.FC<{
  children: ReactNode;
  color?: string;
}> = ({ children, color = colors.green }) => (
  <div
    style={{
      display: "inline-flex",
      alignItems: "center",
      gap: 12,
      color,
      fontFamily: mono,
      fontSize: 18,
      fontWeight: 500,
      letterSpacing: 3.4,
      textTransform: "uppercase",
    }}
  >
    <span
      style={{
        width: 8,
        height: 8,
        borderRadius: "50%",
        background: color,
        boxShadow: `0 0 18px ${color}`,
      }}
    />
    {children}
  </div>
);

export const TerminalWindow: React.FC<{
  title?: string;
  command: string;
  lines?: Array<{ text: string; color?: string }>;
  typeFrom?: number;
  width?: number;
  scale?: number;
}> = ({
  title = "ssher — zsh",
  command,
  lines = [],
  typeFrom = 16,
  width = 1100,
  scale = 1,
}) => {
  const frame = useCurrentFrame();
  const visibleChars = Math.floor(clamp(frame, [typeFrom, typeFrom + 32], [0, command.length]));
  const typed = command.slice(0, visibleChars);
  const lineProgress = clamp(frame, [typeFrom + 28, typeFrom + 62], [0, lines.length + 0.4]);
  return (
    <div
      style={{
        width,
        border: "1px solid rgba(200,255,218,.19)",
        borderRadius: 18 * scale,
        overflow: "hidden",
        background: "rgba(3,10,7,.94)",
        boxShadow: shadow,
        fontFamily: mono,
        transformOrigin: "center center",
      }}
    >
      <div
        style={{
          height: 66 * scale,
          display: "grid",
          gridTemplateColumns: "1fr auto 1fr",
          alignItems: "center",
          padding: `0 ${22 * scale}px`,
          borderBottom: "1px solid rgba(200,255,218,.11)",
          color: colors.dim,
          fontSize: 14 * scale,
        }}
      >
        <div style={{ display: "flex", gap: 9 * scale }}>
          {[colors.red, colors.amber, colors.greenDark].map((color) => (
            <span
              key={color}
              style={{
                width: 12 * scale,
                height: 12 * scale,
                borderRadius: "50%",
                background: color,
              }}
            />
          ))}
        </div>
        <span>{title}</span>
        <span
          style={{
            justifySelf: "end",
            color: colors.green,
            fontSize: 12 * scale,
            letterSpacing: 1.5,
          }}
        >
          ENCRYPTED
        </span>
      </div>
      <div
        style={{
          minHeight: 300 * scale,
          padding: `${34 * scale}px ${38 * scale}px`,
          color: colors.mint,
          fontSize: 24 * scale,
          lineHeight: 1.8,
        }}
      >
        <div>
          <span style={{ color: colors.green }}>$ </span>
          {typed}
          <span
            style={{
              display: "inline-block",
              width: 12 * scale,
              height: 25 * scale,
              marginLeft: 5,
              verticalAlign: -4,
              background: colors.green,
              opacity: frame % 18 < 12 ? 1 : 0,
            }}
          />
        </div>
        <div style={{ marginTop: 18 * scale }}>
          {lines.map((line, index) => {
            const appear = ease(lineProgress, [index, index + 0.7]);
            return (
              <div
                key={`${line.text}-${index}`}
                style={{
                  color: line.color ?? colors.gray,
                  opacity: appear,
                  transform: `translateY(${(1 - appear) * 12}px)`,
                }}
              >
                {line.text}
              </div>
            );
          })}
        </div>
      </div>
      <div
        style={{
          height: 46 * scale,
          borderTop: "1px solid rgba(200,255,218,.1)",
          padding: `0 ${24 * scale}px`,
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          color: colors.dim,
          fontSize: 12 * scale,
        }}
      >
        <span>vault: unlocked</span>
        <span>known_hosts: verified</span>
      </div>
    </div>
  );
};

export const Pill: React.FC<{
  children: ReactNode;
  color?: string;
  active?: boolean;
}> = ({ children, color = colors.green, active = true }) => (
  <div
    style={{
      padding: "12px 18px",
      border: `1px solid ${color}${active ? "66" : "26"}`,
      borderRadius: 999,
      background: `${color}${active ? "12" : "05"}`,
      color: active ? color : colors.gray,
      fontFamily: mono,
      fontSize: 15,
      letterSpacing: 1.4,
      boxShadow: active ? `0 0 24px ${color}14` : "none",
    }}
  >
    {children}
  </div>
);

export const ServerNode: React.FC<{
  name: string;
  meta: string;
  x: number;
  y: number;
  progress: number;
  delay?: number;
}> = ({ name, meta, x, y, progress, delay = 0 }) => {
  const local = Math.max(0, Math.min(1, (progress - delay) / (1 - delay)));
  return (
    <div
      style={{
        position: "absolute",
        left: x,
        top: y,
        width: 245,
        height: 112,
        borderRadius: 14,
        border: "1px solid rgba(76,255,139,.35)",
        background: "rgba(6,18,12,.95)",
        padding: "21px 22px",
        boxShadow: `0 20px 55px rgba(0,0,0,.45), 0 0 ${local * 42}px rgba(76,255,139,.12)`,
        opacity: local,
        transform: `translateY(${(1 - local) * 22}px) scale(${0.88 + local * 0.12})`,
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        <span style={{ fontFamily: mono, fontSize: 18, color: colors.white }}>
          {name}
        </span>
        <span
          style={{
            width: 9,
            height: 9,
            borderRadius: "50%",
            background: colors.green,
            boxShadow: `0 0 15px ${colors.green}`,
          }}
        />
      </div>
      <div
        style={{
          fontFamily: mono,
          fontSize: 13,
          color: colors.gray,
          marginTop: 15,
        }}
      >
        {meta}
      </div>
    </div>
  );
};

export const GlitchText: React.FC<{
  text: string;
  style?: CSSProperties;
  intensity?: number;
}> = ({ text, style, intensity = 1 }) => {
  const frame = useCurrentFrame();
  const glitch = frame % 17 < 3;
  return (
    <div style={{ position: "relative", ...style }}>
      {text}
      {glitch ? (
        <>
          <div
            style={{
              position: "absolute",
              inset: 0,
              color: colors.cyan,
              clipPath: "inset(14% 0 62% 0)",
              transform: `translateX(${-7 * intensity}px)`,
              opacity: 0.7,
            }}
          >
            {text}
          </div>
          <div
            style={{
              position: "absolute",
              inset: 0,
              color: colors.green,
              clipPath: "inset(68% 0 8% 0)",
              transform: `translateX(${9 * intensity}px)`,
              opacity: 0.75,
            }}
          >
            {text}
          </div>
        </>
      ) : null}
    </div>
  );
};

export const DataBeam: React.FC<{
  x1: number;
  y1: number;
  x2: number;
  y2: number;
  progress: number;
  color?: string;
}> = ({ x1, y1, x2, y2, progress, color = colors.green }) => {
  const length = Math.hypot(x2 - x1, y2 - y1);
  const angle = (Math.atan2(y2 - y1, x2 - x1) * 180) / Math.PI;
  return (
    <div
      style={{
        position: "absolute",
        left: x1,
        top: y1,
        width: length,
        height: 2,
        transformOrigin: "left center",
        transform: `rotate(${angle}deg) scaleX(${progress})`,
        background: `linear-gradient(90deg, ${color}15, ${color}, ${color}15)`,
        boxShadow: `0 0 16px ${color}`,
        opacity: progress,
      }}
    />
  );
};

export const SceneLabel: React.FC<{
  number: string;
  children: ReactNode;
}> = ({ number, children }) => (
  <div
    style={{
      position: "absolute",
      left: 108,
      top: 94,
      display: "flex",
      alignItems: "center",
      gap: 16,
      color: colors.gray,
      fontFamily: mono,
      fontSize: 15,
      letterSpacing: 2.5,
      textTransform: "uppercase",
    }}
  >
    <span style={{ color: colors.green }}>{number}</span>
    <span style={{ width: 44, height: 1, background: colors.dim }} />
    {children}
  </div>
);

export const CircularMeter: React.FC<{
  progress: number;
  size?: number;
  label: string;
  value: string;
}> = ({ progress, size = 150, label, value }) => {
  const radius = size * 0.42;
  const circumference = 2 * Math.PI * radius;
  return (
    <div style={{ width: size, height: size, position: "relative", ...center }}>
      <svg width={size} height={size} style={{ position: "absolute" }}>
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="rgba(76,255,139,.02)"
          stroke="rgba(76,255,139,.12)"
          strokeWidth="5"
        />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke={colors.green}
          strokeWidth="5"
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={circumference * (1 - progress)}
          transform={`rotate(-90 ${size / 2} ${size / 2})`}
          style={{ filter: "drop-shadow(0 0 8px rgba(76,255,139,.55))" }}
        />
      </svg>
      <div style={{ textAlign: "center" }}>
        <div
          style={{
            fontFamily: mono,
            color: colors.white,
            fontSize: 22,
            fontWeight: 600,
          }}
        >
          {value}
        </div>
        <div
          style={{
            fontFamily: mono,
            color: colors.gray,
            fontSize: 11,
            marginTop: 5,
            letterSpacing: 1.4,
          }}
        >
          {label}
        </div>
      </div>
    </div>
  );
};

export const CameraFloat: React.FC<{
  children: ReactNode;
  amount?: number;
}> = ({ children, amount = 1 }) => {
  const frame = useCurrentFrame();
  const x = Math.sin(frame / 42) * 7 * amount;
  const y = Math.cos(frame / 55) * 5 * amount;
  return <div style={{ transform: `translate(${x}px, ${y}px)` }}>{children}</div>;
};

export const Reveal: React.FC<{
  children: ReactNode;
  from?: number;
  y?: number;
  style?: CSSProperties;
}> = ({ children, from = 0, y = 28, style }) => {
  const frame = useCurrentFrame();
  const value = ease(frame, [from, from + 20]);
  return (
    <div
      style={{
        opacity: value,
        transform: `translateY(${(1 - value) * y}px)`,
        ...style,
      }}
    >
      {children}
    </div>
  );
};

export const FullBleedFlash: React.FC<{
  at: number;
  color?: string;
  duration?: number;
}> = ({ at, color = colors.green, duration = 8 }) => {
  const frame = useCurrentFrame();
  const opacity = interpolate(
    frame,
    [at, at + 1, at + duration],
    [0, 0.75, 0],
    {
      extrapolateLeft: "clamp",
      extrapolateRight: "clamp",
    },
  );
  return (
    <AbsoluteFill
      style={{
        background: color,
        opacity,
        mixBlendMode: "screen",
        pointerEvents: "none",
      }}
    />
  );
};
