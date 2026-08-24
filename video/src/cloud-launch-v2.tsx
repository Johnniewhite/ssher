import React, { type CSSProperties, type ReactNode } from "react";
import {
  AbsoluteFill,
  Easing,
  interpolate,
  spring,
  useCurrentFrame,
  useVideoConfig,
} from "remotion";
import { mono, sans } from "./theme";

const palette = {
  black: "#0B0D0B",
  blackSoft: "#141714",
  white: "#F7F7F3",
  paper: "#EFEFE9",
  line: "#D8D8D1",
  gray: "#727872",
  dim: "#979D97",
  acid: "#D8FF4F",
  indigo: "#5B67F5",
  red: "#FF6D67",
};

const ease = (
  frame: number,
  input: [number, number],
  output: [number, number] = [0, 1],
) =>
  interpolate(frame, input, output, {
    easing: Easing.bezier(0.16, 1, 0.3, 1),
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });

const linear = (
  frame: number,
  input: [number, number],
  output: [number, number],
) =>
  interpolate(frame, input, output, {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });

const sceneOpacity = (
  frame: number,
  start: number,
  end: number,
  enter = 22,
  exit = 24,
) =>
  Math.min(
    ease(frame, [start, start + enter]),
    ease(frame, [end - exit, end], [1, 0]),
  );

const springAt = (frame: number, at: number, fps: number) =>
  spring({
    frame: frame - at,
    fps,
    config: { damping: 20, stiffness: 130, mass: 0.8 },
  });

const Brand: React.FC<{ inverse?: boolean; compact?: boolean }> = ({
  inverse = false,
  compact = false,
}) => (
  <div style={{ display: "flex", alignItems: "center", gap: compact ? 11 : 16 }}>
    <div
      style={{
        width: compact ? 38 : 52,
        height: compact ? 38 : 52,
        display: "grid",
        placeItems: "center",
        color: inverse ? palette.black : palette.white,
        background: inverse ? palette.white : palette.black,
        font: `${compact ? 700 : 650} ${compact ? 14 : 18}px ${mono}`,
      }}
    >
      &gt;_
    </div>
    <div
      style={{
        color: inverse ? palette.white : palette.black,
        fontFamily: sans,
        fontWeight: 800,
        fontSize: compact ? 24 : 32,
        letterSpacing: compact ? -1.1 : -1.7,
      }}
    >
      ssher<span style={{ color: palette.gray, fontWeight: 600 }}>cloud</span>
    </div>
  </div>
);

const TinyLabel: React.FC<{
  children: ReactNode;
  light?: boolean;
  style?: CSSProperties;
}> = ({ children, light = false, style }) => (
  <div
    style={{
      color: light ? "rgba(255,255,255,.54)" : palette.gray,
      font: `600 13px ${mono}`,
      letterSpacing: 2.4,
      textTransform: "uppercase",
      ...style,
    }}
  >
    {children}
  </div>
);

const SceneChrome: React.FC<{ light?: boolean; chapter: string }> = ({
  light = false,
  chapter,
}) => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill
      style={{
        pointerEvents: "none",
        color: light ? "rgba(255,255,255,.45)" : "rgba(11,13,11,.43)",
        font: `500 11px ${mono}`,
        letterSpacing: 2,
        textTransform: "uppercase",
      }}
    >
      <div style={{ position: "absolute", left: 38, top: 34 }}>{chapter}</div>
      <div style={{ position: "absolute", right: 38, top: 34 }}>
        {String(Math.floor(frame / 30)).padStart(2, "0")}:
        {String(frame % 30).padStart(2, "0")}
      </div>
      <div style={{ position: "absolute", left: 38, bottom: 34 }}>
        ssh access / reimagined
      </div>
      <div style={{ position: "absolute", right: 38, bottom: 34 }}>
        getssher.com
      </div>
    </AbsoluteFill>
  );
};

const Grain: React.FC<{ light?: boolean }> = ({ light = false }) => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill style={{ pointerEvents: "none", opacity: 0.16 }}>
      {Array.from({ length: 34 }, (_, index) => {
        const x = (index * 293 + frame * (index % 2 ? 0.31 : -0.17)) % 1920;
        const y = (index * 179 + frame * (index % 3 ? 0.11 : -0.23)) % 1080;
        return (
          <div
            key={index}
            style={{
              position: "absolute",
              left: x,
              top: y,
              width: index % 7 === 0 ? 2 : 1,
              height: index % 7 === 0 ? 2 : 1,
              background: light ? palette.white : palette.black,
              borderRadius: "50%",
            }}
          />
        );
      })}
    </AbsoluteFill>
  );
};

const OpeningScene: React.FC = () => {
  const frame = useCurrentFrame();
  const opacity = sceneOpacity(frame, 0, 150, 14, 24);
  const incident = springAt(frame, 14, 30);
  const lineOne = ease(frame, [28, 57]);
  const lineTwo = ease(frame, [50, 82]);
  const answer = ease(frame, [91, 120]);
  const sweep = linear(frame, [0, 148], [-200, 2120]);
  return (
    <AbsoluteFill
      style={{
        opacity,
        overflow: "hidden",
        color: palette.white,
        background: palette.black,
        fontFamily: sans,
      }}
    >
      <div
        style={{
          position: "absolute",
          inset: 0,
          background:
            "radial-gradient(circle at 78% 30%, rgba(91,103,245,.20), transparent 34%), radial-gradient(circle at 12% 100%, rgba(216,255,79,.08), transparent 30%)",
        }}
      />
      <div
        style={{
          position: "absolute",
          left: sweep,
          top: 0,
          width: 2,
          height: "100%",
          background: "rgba(255,255,255,.09)",
        }}
      />
      <div
        style={{
          position: "absolute",
          left: 142,
          top: 118,
          display: "flex",
          alignItems: "center",
          gap: 13,
          opacity: incident,
          transform: `translateY(${(1 - incident) * 18}px)`,
        }}
      >
        <span
          style={{
            width: 10,
            height: 10,
            background: palette.red,
            borderRadius: "50%",
            boxShadow: `0 0 0 7px rgba(255,109,103,.12)`,
          }}
        />
        <TinyLabel light>Production incident · 03:12</TinyLabel>
      </div>
      <div style={{ position: "absolute", left: 135, top: 245 }}>
        <div
          style={{
            fontSize: 112,
            lineHeight: 0.9,
            fontWeight: 800,
            letterSpacing: -8,
            opacity: lineOne,
            transform: `translateY(${(1 - lineOne) * 70}px)`,
          }}
        >
          Your servers don’t
        </div>
        <div
          style={{
            marginTop: 10,
            color: palette.acid,
            fontSize: 112,
            lineHeight: 0.9,
            fontWeight: 800,
            letterSpacing: -8,
            opacity: lineTwo,
            transform: `translateY(${(1 - lineTwo) * 70}px)`,
          }}
        >
          live at your desk.
        </div>
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 20,
            marginTop: 74,
            opacity: answer,
            transform: `translateX(${(1 - answer) * 42}px)`,
          }}
        >
          <div style={{ width: 76, height: 1, background: palette.white }} />
          <div
            style={{
              color: "rgba(255,255,255,.7)",
              fontSize: 31,
              fontWeight: 600,
              letterSpacing: -1,
            }}
          >
            Neither should your access.
          </div>
        </div>
      </div>
      <div
        style={{
          position: "absolute",
          right: 124,
          bottom: 105,
          width: 305,
          padding: "24px 26px",
          border: "1px solid rgba(255,255,255,.13)",
          background: "rgba(255,255,255,.045)",
          opacity: ease(frame, [66, 92]),
        }}
      >
        <TinyLabel light>api.service</TinyLabel>
        <div
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            marginTop: 15,
            font: `650 22px ${mono}`,
          }}
        >
          <span>502 / unhealthy</span>
          <span style={{ color: palette.red }}>●</span>
        </div>
      </div>
      <SceneChrome light chapter="01 / the constraint" />
      <Grain light />
    </AbsoluteFill>
  );
};

type DeviceProps = {
  type: "phone" | "tablet" | "laptop";
  x: number;
  y: number;
  scale: number;
  at: number;
  label: string;
};

const Device: React.FC<DeviceProps> = ({ type, x, y, scale, at, label }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const enter = springAt(frame, at, fps);
  const width = type === "phone" ? 300 : type === "tablet" ? 470 : 720;
  const height = type === "phone" ? 590 : type === "tablet" ? 600 : 455;
  const radius = type === "phone" ? 42 : type === "tablet" ? 27 : 18;
  const terminalLines = [
    ["deploy@prod", "~ $ ssher connect production-api"],
    ["status", "connected · session 7F2A"],
    ["deploy@prod", "~ $ uptime"],
    ["result", "up 41 days · load 0.42"],
  ];
  return (
    <div
      style={{
        position: "absolute",
        left: x,
        top: y,
        width,
        height,
        padding: type === "phone" ? 13 : 16,
        color: palette.white,
        background: palette.black,
        border: "1px solid rgba(255,255,255,.24)",
        borderRadius: radius,
        boxShadow: "0 42px 90px rgba(11,13,11,.22)",
        opacity: enter,
        transform: `translateY(${(1 - enter) * 90}px) rotate(${(1 - enter) * (type === "phone" ? 5 : -2)}deg) scale(${scale * (0.92 + enter * 0.08)})`,
        transformOrigin: "center center",
      }}
    >
      <div
        style={{
          height: "100%",
          overflow: "hidden",
          background: "#080A08",
          borderRadius: Math.max(8, radius - 14),
          border: "1px solid rgba(255,255,255,.08)",
        }}
      >
        <div
          style={{
            height: type === "phone" ? 58 : 52,
            padding: "0 18px",
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            borderBottom: "1px solid rgba(255,255,255,.08)",
            font: `600 ${type === "phone" ? 12 : 13}px ${mono}`,
          }}
        >
          <span>{label}</span>
          <span style={{ color: palette.acid }}>● live</span>
        </div>
        <div
          style={{
            padding: type === "phone" ? "25px 18px" : "27px 25px",
            font: `${type === "phone" ? 12 : 15}px/1.72 ${mono}`,
          }}
        >
          {terminalLines.map(([kind, text], index) => {
            const show = ease(frame, [at + 18 + index * 10, at + 30 + index * 10]);
            return (
              <div key={`${kind}-${text}`} style={{ opacity: show }}>
                <span
                  style={{
                    color:
                      kind === "result"
                        ? "#AAB1AA"
                        : kind === "status"
                          ? palette.acid
                          : palette.white,
                  }}
                >
                  {text}
                </span>
              </div>
            );
          })}
        </div>
      </div>
      {type === "phone" ? (
        <div
          style={{
            position: "absolute",
            left: "50%",
            top: 7,
            width: 78,
            height: 18,
            transform: "translateX(-50%)",
            background: palette.black,
            borderRadius: 20,
          }}
        />
      ) : null}
    </div>
  );
};

const EverywhereScene: React.FC = () => {
  const frame = useCurrentFrame();
  const opacity = sceneOpacity(frame, 120, 360);
  const title = ease(frame, [140, 180]);
  const session = ease(frame, [262, 298]);
  return (
    <AbsoluteFill
      style={{
        opacity,
        overflow: "hidden",
        color: palette.black,
        background: palette.white,
        fontFamily: sans,
      }}
    >
      <div
        style={{
          position: "absolute",
          inset: 0,
          background:
            "radial-gradient(circle at 92% 8%, rgba(91,103,245,.12), transparent 27%), radial-gradient(circle at 8% 95%, rgba(216,255,79,.24), transparent 25%)",
        }}
      />
      <div style={{ position: "absolute", left: 100, top: 138, width: 700 }}>
        <TinyLabel>One encrypted session</TinyLabel>
        <div
          style={{
            marginTop: 30,
            fontSize: 96,
            lineHeight: 0.93,
            fontWeight: 800,
            letterSpacing: -7,
            opacity: title,
            transform: `translateY(${(1 - title) * 55}px)`,
          }}
        >
          Every server.
          <br />
          <span style={{ color: palette.indigo }}>Any screen.</span>
        </div>
        <div
          style={{
            width: 510,
            marginTop: 34,
            color: palette.gray,
            fontSize: 24,
            lineHeight: 1.55,
            opacity: ease(frame, [174, 205]),
          }}
        >
          Start on your laptop. Continue on your phone. The shell stays alive in
          ssher Cloud.
        </div>
      </div>
      <div
        style={{
          position: "absolute",
          left: 864,
          top: 120,
          width: 930,
          height: 760,
        }}
      >
        <div
          style={{
            position: "absolute",
            left: 30,
            right: 40,
            top: 390,
            height: 1,
            background: `linear-gradient(90deg, transparent, ${palette.black}, transparent)`,
            opacity: 0.2,
          }}
        />
        <Device type="laptop" x={0} y={65} scale={0.79} at={165} label="MacBook · Lagos" />
        <Device type="tablet" x={380} y={125} scale={0.74} at={188} label="Browser · London" />
        <Device type="phone" x={670} y={90} scale={0.78} at={212} label="Mobile · Anywhere" />
      </div>
      <div
        style={{
          position: "absolute",
          left: 102,
          bottom: 98,
          display: "flex",
          alignItems: "center",
          gap: 13,
          padding: "15px 19px",
          color: palette.white,
          background: palette.black,
          opacity: session,
          transform: `translateY(${(1 - session) * 20}px)`,
        }}
      >
        <span style={{ color: palette.acid }}>●</span>
        <span style={{ font: `600 15px ${mono}` }}>session 7F2A · synchronized</span>
      </div>
      <SceneChrome chapter="02 / access from anywhere" />
      <Grain />
    </AbsoluteFill>
  );
};

const DashboardSidebar: React.FC = () => (
  <aside
    style={{
      width: 280,
      flex: "0 0 auto",
      padding: "27px 21px",
      color: palette.white,
      background: palette.black,
      fontFamily: sans,
    }}
  >
    <Brand inverse compact />
    <div
      style={{
        display: "grid",
        gridTemplateColumns: "38px 1fr",
        alignItems: "center",
        gap: 11,
        marginTop: 31,
        padding: "15px 4px",
        borderTop: "1px solid rgba(255,255,255,.17)",
        borderBottom: "1px solid rgba(255,255,255,.17)",
      }}
    >
      <span
        style={{
          width: 38,
          height: 38,
          display: "grid",
          placeItems: "center",
          color: palette.black,
          background: palette.white,
          fontWeight: 800,
        }}
      >
        A
      </span>
      <div>
        <strong style={{ display: "block", fontSize: 13 }}>Acme infrastructure</strong>
        <small style={{ color: "rgba(255,255,255,.45)", fontSize: 9, textTransform: "uppercase" }}>
          owner workspace
        </small>
      </div>
    </div>
    {[
      ["Workspace", "Command center"],
      ["", "Servers"],
      ["", "Private networks"],
      ["", "Terminals"],
      ["Access", "People"],
      ["", "Teams"],
      ["", "Activity"],
    ].map(([group, item]) => (
      <React.Fragment key={`${group}-${item}`}>
        {group ? (
          <div
            style={{
              margin: "21px 9px 7px",
              color: "rgba(255,255,255,.34)",
              font: `600 8px ${mono}`,
              letterSpacing: 1.5,
              textTransform: "uppercase",
            }}
          >
            {group}
          </div>
        ) : null}
        <div
          style={{
            height: 41,
            display: "flex",
            alignItems: "center",
            padding: "0 12px",
            color: item === "Servers" ? palette.black : "rgba(255,255,255,.58)",
            background: item === "Servers" ? palette.white : "transparent",
            fontSize: 12,
            fontWeight: 650,
          }}
        >
          <span style={{ width: 25 }}>{item === "Servers" ? "▤" : "·"}</span>
          {item}
        </div>
      </React.Fragment>
    ))}
    <div
      style={{
        position: "absolute",
        left: 21,
        bottom: 26,
        width: 238,
        padding: "13px 11px",
        border: "1px solid rgba(255,255,255,.15)",
        color: "rgba(255,255,255,.75)",
        fontSize: 10,
      }}
    >
      <span style={{ marginRight: 8 }}>✦</span> Zero-knowledge cloud
    </div>
  </aside>
);

const WorkspaceScene: React.FC = () => {
  const frame = useCurrentFrame();
  const opacity = sceneOpacity(frame, 330, 630);
  const windowIn = springAt(frame, 355, 30);
  const cards = [
    { name: "production-api", host: "23.20.159.49", team: "Platform", live: true },
    { name: "payments-db", host: "10.0.18.42", team: "Data", live: true },
    { name: "edge-cache", host: "172.31.8.11", team: "Network", live: false },
  ];
  const sync = ease(frame, [505, 535]) * ease(frame, [600, 626], [1, 0]);
  return (
    <AbsoluteFill
      style={{
        opacity,
        overflow: "hidden",
        color: palette.black,
        background: palette.paper,
        fontFamily: sans,
      }}
    >
      <div style={{ position: "absolute", left: 94, top: 87 }}>
        <TinyLabel>One command from local to cloud</TinyLabel>
        <div
          style={{
            marginTop: 15,
            fontSize: 44,
            fontWeight: 800,
            letterSpacing: -2.4,
          }}
        >
          Your infrastructure, organized.
        </div>
      </div>
      <div
        style={{
          position: "absolute",
          left: 90,
          right: 90,
          top: 175,
          height: 800,
          display: "flex",
          overflow: "hidden",
          background: palette.white,
          border: `1px solid ${palette.line}`,
          boxShadow: "0 46px 100px rgba(11,13,11,.14)",
          opacity: windowIn,
          transform: `translateY(${(1 - windowIn) * 80}px) scale(${0.96 + windowIn * 0.04})`,
        }}
      >
        <DashboardSidebar />
        <main style={{ flex: 1, minWidth: 0, background: "#F7F7F3" }}>
          <div
            style={{
              height: 128,
              padding: "28px 40px",
              display: "flex",
              alignItems: "flex-end",
              justifyContent: "space-between",
              background: palette.white,
              borderBottom: `1px solid ${palette.line}`,
            }}
          >
            <div>
              <TinyLabel>Infrastructure</TinyLabel>
              <div style={{ marginTop: 8, fontSize: 34, fontWeight: 800, letterSpacing: -1.8 }}>
                Server inventory
              </div>
            </div>
            <div
              style={{
                padding: "13px 19px",
                color: palette.white,
                background: palette.black,
                borderRadius: 99,
                fontSize: 12,
                fontWeight: 700,
              }}
            >
              + Add server
            </div>
          </div>
          <div style={{ padding: 38 }}>
            <div
              style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                marginBottom: 23,
              }}
            >
              <div>
                <div style={{ fontSize: 25, fontWeight: 800, letterSpacing: -1.1 }}>
                  Your team’s servers
                </div>
                <div style={{ color: palette.gray, fontSize: 12, marginTop: 5 }}>
                  Credentials stay encrypted until a terminal opens.
                </div>
              </div>
              <div
                style={{
                  padding: "8px 12px",
                  color: palette.white,
                  background: palette.black,
                  borderRadius: 99,
                  font: `600 10px ${mono}`,
                }}
              >
                3 encrypted
              </div>
            </div>
            <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 16 }}>
              {cards.map((card, index) => {
                const enter = springAt(frame, 405 + index * 18, 30);
                return (
                  <div
                    key={card.name}
                    style={{
                      minHeight: 330,
                      padding: 22,
                      background: palette.white,
                      border: `1px solid ${palette.line}`,
                      opacity: enter,
                      transform: `translateY(${(1 - enter) * 38}px)`,
                    }}
                  >
                    <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
                      <span
                        style={{
                          width: 42,
                          height: 42,
                          display: "grid",
                          placeItems: "center",
                          color: palette.white,
                          background: palette.black,
                        }}
                      >
                        ▤
                      </span>
                      <div style={{ minWidth: 0 }}>
                        <strong style={{ display: "block", fontSize: 15 }}>{card.name}</strong>
                        <small style={{ color: palette.gray, font: `10px ${mono}` }}>{card.host}</small>
                      </div>
                    </div>
                    <div
                      style={{
                        display: "grid",
                        gap: 9,
                        marginTop: 25,
                        padding: 15,
                        background: "#F0F0EB",
                        color: palette.gray,
                        fontSize: 10,
                      }}
                    >
                      <span>USER&nbsp;&nbsp; deploy</span>
                      <span>TEAM&nbsp;&nbsp; {card.team}</span>
                      <span>AUTH&nbsp;&nbsp; AES-256-GCM</span>
                    </div>
                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                        alignItems: "center",
                        marginTop: 25,
                      }}
                    >
                      <span style={{ color: card.live ? "#2D673A" : palette.gray, fontSize: 10 }}>
                        ● {card.live ? "reachable" : "private network"}
                      </span>
                      <span
                        style={{
                          padding: "9px 13px",
                          color: palette.white,
                          background: palette.black,
                          borderRadius: 99,
                          fontSize: 10,
                          fontWeight: 700,
                        }}
                      >
                        Open terminal
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </main>
      </div>
      <div
        style={{
          position: "absolute",
          right: 122,
          bottom: 75,
          display: "flex",
          alignItems: "center",
          gap: 13,
          padding: "17px 20px",
          color: palette.white,
          background: palette.black,
          boxShadow: "0 20px 50px rgba(11,13,11,.28)",
          opacity: sync,
          transform: `translateY(${(1 - sync) * 26}px)`,
        }}
      >
        <span style={{ color: palette.acid }}>✓</span>
        <div>
          <strong style={{ display: "block", fontSize: 12 }}>3 servers pushed from the CLI</strong>
          <small style={{ color: "rgba(255,255,255,.48)", font: `9px ${mono}` }}>
            encrypted before upload
          </small>
        </div>
      </div>
      <SceneChrome chapter="03 / the workspace" />
    </AbsoluteFill>
  );
};

const typed = (text: string, frame: number, start: number, end: number) =>
  text.slice(0, Math.floor(linear(frame, [start, end], [0, text.length])));

const TerminalScene: React.FC = () => {
  const frame = useCurrentFrame();
  const opacity = sceneOpacity(frame, 600, 960);
  const windowIn = springAt(frame, 620, 30);
  const prompt = typed("Find why checkout is returning 502s.", frame, 655, 716);
  const proposal = ease(frame, [718, 755]);
  const approve = ease(frame, [787, 812]);
  const run = ease(frame, [818, 846]);
  const recovered = ease(frame, [860, 890]);
  return (
    <AbsoluteFill
      style={{
        opacity,
        overflow: "hidden",
        color: palette.white,
        background: palette.black,
        fontFamily: sans,
      }}
    >
      <div
        style={{
          position: "absolute",
          inset: 0,
          background:
            "radial-gradient(circle at 82% 14%, rgba(91,103,245,.22), transparent 30%), radial-gradient(circle at 6% 92%, rgba(216,255,79,.07), transparent 26%)",
        }}
      />
      <div style={{ position: "absolute", left: 86, top: 88 }}>
        <TinyLabel light>Browser terminal + AI operator</TinyLabel>
        <div style={{ marginTop: 13, fontSize: 42, fontWeight: 800, letterSpacing: -2 }}>
          Diagnose. Review. Recover.
        </div>
      </div>
      <div
        style={{
          position: "absolute",
          left: 78,
          right: 78,
          top: 175,
          height: 790,
          display: "grid",
          gridTemplateColumns: "minmax(0, 1fr) 490px",
          overflow: "hidden",
          background: "#070907",
          border: "1px solid rgba(255,255,255,.16)",
          boxShadow: "0 50px 120px rgba(0,0,0,.48)",
          opacity: windowIn,
          transform: `translateY(${(1 - windowIn) * 65}px) scale(${0.97 + windowIn * 0.03})`,
        }}
      >
        <section style={{ display: "grid", gridTemplateRows: "68px 1fr 92px", minWidth: 0 }}>
          <header
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              padding: "0 25px",
              background: "#111411",
              borderBottom: "1px solid rgba(255,255,255,.1)",
            }}
          >
            <div>
              <strong style={{ fontSize: 14 }}>production-api</strong>
              <span style={{ marginLeft: 14, color: "rgba(255,255,255,.38)", font: `10px ${mono}` }}>
                deploy@23.20.159.49
              </span>
            </div>
            <span style={{ color: palette.acid, font: `10px ${mono}` }}>● connected · session 7F2A</span>
          </header>
          <div style={{ padding: "35px 38px", font: `15px/1.75 ${mono}` }}>
            <div style={{ color: "rgba(255,255,255,.38)" }}>Ubuntu 24.04.3 LTS · api-2</div>
            <div style={{ marginTop: 27 }}>
              <span style={{ color: palette.acid }}>deploy@prod</span>:~$ systemctl --failed
            </div>
            <div style={{ color: palette.red, opacity: ease(frame, [645, 665]) }}>
              ● api.service&nbsp;&nbsp;&nbsp; failed&nbsp;&nbsp;&nbsp; exit-code
            </div>
            <div style={{ marginTop: 27, opacity: run }}>
              <span style={{ color: palette.acid }}>operator</span>:~$ journalctl -u api --since "10 min ago" --no-pager
            </div>
            <div style={{ color: "rgba(255,255,255,.52)", opacity: run }}>
              Error: address already in use :8080
            </div>
            <div style={{ opacity: run }}>
              <span style={{ color: palette.acid }}>operator</span>:~$ fuser -k 8080/tcp && systemctl restart api
            </div>
            <div style={{ marginTop: 25, color: palette.acid, opacity: recovered }}>
              ✓ api.service active (running)
            </div>
            <div style={{ color: palette.white, opacity: recovered }}>
              ✓ GET /health&nbsp;&nbsp;200 OK&nbsp;&nbsp;42ms
            </div>
          </div>
          <footer
            style={{
              padding: "16px 22px",
              display: "flex",
              alignItems: "center",
              gap: 12,
              background: "#111411",
              borderTop: "1px solid rgba(255,255,255,.1)",
            }}
          >
            <div
              style={{
                flex: 1,
                height: 55,
                padding: "0 18px",
                display: "flex",
                alignItems: "center",
                color: palette.white,
                background: "rgba(255,255,255,.055)",
                border: "1px solid rgba(255,255,255,.12)",
                fontSize: 13,
              }}
            >
              {prompt}
              <span style={{ opacity: Math.floor(frame / 10) % 2, color: palette.acid }}>▌</span>
            </div>
            <div style={{ width: 55, height: 55, display: "grid", placeItems: "center", color: palette.black, background: palette.acid }}>
              ↑
            </div>
          </footer>
        </section>
        <aside
          style={{
            minWidth: 0,
            padding: 23,
            background: "#111411",
            borderLeft: "1px solid rgba(255,255,255,.1)",
          }}
        >
          <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
            <span style={{ width: 40, height: 40, display: "grid", placeItems: "center", color: palette.black, background: palette.acid }}>
              ✦
            </span>
            <div>
              <strong style={{ display: "block", fontSize: 14 }}>AI operator</strong>
              <small style={{ color: "rgba(255,255,255,.42)", fontSize: 10 }}>Proposes commands. You approve.</small>
            </div>
          </div>
          <div
            style={{
              marginTop: 25,
              padding: 18,
              color: "rgba(255,255,255,.72)",
              background: "rgba(255,255,255,.045)",
              border: "1px solid rgba(255,255,255,.1)",
              fontSize: 12,
              lineHeight: 1.55,
              opacity: ease(frame, [688, 716]),
            }}
          >
            I’ll inspect the failed unit, identify the process holding port 8080, then restart and verify the health endpoint.
          </div>
          <div
            style={{
              marginTop: 16,
              padding: 18,
              background: "#080A08",
              border: "1px solid rgba(255,255,255,.13)",
              opacity: proposal,
              transform: `translateY(${(1 - proposal) * 24}px)`,
            }}
          >
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
              <TinyLabel light>Command plan</TinyLabel>
              <span style={{ padding: "6px 9px", color: "#3D310C", background: "#E6C976", font: `700 8px ${mono}` }}>
                MUTATING
              </span>
            </div>
            {["journalctl -u api --since '10 min ago'", "fuser -k 8080/tcp", "systemctl restart api", "curl -fsS localhost:8080/health"].map((command, index) => (
              <div
                key={command}
                style={{
                  marginTop: index === 0 ? 18 : 10,
                  padding: "10px 11px",
                  color: "rgba(255,255,255,.72)",
                  background: "rgba(255,255,255,.045)",
                  font: `9px ${mono}`,
                }}
              >
                {index + 1}. {command}
              </div>
            ))}
            <div
              style={{
                marginTop: 18,
                height: 44,
                display: "grid",
                placeItems: "center",
                color: palette.black,
                background: palette.acid,
                fontSize: 11,
                fontWeight: 800,
                transform: `scale(${0.96 + approve * 0.04})`,
                boxShadow: approve ? "0 0 0 4px rgba(216,255,79,.12)" : "none",
              }}
            >
              {approve ? "Approved · running" : "Review and run"}
            </div>
          </div>
          <div
            style={{
              marginTop: 16,
              padding: "15px 17px",
              display: "flex",
              alignItems: "center",
              gap: 10,
              color: palette.black,
              background: palette.white,
              opacity: recovered,
            }}
          >
            <span style={{ color: "#2D673A", fontSize: 18 }}>✓</span>
            <strong style={{ fontSize: 11 }}>Incident recovered in 42 seconds</strong>
          </div>
        </aside>
      </div>
      <SceneChrome light chapter="04 / operate with confidence" />
      <Grain light />
    </AbsoluteFill>
  );
};

const PersistenceScene: React.FC = () => {
  const frame = useCurrentFrame();
  const opacity = sceneOpacity(frame, 930, 1140);
  const first = springAt(frame, 952, 30);
  const second = springAt(frame, 1015, 30);
  const handoff = ease(frame, [1000, 1055]);
  const copy = ease(frame, [1030, 1070]);
  return (
    <AbsoluteFill
      style={{
        opacity,
        overflow: "hidden",
        color: palette.black,
        background: palette.white,
        fontFamily: sans,
      }}
    >
      <div
        style={{
          position: "absolute",
          inset: 0,
          background: "radial-gradient(circle at 85% 20%, rgba(91,103,245,.11), transparent 29%)",
        }}
      />
      <div style={{ position: "absolute", left: 110, top: 125, width: 715 }}>
        <TinyLabel>Persistent cloud sessions</TinyLabel>
        <div
          style={{
            marginTop: 27,
            fontSize: 88,
            lineHeight: 0.94,
            fontWeight: 800,
            letterSpacing: -6.5,
            opacity: copy,
            transform: `translateY(${(1 - copy) * 45}px)`,
          }}
        >
          Close the laptop.
          <br />
          <span style={{ color: palette.indigo }}>Not the session.</span>
        </div>
        <div style={{ width: 540, marginTop: 33, color: palette.gray, fontSize: 22, lineHeight: 1.55, opacity: copy }}>
          Detach now. Resume the exact shell later—from any authorized device.
        </div>
      </div>
      <div
        style={{
          position: "absolute",
          right: 85,
          top: 118,
          width: 930,
          height: 790,
        }}
      >
        <div
          style={{
            position: "absolute",
            left: 20,
            top: 85,
            width: 640,
            height: 425,
            padding: 13,
            color: palette.white,
            background: palette.black,
            border: "1px solid rgba(255,255,255,.16)",
            boxShadow: "0 38px 90px rgba(11,13,11,.2)",
            opacity: first * (1 - handoff * 0.42),
            transform: `translateX(${-handoff * 90}px) translateY(${(1 - first) * 70}px) scale(${0.94 + first * 0.06})`,
          }}
        >
          <div style={{ height: 47, padding: "0 14px", display: "flex", alignItems: "center", justifyContent: "space-between", borderBottom: "1px solid rgba(255,255,255,.09)", font: `10px ${mono}` }}>
            <span>production-api</span><span style={{ color: palette.acid }}>● attached</span>
          </div>
          <div style={{ padding: 25, font: `14px/1.75 ${mono}` }}>
            <div><span style={{ color: palette.acid }}>deploy@prod</span>:~$ tail -f /var/log/api.log</div>
            <div style={{ color: "rgba(255,255,255,.45)" }}>200 GET /health 42ms</div>
            <div style={{ color: "rgba(255,255,255,.45)" }}>200 POST /checkout 118ms</div>
          </div>
          <div style={{ position: "absolute", right: 18, bottom: 18, padding: "10px 14px", color: palette.black, background: palette.acid, fontSize: 10, fontWeight: 800 }}>
            Detach
          </div>
        </div>
        <div
          style={{
            position: "absolute",
            right: 5,
            bottom: 20,
            width: 325,
            height: 620,
            padding: 13,
            color: palette.white,
            background: palette.black,
            border: "1px solid rgba(255,255,255,.18)",
            borderRadius: 43,
            boxShadow: "0 42px 100px rgba(11,13,11,.25)",
            opacity: second,
            transform: `translateY(${(1 - second) * 100}px) scale(${0.92 + second * 0.08})`,
          }}
        >
          <div style={{ height: "100%", padding: "75px 18px 22px", background: "#080A08", borderRadius: 31 }}>
            <TinyLabel light>Resume session</TinyLabel>
            <div style={{ marginTop: 18, fontSize: 21, fontWeight: 800 }}>production-api</div>
            <div style={{ color: "rgba(255,255,255,.42)", font: `10px ${mono}`, marginTop: 5 }}>session 7F2A · 2m ago</div>
            <div style={{ marginTop: 30, padding: 17, background: "rgba(255,255,255,.05)", font: `11px/1.6 ${mono}` }}>
              <span style={{ color: palette.acid }}>deploy@prod</span>:~$ tail -f /var/log/api.log
              <br /><span style={{ color: "rgba(255,255,255,.45)" }}>200 POST /checkout 118ms</span>
            </div>
            <div style={{ position: "absolute", left: 31, right: 31, bottom: 35, height: 49, display: "grid", placeItems: "center", color: palette.black, background: palette.acid, fontSize: 11, fontWeight: 800 }}>
              Resume on this device
            </div>
          </div>
          <div style={{ position: "absolute", top: 8, left: "50%", width: 82, height: 20, transform: "translateX(-50%)", background: palette.black, borderRadius: 20 }} />
        </div>
        <div
          style={{
            position: "absolute",
            left: 545,
            top: 520,
            width: 145,
            height: 2,
            background: palette.black,
            opacity: handoff,
            transform: `scaleX(${handoff})`,
            transformOrigin: "left",
          }}
        >
          <span style={{ position: "absolute", right: -3, top: -4, width: 10, height: 10, background: palette.black, borderRadius: "50%" }} />
        </div>
      </div>
      <SceneChrome chapter="05 / leave it running" />
    </AbsoluteFill>
  );
};

const TrustScene: React.FC = () => {
  const frame = useCurrentFrame();
  const opacity = sceneOpacity(frame, 1110, 1260);
  const cards = [
    { number: "01", title: "Zero-knowledge", body: "Secrets decrypt only on authorized devices." },
    { number: "02", title: "Team-controlled", body: "Give the right people the right infrastructure." },
    { number: "03", title: "Fully accountable", body: "Every access and infrastructure change stays visible." },
  ];
  return (
    <AbsoluteFill
      style={{
        opacity,
        overflow: "hidden",
        color: palette.white,
        background: palette.black,
        fontFamily: sans,
      }}
    >
      <div style={{ position: "absolute", left: 105, top: 110 }}>
        <TinyLabel light>Built for serious infrastructure</TinyLabel>
        <div style={{ marginTop: 22, fontSize: 68, lineHeight: 0.98, fontWeight: 800, letterSpacing: -4.3 }}>
          Fast access.
          <br /><span style={{ color: palette.acid }}>Without giving up control.</span>
        </div>
      </div>
      <div style={{ position: "absolute", left: 105, right: 105, top: 420, display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 18 }}>
        {cards.map((card, index) => {
          const enter = springAt(frame, 1150 + index * 18, 30);
          const dark = index === 1;
          return (
            <div
              key={card.number}
              style={{
                minHeight: 390,
                padding: 34,
                color: dark ? palette.black : palette.white,
                background: dark ? palette.acid : "rgba(255,255,255,.055)",
                border: dark ? "none" : "1px solid rgba(255,255,255,.13)",
                opacity: enter,
                transform: `translateY(${(1 - enter) * 65}px)`,
              }}
            >
              <TinyLabel light={!dark}>{card.number}</TinyLabel>
              <div style={{ marginTop: 150, fontSize: 35, fontWeight: 800, letterSpacing: -1.8 }}>{card.title}</div>
              <div style={{ marginTop: 17, maxWidth: 350, color: dark ? "rgba(11,13,11,.62)" : "rgba(255,255,255,.52)", fontSize: 17, lineHeight: 1.5 }}>
                {card.body}
              </div>
            </div>
          );
        })}
      </div>
      <SceneChrome light chapter="06 / trust by design" />
      <Grain light />
    </AbsoluteFill>
  );
};

const FinaleScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const opacity = ease(frame, [1228, 1252]);
  const reveal = springAt(frame, 1245, fps);
  const line = ease(frame, [1280, 1325]);
  return (
    <AbsoluteFill
      style={{
        opacity,
        overflow: "hidden",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        color: palette.black,
        background: palette.white,
        fontFamily: sans,
        textAlign: "center",
      }}
    >
      <div
        style={{
          position: "absolute",
          width: 980,
          height: 980,
          border: `1px solid ${palette.line}`,
          borderRadius: "50%",
          transform: `scale(${0.72 + reveal * 0.28})`,
          opacity: 0.72,
        }}
      />
      <div
        style={{
          position: "absolute",
          width: 680,
          height: 680,
          border: `1px solid ${palette.line}`,
          borderRadius: "50%",
          transform: `scale(${0.68 + reveal * 0.32})`,
        }}
      />
      <div style={{ position: "relative", opacity: reveal, transform: `translateY(${(1 - reveal) * 45}px)` }}>
        <div style={{ display: "flex", justifyContent: "center" }}><Brand /></div>
        <div style={{ marginTop: 52, fontSize: 77, lineHeight: 0.98, fontWeight: 800, letterSpacing: -5.4 }}>
          Every server. Any screen.
          <br /><span style={{ color: palette.indigo }}>Exactly where you left it.</span>
        </div>
        <div style={{ marginTop: 31, color: palette.gray, fontSize: 20 }}>
          Browser SSH · persistent sessions · team access · AI operator
        </div>
        <div
          style={{
            display: "inline-flex",
            marginTop: 46,
            padding: "18px 27px",
            color: palette.white,
            background: palette.black,
            borderRadius: 99,
            fontSize: 16,
            fontWeight: 800,
          }}
        >
          Open cloud.getssher.com&nbsp;&nbsp;→
        </div>
      </div>
      <div
        style={{
          position: "absolute",
          left: 0,
          bottom: 0,
          width: `${line * 100}%`,
          height: 8,
          background: palette.acid,
        }}
      />
      <SceneChrome chapter="07 / ssher cloud" />
    </AbsoluteFill>
  );
};

export const SsherCloudLaunchV2: React.FC = () => (
  <AbsoluteFill style={{ background: palette.black }}>
    <OpeningScene />
    <EverywhereScene />
    <WorkspaceScene />
    <TerminalScene />
    <PersistenceScene />
    <TrustScene />
    <FinaleScene />
  </AbsoluteFill>
);
