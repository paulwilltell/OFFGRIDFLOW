import "../styles/globals.css";
import type { ReactNode } from "react";

export const metadata = {
  title: "OffGridFlow",
  description: "Carbon accounting and compliance platform",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
