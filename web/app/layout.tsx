import "../styles/globals.css";
import type { ReactNode } from "react";
import { AppProviders } from "./providers";

export const metadata = {
  title: "OffGridFlow",
  description: "Carbon accounting and compliance platform",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body>
        <AppProviders>
          {children}
        </AppProviders>
      </body>
    </html>
  );
}