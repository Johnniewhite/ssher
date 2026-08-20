import { loadFont as loadManrope } from "@remotion/google-fonts/Manrope";
import { loadFont as loadMono } from "@remotion/google-fonts/IBMPlexMono";

export const { fontFamily: sans } = loadManrope("normal", {
  weights: ["400", "600", "700", "800"],
  subsets: ["latin"],
});

export const { fontFamily: mono } = loadMono("normal", {
  weights: ["400", "500", "600"],
  subsets: ["latin"],
});

export const colors = {
  black: "#020705",
  blackSoft: "#06100B",
  panel: "#0A1710",
  green: "#4CFF8B",
  greenBright: "#88FFB3",
  greenDark: "#14B858",
  mint: "#C8FFDA",
  white: "#F4FFF7",
  gray: "#8DA398",
  dim: "#42584D",
  cyan: "#5DE7FF",
  amber: "#FFBE72",
  red: "#FF6B70",
};

export const shadow = "0 50px 140px rgba(0,0,0,.58)";
