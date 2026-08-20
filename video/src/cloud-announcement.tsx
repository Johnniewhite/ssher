import React from "react";
import {
  AbsoluteFill,
  Easing,
  interpolate,
  spring,
  useCurrentFrame,
  useVideoConfig,
} from "remotion";
import { mono, sans } from "./theme";

const ink = "#111416";
const blue = "#0A84FF";
const green = "#44F08A";
const night = "#06110C";

const clamp = (
  frame: number,
  input: [number, number],
  output: [number, number],
) =>
  interpolate(frame, input, output, {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
    easing: Easing.bezier(0.16, 1, 0.3, 1),
  });

const Brand: React.FC<{ light?: boolean }> = ({ light = false }) => (
  <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
    <div
      style={{
        width: 52,
        height: 52,
        borderRadius: 15,
        display: "grid",
        placeItems: "center",
        background: light ? "rgba(68,240,138,.14)" : green,
        border: light ? "1px solid rgba(68,240,138,.34)" : "none",
        color: light ? green : night,
        fontFamily: mono,
        fontSize: 19,
        fontWeight: 700,
      }}
    >
      &gt;_
    </div>
    <div
      style={{
        color: light ? "#F5FFF8" : ink,
        fontFamily: sans,
        lineHeight: 1,
      }}
    >
      <div style={{ fontSize: 28, fontWeight: 800, letterSpacing: -1.3 }}>
        ssher
      </div>
      <div
        style={{
          marginTop: 6,
          color: light ? "#8DA398" : "#72777A",
          fontSize: 11,
          fontWeight: 700,
          letterSpacing: 2.4,
          textTransform: "uppercase",
        }}
      >
        Cloud
      </div>
    </div>
  </div>
);

type BubbleProps = {
  at: number;
  side: "left" | "right";
  children: React.ReactNode;
  width?: number;
  dark?: boolean;
};

const Bubble: React.FC<BubbleProps> = ({
  at,
  side,
  children,
  width = 520,
  dark = false,
}) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const enter = spring({
    frame: frame - at,
    fps,
    config: { damping: 18, stiffness: 180 },
  });
  const opacity = clamp(frame, [at, at + 8], [0, 1]);
  return (
    <div
      style={{
        alignSelf: side === "right" ? "flex-end" : "flex-start",
        maxWidth: width,
        padding: "22px 28px",
        borderRadius:
          side === "right" ? "29px 29px 8px 29px" : "29px 29px 29px 8px",
        background: side === "right" ? blue : dark ? "#17231D" : "#E8E8ED",
        color: side === "right" || dark ? "white" : ink,
        fontFamily: sans,
        fontSize: 30,
        fontWeight: 500,
        lineHeight: 1.28,
        letterSpacing: -0.65,
        opacity,
        transform: `translateY(${(1 - enter) * 28}px) scale(${0.92 + enter * 0.08})`,
        transformOrigin: side === "right" ? "100% 100%" : "0% 100%",
        boxShadow:
          side === "right"
            ? "0 14px 35px rgba(10,132,255,.2)"
            : "0 10px 30px rgba(17,20,22,.08)",
      }}
    >
      {children}
    </div>
  );
};

const ChatHeader: React.FC<{ dark?: boolean }> = ({ dark = false }) => (
  <div
    style={{
      display: "flex",
      alignItems: "center",
      justifyContent: "space-between",
      height: 124,
      padding: "0 42px",
      borderBottom: `1px solid ${dark ? "rgba(255,255,255,.08)" : "rgba(17,20,22,.09)"}`,
    }}
  >
    <div style={{ color: blue, fontFamily: sans, fontSize: 26 }}>
      ‹ Messages
    </div>
    <div
      style={{
        textAlign: "center",
        color: dark ? "white" : ink,
        fontFamily: sans,
      }}
    >
      <div style={{ fontSize: 28, fontWeight: 750, letterSpacing: -0.6 }}>
        Ops room
      </div>
      <div
        style={{
          color: dark ? "#87958D" : "#7B7F82",
          fontSize: 17,
          marginTop: 4,
        }}
      >
        4 people ›
      </div>
    </div>
    <div
      style={{
        width: 46,
        height: 46,
        borderRadius: 23,
        display: "grid",
        placeItems: "center",
        color: blue,
        background: dark ? "rgba(10,132,255,.14)" : "#E9F3FF",
        fontFamily: sans,
        fontSize: 24,
        fontWeight: 800,
      }}
    >
      i
    </div>
  </div>
);

const ChatScene: React.FC = () => {
  const frame = useCurrentFrame();
  const visible =
    clamp(frame, [0, 18], [0, 1]) * clamp(frame, [285, 315], [1, 0]);
  const lift = clamp(frame, [0, 285], [25, -12]);
  return (
    <AbsoluteFill
      style={{
        background:
          "radial-gradient(circle at 50% -20%, #FFFFFF 0%, #F2F3F5 55%, #E7E8EB 100%)",
        opacity: visible,
      }}
    >
      <div
        style={{
          position: "absolute",
          left: 250,
          top: 50 + lift,
          width: 1420,
          height: 980,
          borderRadius: 45,
          overflow: "hidden",
          background: "rgba(255,255,255,.94)",
          border: "1px solid rgba(17,20,22,.08)",
          boxShadow: "0 45px 120px rgba(28,31,34,.16)",
        }}
      >
        <ChatHeader />
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            gap: 22,
            padding: "48px 74px",
          }}
        >
          <div
            style={{
              color: "#8A8E91",
              fontFamily: sans,
              fontSize: 16,
              textAlign: "center",
              marginBottom: 6,
            }}
          >
            Today 09:41
          </div>
          <Bubble at={38} side="left">
            Prod needs a restart. Anyone near a laptop?
          </Bubble>
          <Bubble at={102} side="right" width={430}>
            I’m on my phone 😬
          </Bubble>
          <Bubble at={166} side="left" width={550}>
            That’s fine. Open ssher Cloud.
          </Bubble>
          <Bubble at={226} side="left" width={640}>
            <div style={{ display: "flex", alignItems: "center", gap: 20 }}>
              <Brand />
              <div
                style={{
                  borderLeft: "1px solid rgba(17,20,22,.12)",
                  paddingLeft: 22,
                }}
              >
                <div style={{ fontSize: 22, fontWeight: 800 }}>
                  Your workspace is ready
                </div>
                <div style={{ color: "#707578", fontSize: 18, marginTop: 5 }}>
                  cloud.getssher.com
                </div>
              </div>
            </div>
          </Bubble>
        </div>
      </div>
      <div
        style={{
          position: "absolute",
          top: 38,
          left: 62,
          fontFamily: mono,
          color: "#8B8F92",
          fontSize: 16,
          letterSpacing: 3,
          textTransform: "uppercase",
        }}
      >
        A small emergency
      </div>
    </AbsoluteFill>
  );
};

const TerminalScene: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const at = 295;
  const enter = spring({
    frame: frame - at,
    fps,
    config: { damping: 20, stiffness: 120 },
  });
  const exit = clamp(frame, [675, 710], [1, 0]);
  const commandText = "sudo systemctl restart api";
  const typed = Math.floor(clamp(frame, [425, 500], [0, commandText.length]));
  const command = commandText.slice(0, typed);
  const cursor = Math.floor(frame / 12) % 2 === 0;
  return (
    <AbsoluteFill
      style={{
        background:
          "radial-gradient(circle at 75% 10%, rgba(68,240,138,.14), transparent 34%), #06110C",
        opacity: enter * exit,
        transform: `scale(${0.96 + enter * 0.04})`,
      }}
    >
      <div style={{ position: "absolute", top: 58, left: 70 }}>
        <Brand light />
      </div>
      <div
        style={{
          position: "absolute",
          top: 61,
          right: 72,
          color: "#8DA398",
          fontFamily: mono,
          fontSize: 17,
          letterSpacing: 2.5,
          textTransform: "uppercase",
        }}
      >
        Browser SSH · encrypted in transit
      </div>
      <div
        style={{
          position: "absolute",
          left: 92,
          right: 92,
          top: 155,
          bottom: 92,
          display: "grid",
          gridTemplateColumns: "360px 1fr",
          overflow: "hidden",
          border: "1px solid rgba(197,255,216,.16)",
          borderRadius: 28,
          background: "#09150F",
          boxShadow: "0 55px 130px rgba(0,0,0,.48)",
        }}
      >
        <aside
          style={{
            padding: "38px 30px",
            borderRight: "1px solid rgba(255,255,255,.07)",
            background: "#0B1912",
          }}
        >
          <div
            style={{
              color: "#697D71",
              font: `600 13px ${mono}`,
              letterSpacing: 2.2,
              textTransform: "uppercase",
            }}
          >
            Workspace
          </div>
          <div
            style={{ color: "white", font: `750 25px ${sans}`, marginTop: 10 }}
          >
            Acme infrastructure
          </div>
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              gap: 12,
              marginTop: 40,
            }}
          >
            {[
              "Command center",
              "Servers",
              "Teams",
              "Terminals",
              "Activity",
            ].map((item) => (
              <div
                key={item}
                style={{
                  padding: "15px 17px",
                  borderRadius: 12,
                  background:
                    item === "Terminals"
                      ? "rgba(68,240,138,.11)"
                      : "transparent",
                  color: item === "Terminals" ? green : "#87998E",
                  font: `${item === "Terminals" ? 750 : 550} 19px ${sans}`,
                }}
              >
                {item}
              </div>
            ))}
          </div>
          <div
            style={{
              position: "absolute",
              bottom: 34,
              left: 30,
              color: "#64766B",
              font: `14px ${mono}`,
            }}
          >
            ● zero-knowledge workspace
          </div>
        </aside>
        <main style={{ padding: 30 }}>
          <div
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              height: 55,
            }}
          >
            <div>
              <span style={{ color: "white", font: `750 24px ${sans}` }}>
                production-api
              </span>
              <span
                style={{
                  color: "#6D8175",
                  font: `15px ${mono}`,
                  marginLeft: 17,
                }}
              >
                deploy@23.20.159.49
              </span>
            </div>
            <div style={{ display: "flex", gap: 10 }}>
              <span
                style={{
                  padding: "8px 12px",
                  borderRadius: 9,
                  color: green,
                  background: "rgba(68,240,138,.1)",
                  font: `14px ${mono}`,
                }}
              >
                ● connected
              </span>
              <span
                style={{
                  padding: "8px 12px",
                  borderRadius: 9,
                  color: "#A6B7AC",
                  border: "1px solid rgba(255,255,255,.1)",
                  font: `14px ${mono}`,
                }}
              >
                Detach
              </span>
            </div>
          </div>
          <div
            style={{
              height: 645,
              marginTop: 22,
              padding: "34px 40px",
              borderRadius: 17,
              background: "#020705",
              border: "1px solid rgba(255,255,255,.07)",
              color: "#C7D6CC",
              font: `20px/1.7 ${mono}`,
            }}
          >
            <div style={{ color: "#64786B" }}>
              Ubuntu 24.04.3 LTS · production-api
            </div>
            <div style={{ marginTop: 28 }}>
              <span style={{ color: green }}>deploy@prod</span>:
              <span style={{ color: "#5DE7FF" }}>~</span>$ {command}
              <span style={{ opacity: cursor ? 1 : 0, color: green }}>▌</span>
            </div>
            <div
              style={{
                opacity: clamp(frame, [505, 525], [0, 1]),
                color: "#7F9186",
                marginTop: 13,
              }}
            >
              ↳ restarting api.service…
            </div>
            <div
              style={{
                opacity: clamp(frame, [540, 562], [0, 1]),
                color: green,
                marginTop: 6,
              }}
            >
              ✓ api.service is active (running)
            </div>
            <div
              style={{
                opacity: clamp(frame, [585, 610], [0, 1]),
                marginTop: 28,
              }}
            >
              <span style={{ color: green }}>deploy@prod</span>:
              <span style={{ color: "#5DE7FF" }}>~</span>${" "}
              <span style={{ opacity: cursor ? 1 : 0, color: green }}>▌</span>
            </div>
          </div>
        </main>
      </div>
      <div
        style={{
          position: "absolute",
          left: 134,
          bottom: 29,
          display: "flex",
          gap: 12,
        }}
      >
        {[
          "Team-controlled access",
          "Detach & resume",
          "Credentials decrypt here",
        ].map((label, index) => {
          const show = clamp(
            frame,
            [345 + index * 22, 365 + index * 22],
            [0, 1],
          );
          return (
            <span
              key={label}
              style={{
                opacity: show,
                transform: `translateY(${(1 - show) * 12}px)`,
                padding: "9px 14px",
                borderRadius: 99,
                color: index === 1 ? night : "#A6B7AC",
                background: index === 1 ? green : "rgba(255,255,255,.06)",
                font: `700 14px ${sans}`,
              }}
            >
              {label}
            </span>
          );
        })}
      </div>
    </AbsoluteFill>
  );
};

const ResumeScene: React.FC = () => {
  const frame = useCurrentFrame();
  const visible =
    clamp(frame, [690, 720], [0, 1]) * clamp(frame, [815, 838], [1, 0]);
  return (
    <AbsoluteFill style={{ background: "#F1F2F4", opacity: visible }}>
      <div
        style={{
          position: "absolute",
          left: 370,
          top: 62,
          width: 1180,
          height: 940,
          borderRadius: 42,
          overflow: "hidden",
          background: "white",
          boxShadow: "0 45px 120px rgba(28,31,34,.15)",
        }}
      >
        <ChatHeader />
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            gap: 24,
            padding: "52px 78px",
          }}
        >
          <Bubble at={710} side="right" width={430}>
            Connected. Prod is back. ✅
          </Bubble>
          <Bubble at={752} side="left" width={520}>
            Nice. Leave the session running.
          </Bubble>
          <Bubble at={785} side="left" width={620}>
            <div style={{ fontWeight: 800 }}>Pick it up anywhere.</div>
            <div style={{ marginTop: 8, color: "#6F7477", fontSize: 22 }}>
              Your cloud terminal will be waiting.
            </div>
          </Bubble>
        </div>
      </div>
    </AbsoluteFill>
  );
};

const Finale: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const reveal = spring({
    frame: frame - 822,
    fps,
    config: { damping: 20, stiffness: 120 },
  });
  return (
    <AbsoluteFill
      style={{
        background:
          "radial-gradient(circle at 50% 30%, #113422 0%, #06110C 58%)",
        opacity: clamp(frame, [812, 838], [0, 1]),
        color: "white",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        textAlign: "center",
      }}
    >
      <div
        style={{
          transform: `translateY(${(1 - reveal) * 40}px)`,
          opacity: reveal,
        }}
      >
        <div style={{ display: "flex", justifyContent: "center" }}>
          <Brand light />
        </div>
        <div
          style={{
            marginTop: 58,
            fontFamily: sans,
            fontSize: 76,
            lineHeight: 1.02,
            fontWeight: 750,
            letterSpacing: -4.5,
          }}
        >
          Your servers.
          <br />
          <span style={{ color: green }}>Right where you left them.</span>
        </div>
        <div
          style={{ marginTop: 35, color: "#9EB0A4", font: `24px/1.5 ${sans}` }}
        >
          Browser SSH · team workspaces · persistent cloud sessions
        </div>
        <div
          style={{
            display: "inline-flex",
            marginTop: 48,
            padding: "18px 28px",
            borderRadius: 13,
            background: green,
            color: night,
            font: `800 20px ${sans}`,
            boxShadow: "0 18px 55px rgba(68,240,138,.18)",
          }}
        >
          Open cloud.getssher.com&nbsp;&nbsp;→
        </div>
      </div>
      <div
        style={{
          position: "absolute",
          bottom: 33,
          font: `13px ${mono}`,
          color: "#617369",
          letterSpacing: 2.2,
          textTransform: "uppercase",
        }}
      >
        Built for the terminal. Ready for everywhere.
      </div>
    </AbsoluteFill>
  );
};

export const SsherCloudAnnouncement: React.FC = () => (
  <AbsoluteFill>
    <ChatScene />
    <TerminalScene />
    <ResumeScene />
    <Finale />
  </AbsoluteFill>
);
