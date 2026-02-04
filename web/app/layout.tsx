import "../styles/globals.css";
import type { ReactNode } from "react";
import { JetBrains_Mono, Space_Grotesk, Syncopate } from "next/font/google";

const spaceGrotesk = Space_Grotesk({
  subsets: ["latin"],
  display: "swap",
  variable: "--font-body",
});

const jetBrainsMono = JetBrains_Mono({
  subsets: ["latin"],
  display: "swap",
  variable: "--font-mono",
});

const syncopate = Syncopate({
  subsets: ["latin"],
  weight: ["400", "700"],
  display: "swap",
  variable: "--font-display",
});

export const metadata = {
  title: "OffGridFlow",
  description: "Carbon accounting and compliance platform",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html
      lang="en"
      className={`${spaceGrotesk.variable} ${jetBrainsMono.variable} ${syncopate.variable}`}
    >
      <body>{children}</body>
    </html>
  );
}
